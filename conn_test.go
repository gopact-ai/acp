package acp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type testAgent struct {
	client        *ClientCaller
	updateStarted chan struct{}
	releaseUpdate chan struct{}
}

func (a *testAgent) Initialize(context.Context, *InitializeRequest) (*InitializeResponse, error) {
	return &InitializeResponse{ProtocolVersion: ProtocolVersionV1}, nil
}

func (a *testAgent) NewSession(context.Context, *NewSessionRequest) (*NewSessionResponse, error) {
	return &NewSessionResponse{SessionID: "session-1"}, nil
}

func (a *testAgent) Prompt(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	if err := a.client.Update(ctx, &SessionNotification{
		SessionID: req.SessionID,
		Update:    AgentMessageChunkSessionUpdate(TextContentBlock("hello")),
	}); err != nil {
		return nil, err
	}
	return &PromptResponse{StopReason: StopReasonEndTurn}, nil
}

func (a *testAgent) Cancel(context.Context, *CancelNotification) error {
	return nil
}

func (a *testAgent) DeleteSession(context.Context, *DeleteSessionRequest) (*DeleteSessionResponse, error) {
	return &DeleteSessionResponse{}, nil
}

type testClient struct {
	updateStarted chan struct{}
	releaseUpdate chan struct{}
	once          sync.Once
}

func (c *testClient) RequestPermission(context.Context, *RequestPermissionRequest) (*RequestPermissionResponse, error) {
	return &RequestPermissionResponse{Outcome: SelectedRequestPermissionOutcome("allow")}, nil
}

func (c *testClient) Update(ctx context.Context, _ *SessionNotification) error {
	c.once.Do(func() { close(c.updateStarted) })
	select {
	case <-c.releaseUpdate:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

type readOnlyClient struct{}

func (*readOnlyClient) RequestPermission(context.Context, *RequestPermissionRequest) (*RequestPermissionResponse, error) {
	return &RequestPermissionResponse{}, nil
}

func (*readOnlyClient) Update(context.Context, *SessionNotification) error {
	return nil
}

func (*readOnlyClient) ReadTextFile(context.Context, *ReadTextFileRequest) (*ReadTextFileResponse, error) {
	return &ReadTextFileResponse{Content: "contents"}, nil
}

type blockingWriter struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	w.once.Do(func() { close(w.started) })
	<-w.release
	return len(p), nil
}

type responseThenEOFWriter struct {
	peer *io.PipeWriter
	conn *Conn
}

func (w *responseThenEOFWriter) Write(p []byte) (int, error) {
	frames := `{"jsonrpc":"2.0","method":"test/notification"}` + "\n" + `{"jsonrpc":"2.0","id":1,"result":{"ok":true}}` + "\n"
	if _, err := io.WriteString(w.peer, frames); err != nil {
		return 0, err
	}
	if err := w.peer.Close(); err != nil {
		return 0, err
	}
	<-w.conn.Done()
	return len(p), nil
}

func TestClientAgentRoundTripPreservesNotificationOrder(t *testing.T) {
	agentSide, clientSide := net.Pipe()
	updateStarted := make(chan struct{})
	releaseUpdate := make(chan struct{})

	var agentCaller *AgentCaller
	agentConn, err := NewAgent(agentSide, agentSide, func(client *ClientCaller) AgentHandler {
		return &testAgent{client: client, updateStarted: updateStarted, releaseUpdate: releaseUpdate}
	})
	if err != nil {
		t.Fatal(err)
	}
	clientConn, err := NewClient(clientSide, clientSide, func(agent *AgentCaller) ClientHandler {
		agentCaller = agent
		return &testClient{updateStarted: updateStarted, releaseUpdate: releaseUpdate}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = agentConn.Close()
	})

	initResp, err := agentCaller.Initialize(t.Context(), &InitializeRequest{ProtocolVersion: ProtocolVersionV1})
	if err != nil {
		t.Fatal(err)
	}
	if initResp.ProtocolVersion != ProtocolVersionV1 {
		t.Fatalf("protocol version = %d, want %d", initResp.ProtocolVersion, ProtocolVersionV1)
	}

	session, err := agentCaller.NewSession(t.Context(), &NewSessionRequest{Cwd: "/tmp", MCPServers: []MCPServer{}})
	if err != nil {
		t.Fatal(err)
	}

	promptDone := make(chan error, 1)
	go func() {
		resp, promptErr := agentCaller.Prompt(t.Context(), &PromptRequest{
			SessionID: session.SessionID,
			Prompt:    []ContentBlock{TextContentBlock("hi")},
		})
		if promptErr == nil && resp.StopReason != StopReasonEndTurn {
			promptErr = errors.New("unexpected stop reason")
		}
		promptDone <- promptErr
	}()

	select {
	case <-updateStarted:
	case <-time.After(time.Second):
		t.Fatal("session update was not delivered")
	}
	select {
	case err := <-promptDone:
		t.Fatalf("prompt returned before preceding update completed: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(releaseUpdate)
	if err := <-promptDone; err != nil {
		t.Fatal(err)
	}

	if _, err := agentCaller.DeleteSession(t.Context(), &DeleteSessionRequest{SessionID: session.SessionID}); err != nil {
		t.Fatal(err)
	}
}

func TestNotificationHandlerCanCallPeer(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	var server *Conn
	server, err := New(serverSide, serverSide, func(ctx context.Context, method string, _ json.RawMessage) (any, error) {
		if method != "test/nested" {
			return nil, methodNotFound(method)
		}
		if notifyErr := server.Notify(ctx, "test/followup", nil); notifyErr != nil {
			return nil, notifyErr
		}
		return map[string]bool{"ok": true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var client *Conn
	nestedDone := make(chan error, 1)
	followupDone := make(chan struct{})
	client, err = New(clientSide, clientSide, func(ctx context.Context, method string, _ json.RawMessage) (any, error) {
		switch method {
		case "test/followup":
			close(followupDone)
			return nil, nil
		case "test/notify":
		default:
			return nil, methodNotFound(method)
		}
		var result map[string]bool
		if callErr := client.Call(ctx, "test/nested", nil, &result); callErr != nil {
			nestedDone <- callErr
			return nil, callErr
		}
		if !result["ok"] {
			callErr := errors.New("nested call returned unexpected result")
			nestedDone <- callErr
			return nil, callErr
		}
		nestedDone <- nil
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	if err := server.Notify(t.Context(), "test/notify", nil); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-nestedDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("nested call from notification handler deadlocked")
	}
	select {
	case <-followupDone:
	case <-time.After(time.Second):
		t.Fatal("notification preceding nested response was not delivered")
	}
}

func TestCallCancellationCancelsInboundRequest(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	started := make(chan struct{})
	cancelled := make(chan struct{})

	server, err := New(serverSide, serverSide, func(ctx context.Context, method string, _ json.RawMessage) (any, error) {
		if method != "test/block" {
			return nil, methodNotFound(method)
		}
		close(started)
		<-ctx.Done()
		close(cancelled)
		return nil, context.Cause(ctx)
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(clientSide, clientSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	ctx, cancel := context.WithCancel(t.Context())
	callDone := make(chan error, 1)
	go func() { callDone <- client.Call(ctx, "test/block", struct{}{}, nil) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	if err := <-callDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("call error = %v, want context canceled", err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("inbound request was not cancelled")
	}
}

func TestEquivalentJSONRequestIDsMatch(t *testing.T) {
	t.Run("numeric response", func(t *testing.T) {
		connSide, peer := net.Pipe()
		conn, err := New(connSide, connSide, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = peer.Close()
			_ = conn.Close()
		})
		go func() {
			reader := bufio.NewReader(peer)
			if _, err := reader.ReadBytes('\n'); err != nil {
				return
			}
			if _, err := io.WriteString(peer, `{"jsonrpc":"2.0","id":1.0,"result":{"ok":true}}`+"\n"); err != nil {
				return
			}
			_, _ = reader.ReadBytes('\n')
		}()
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		var result map[string]bool
		if err := conn.Call(ctx, "test", nil, &result); err != nil {
			t.Fatal(err)
		}
		if !result["ok"] {
			t.Fatalf("result = %#v", result)
		}
	})

	t.Run("escaped cancellation ID", func(t *testing.T) {
		serverSide, peer := net.Pipe()
		started := make(chan struct{})
		cancelled := make(chan struct{})
		server, err := New(serverSide, serverSide, func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
			close(started)
			<-ctx.Done()
			close(cancelled)
			return nil, context.Cause(ctx)
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = peer.Close()
			_ = server.Close()
		})
		if _, err := io.WriteString(peer, `{"jsonrpc":"2.0","id":"a","method":"test"}`+"\n"); err != nil {
			t.Fatal(err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("request did not start")
		}
		cancelNotification := `{"jsonrpc":"2.0","method":"$/cancel_request","params":{"requestId":"\u0061"}}` + "\n"
		if _, err := io.WriteString(peer, cancelNotification); err != nil {
			t.Fatal(err)
		}
		select {
		case <-cancelled:
		case <-time.After(time.Second):
			t.Fatal("equivalent cancellation ID did not match request")
		}
	})
}

func TestCallCancellationWhileAnotherWriteIsBlocked(t *testing.T) {
	input, keepOpen := io.Pipe()
	output := &blockingWriter{started: make(chan struct{}), release: make(chan struct{})}
	conn, err := New(input, output, nil)
	if err != nil {
		t.Fatal(err)
	}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(output.release) }) }
	t.Cleanup(func() {
		release()
		_ = keepOpen.Close()
		_ = conn.Close()
	})

	notifyDone := make(chan error, 1)
	go func() { notifyDone <- conn.Notify(context.Background(), "test/blocked", nil) }()
	select {
	case <-output.started:
	case <-time.After(time.Second):
		t.Fatal("first write did not start")
	}

	ctx, cancel := context.WithCancel(t.Context())
	callDone := make(chan error, 1)
	go func() { callDone <- conn.Call(ctx, "test/cancelled", nil, nil) }()
	cancel()
	select {
	case err := <-callDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("call error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("call remained blocked after cancellation")
	}

	release()
	if err := <-notifyDone; err != nil {
		t.Fatal(err)
	}
}

func TestCancelledRequestMayReturnResult(t *testing.T) {
	serverSide, peer := net.Pipe()
	started := make(chan struct{})
	server, err := New(serverSide, serverSide, func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		close(started)
		<-ctx.Done()
		return map[string]bool{"partial": true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = server.Close()
	})

	if _, writeErr := io.WriteString(peer, `{"jsonrpc":"2.0","id":1,"method":"test"}`+"\n"); writeErr != nil {
		t.Fatal(writeErr)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancelNotification := `{"jsonrpc":"2.0","method":"$/cancel_request","params":{"requestId":1}}` + "\n"
	if _, writeErr := io.WriteString(peer, cancelNotification); writeErr != nil {
		t.Fatal(writeErr)
	}

	line, err := bufio.NewReader(peer).ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil || string(resp.Result) != `{"partial":true}` {
		t.Fatalf("response = %s, want partial result", line)
	}
}

func TestCallReturnsQueuedResponseBeforeTerminalEOF(t *testing.T) {
	for range 100 {
		func() {
			input, peer := io.Pipe()
			output := &responseThenEOFWriter{peer: peer}
			conn, err := New(input, output, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.Close() }()
			output.conn = conn

			var result struct {
				OK bool `json:"ok"`
			}
			if err := conn.Call(t.Context(), "test/request", nil, &result); err != nil {
				t.Fatalf("call error = %v, want queued response", err)
			}
			if !result.OK {
				t.Fatalf("result = %#v, want ok", result)
			}
		}()
	}
}

func TestWireErrors(t *testing.T) {
	serverSide, peer := net.Pipe()
	server, err := New(serverSide, serverSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = server.Close()
	})
	reader := bufio.NewReader(peer)

	tests := []struct {
		name         string
		line         string
		expectedCode ErrorCode
		expectedID   string
	}{
		{name: "parse error", line: "{]", expectedCode: ErrorCodeParseError, expectedID: "null"},
		{
			name:         "invalid UTF-8",
			line:         "{\"jsonrpc\":\"2.0\",\"method\":\"\xff\"}",
			expectedCode: ErrorCodeParseError,
			expectedID:   "null",
		},
		{name: "valid non-object", line: `[]`, expectedCode: ErrorCodeInvalidRequest, expectedID: "null"},
		{
			name:         "invalid version",
			line:         `{"jsonrpc":"1.0","id":7,"method":"x"}`,
			expectedCode: ErrorCodeInvalidRequest,
			expectedID:   "7",
		},
		{
			name:         "invalid request ID",
			line:         `{"jsonrpc":"2.0","id":true,"method":"x"}`,
			expectedCode: ErrorCodeInvalidRequest,
			expectedID:   "null",
		},
		{
			name:         "method not found",
			line:         `{"jsonrpc":"2.0","id":8,"method":"x","params":{}}`,
			expectedCode: ErrorCodeMethodNotFound,
			expectedID:   "8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := io.WriteString(peer, tt.line+"\n"); err != nil {
				t.Fatal(err)
			}
			line, err := reader.ReadBytes('\n')
			if err != nil {
				t.Fatal(err)
			}
			var resp responseEnvelope
			if err := json.Unmarshal(line, &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error == nil || resp.Error.Code != tt.expectedCode {
				t.Fatalf("error = %#v, want code %d", resp.Error, tt.expectedCode)
			}
			if string(resp.ID) != tt.expectedID {
				t.Fatalf("id = %s, want %s", resp.ID, tt.expectedID)
			}
		})
	}
}

func TestMalformedErrorResponseClosesConnection(t *testing.T) {
	input := io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","id":1,"error":{"code":null,"message":null}}` + "\n"))
	conn, err := New(input, io.Discard, nil)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not stop")
	}
	if !errors.Is(conn.Err(), ErrInvalidResponse) {
		t.Fatalf("connection error = %v, want ErrInvalidResponse", conn.Err())
	}
}

func TestUnencodableRPCErrorFallsBackToInternalError(t *testing.T) {
	serverSide, peer := net.Pipe()
	server, err := New(serverSide, serverSide, func(context.Context, string, json.RawMessage) (any, error) {
		return nil, &Error{
			Code:    ErrorCodeInternalError,
			Message: "broken error data",
			Data:    func() {},
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = server.Close()
	})

	if _, writeErr := io.WriteString(peer, `{"jsonrpc":"2.0","id":1,"method":"test"}`+"\n"); writeErr != nil {
		t.Fatal(writeErr)
	}
	if deadlineErr := peer.SetReadDeadline(time.Now().Add(time.Second)); deadlineErr != nil {
		t.Fatal(deadlineErr)
	}
	line, err := bufio.NewReader(peer).ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil || resp.Error.Code != ErrorCodeInternalError || resp.Error.Message != "Internal error" {
		t.Fatalf("response error = %#v, want Internal error", resp.Error)
	}
}

func TestEmptyTypedResponseIsObject(t *testing.T) {
	serverSide, peer := net.Pipe()
	server, err := New(serverSide, serverSide, func(context.Context, string, json.RawMessage) (any, error) {
		var response *AuthenticateResponse
		return response, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = server.Close()
	})

	authenticateRequest := `{"jsonrpc":"2.0","id":1,"method":"authenticate","params":{"methodId":"token"}}` + "\n"
	if _, writeErr := io.WriteString(peer, authenticateRequest); writeErr != nil {
		t.Fatal(writeErr)
	}
	line, err := bufio.NewReader(peer).ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil || string(resp.Result) != `{}` {
		t.Fatalf("response = %s, want empty result object", line)
	}
}

func TestTypedCallerRejectsNullEmptyResponse(t *testing.T) {
	connSide, peer := net.Pipe()
	conn, err := New(connSide, connSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = conn.Close()
	})

	go func() {
		reader := bufio.NewReader(peer)
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			return
		}
		var req requestEnvelope
		if json.Unmarshal(line, &req) != nil {
			return
		}
		_, _ = io.WriteString(peer, `{"jsonrpc":"2.0","id":`+string(req.ID)+`,"result":null}`+"\n")
	}()

	caller := &AgentCaller{conn: conn}
	_, err = caller.Authenticate(t.Context(), &AuthenticateRequest{MethodID: "token"})
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("authenticate error = %v, want ErrInvalidResponse", err)
	}
	if !errors.Is(conn.Err(), ErrInvalidResponse) {
		t.Fatalf("connection error = %v, want ErrInvalidResponse", conn.Err())
	}
}

func TestMalformedResponseClosesConnection(t *testing.T) {
	line := `{"jsonrpc":"2.0","id":1,"result":{},"error":{"code":-32603,"message":"bad"}}` + "\n"
	input := io.NopCloser(strings.NewReader(line))
	conn, err := New(input, io.Discard, nil)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not stop")
	}
	if !errors.Is(conn.Err(), ErrInvalidResponse) {
		t.Fatalf("connection error = %v, want ErrInvalidResponse", conn.Err())
	}
}

func TestResponseMissingResultClosesConnection(t *testing.T) {
	connSide, peer := net.Pipe()
	conn, err := New(connSide, connSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = conn.Close()
	})

	callDone := make(chan error, 1)
	go func() { callDone <- conn.Call(context.Background(), "test", nil, nil) }()
	reader := bufio.NewReader(peer)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	var req requestEnvelope
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(peer, `{"jsonrpc":"2.0","id":`+string(req.ID)+"}\n"); err != nil {
		t.Fatal(err)
	}

	_, _ = reader.ReadBytes('\n')
	_ = peer.Close()
	select {
	case <-conn.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not stop")
	}
	if !errors.Is(conn.Err(), ErrInvalidResponse) {
		t.Fatalf("connection error = %v, want ErrInvalidResponse", conn.Err())
	}
	if err := <-callDone; !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("call error = %v, want ErrInvalidResponse", err)
	}
}

func TestFrameLimitClosesConnection(t *testing.T) {
	input := io.NopCloser(strings.NewReader(strings.Repeat("x", 9) + "\n"))
	conn, err := New(input, io.Discard, nil, WithMaxFrameBytes(8))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not stop")
	}
	if !errors.Is(conn.Err(), ErrFrameTooLarge) {
		t.Fatalf("connection error = %v, want ErrFrameTooLarge", conn.Err())
	}
}

func TestNotificationBacklogClosesConnection(t *testing.T) {
	var once sync.Once
	started := make(chan struct{})
	release := make(chan struct{})
	input := io.NopCloser(strings.NewReader(strings.Repeat(`{"jsonrpc":"2.0","method":"test"}`+"\n", 3)))
	conn, err := New(input, io.Discard, func(context.Context, string, json.RawMessage) (any, error) {
		once.Do(func() { close(started) })
		<-release
		return nil, nil
	}, WithNotificationBacklog(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("notification did not start")
	}
	select {
	case <-conn.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not stop accepting notifications")
	}
	if !errors.Is(conn.Err(), ErrQueueFull) {
		t.Fatalf("connection error = %v, want ErrQueueFull", conn.Err())
	}
	close(release)
	<-conn.Done()
}

func TestBufferedPayloadLimitClosesConnection(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	line := `{"jsonrpc":"2.0","method":"test","params":{"data":"payload"}}`
	input := io.NopCloser(strings.NewReader(line + "\n" + line + "\n"))
	conn, err := New(input, io.Discard, func(context.Context, string, json.RawMessage) (any, error) {
		close(started)
		<-release
		return nil, nil
	}, WithMaxFrameBytes(len(line)), WithMaxBufferedBytes(len(line)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("notification did not start")
	}
	select {
	case <-conn.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not enforce buffered payload limit")
	}
	if !errors.Is(conn.Err(), ErrBufferFull) {
		t.Fatalf("connection error = %v, want ErrBufferFull", conn.Err())
	}
	close(release)
	<-conn.Done()
}

func TestConcurrentRequestLimitClosesConnection(t *testing.T) {
	var once sync.Once
	started := make(chan struct{})
	release := make(chan struct{})
	input := io.NopCloser(strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"test"}` + "\n" +
			`{"jsonrpc":"2.0","id":2,"method":"test"}` + "\n",
	))
	conn, err := New(input, io.Discard, func(context.Context, string, json.RawMessage) (any, error) {
		once.Do(func() { close(started) })
		<-release
		return nil, nil
	}, WithMaxConcurrentRequests(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	select {
	case <-conn.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not enforce request limit")
	}
	if !errors.Is(conn.Err(), ErrTooManyRequests) {
		t.Fatalf("connection error = %v, want ErrTooManyRequests", conn.Err())
	}
	close(release)
	<-conn.Done()
}

func TestDuplicateRequestIDDoesNotProduceDuplicateResponse(t *testing.T) {
	serverSide, peer := net.Pipe()
	started := make(chan struct{})
	release := make(chan struct{})
	server, err := New(serverSide, serverSide, func(context.Context, string, json.RawMessage) (any, error) {
		close(started)
		<-release
		return map[string]bool{"ok": true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = server.Close()
	})

	request := `{"jsonrpc":"2.0","id":1,"method":"test"}` + "\n"
	if _, writeErr := io.WriteString(peer, request); writeErr != nil {
		t.Fatal(writeErr)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	if _, writeErr := io.WriteString(peer, request); writeErr != nil {
		t.Fatal(writeErr)
	}
	close(release)

	line, err := bufio.NewReader(peer).ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil || string(resp.Result) != `{"ok":true}` {
		t.Fatalf("response = %s, want successful original response", line)
	}
}

func TestIndependentFileSystemCapabilities(t *testing.T) {
	handler := dispatchClient(&readOnlyClient{})
	result, err := handler(t.Context(), MethodFSReadTextFile, json.RawMessage(`{"sessionId":"s","path":"/tmp/file"}`))
	if err != nil {
		t.Fatal(err)
	}
	if response, ok := result.(*ReadTextFileResponse); !ok || response.Content != "contents" {
		t.Fatalf("read response = %#v", result)
	}
	writeParams := json.RawMessage(`{"sessionId":"s","path":"/tmp/file","content":"x"}`)
	if _, err := handler(t.Context(), MethodFSWriteTextFile, writeParams); err == nil {
		t.Fatal("write succeeded, want Method Not Found")
	} else {
		var rpcErr *Error
		if !errors.As(err, &rpcErr) || rpcErr.Code != ErrorCodeMethodNotFound {
			t.Fatalf("write error = %v, want Method Not Found", err)
		}
	}
}

func TestTypedDispatchRejectsMissingRequiredParams(t *testing.T) {
	handler := dispatchAgent(&testAgent{})
	for _, params := range []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"sessionId":null,"prompt":null}`),
	} {
		_, err := handler(t.Context(), MethodSessionPrompt, params)
		var rpcErr *Error
		if !errors.As(err, &rpcErr) || rpcErr.Code != ErrorCodeInvalidParams {
			t.Fatalf("dispatch(%s) error = %v, want Invalid Params", params, err)
		}
	}
}

func TestTypedCallRejectsMissingRequiredResult(t *testing.T) {
	for _, result := range []string{`{}`, `{"stopReason":null}`, `{"stopReason":"unknown"}`} {
		t.Run(result, func(t *testing.T) {
			connSide, peer := net.Pipe()
			conn, err := New(connSide, connSide, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = peer.Close()
				_ = conn.Close()
			})
			go func() {
				reader := bufio.NewReader(peer)
				line, readErr := reader.ReadBytes('\n')
				if readErr != nil {
					return
				}
				var req requestEnvelope
				if json.Unmarshal(line, &req) != nil {
					return
				}
				_, _ = io.WriteString(peer, `{"jsonrpc":"2.0","id":`+string(req.ID)+`,"result":`+result+"}\n")
			}()
			caller := &AgentCaller{conn: conn}
			_, err = caller.Prompt(t.Context(), &PromptRequest{SessionID: "s", Prompt: []ContentBlock{TextContentBlock("hi")}})
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("prompt error = %v, want ErrInvalidResponse", err)
			}
			if !errors.Is(conn.Err(), ErrInvalidResponse) {
				t.Fatalf("connection error = %v, want ErrInvalidResponse", conn.Err())
			}
		})
	}
}

func TestTypedCallerRejectsInvalidOutboundParams(t *testing.T) {
	input, keepOpen := io.Pipe()
	var output bytes.Buffer
	conn, err := New(input, &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = keepOpen.Close()
		_ = conn.Close()
	})
	caller := &AgentCaller{conn: conn}
	if _, err := caller.Prompt(t.Context(), &PromptRequest{
		SessionID: "s",
		Prompt:    []ContentBlock{{Type: "video"}},
	}); err == nil {
		t.Fatal("prompt succeeded with unknown content block")
	}
	if err := caller.Cancel(t.Context(), nil); err == nil {
		t.Fatal("cancel succeeded with null params")
	}
	if err := conn.Call(t.Context(), "test", 1, nil); err == nil {
		t.Fatal("generic call succeeded with scalar params")
	}
	if err := conn.Notify(t.Context(), "test", true); err == nil {
		t.Fatal("generic notification succeeded with scalar params")
	}
	client := &ClientCaller{conn: conn}
	if err := client.Update(t.Context(), &SessionNotification{
		SessionID: "s",
		Update: PlanSessionUpdate([]PlanEntry{{
			Content:  "work",
			Priority: PlanEntryPriority("bogus"),
			Status:   PlanEntryStatusPending,
		}}),
	}); err == nil {
		t.Fatal("session update succeeded with invalid nested enum")
	}
	toolCallUpdate := ToolCallUpdateSessionUpdate("call-1")
	toolCallUpdate.Content = map[string]any{"bad": true}
	if err := client.Update(t.Context(), &SessionNotification{
		SessionID: "s",
		Update:    toolCallUpdate,
	}); err == nil {
		t.Fatal("session update succeeded with invalid tool call content")
	}
	if output.Len() != 0 {
		t.Fatalf("wrote invalid request: %s", output.Bytes())
	}
}

func TestTypedCallerAllowsNullToolCallUpdateContent(t *testing.T) {
	input, keepOpen := io.Pipe()
	var output bytes.Buffer
	conn, err := New(input, &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = keepOpen.Close()
		_ = conn.Close()
	})

	update := ToolCallUpdateSessionUpdate("call-1")
	update.Content = json.RawMessage("null")
	client := &ClientCaller{conn: conn}
	if err := client.Update(t.Context(), &SessionNotification{
		SessionID: "s",
		Update:    update,
	}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"content":null`)) {
		t.Fatalf("notification = %s, want null tool call content", output.Bytes())
	}
}

func TestCloseUnblocksReader(t *testing.T) {
	serverSide, peer := net.Pipe()
	conn, err := New(serverSide, serverSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = peer.Close() })
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-conn.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not stop")
	}
	if !errors.Is(conn.Err(), ErrClosed) {
		t.Fatalf("connection error = %v, want ErrClosed", conn.Err())
	}
}

func TestStableMethodCoverage(t *testing.T) {
	b, err := os.ReadFile("schema/v1/meta.json")
	if err != nil {
		t.Fatal(err)
	}
	var meta struct {
		AgentMethods    map[string]string `json:"agentMethods"`
		ClientMethods   map[string]string `json:"clientMethods"`
		ProtocolMethods map[string]string `json:"protocolMethods"`
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	actual := map[string]struct{}{}
	for _, methods := range []map[string]string{meta.AgentMethods, meta.ClientMethods, meta.ProtocolMethods} {
		for _, method := range methods {
			actual[method] = struct{}{}
		}
	}
	expected := map[string]struct{}{}
	for _, method := range []string{
		MethodInitialize, MethodAuthenticate, MethodLogout,
		MethodSessionNew, MethodSessionLoad, MethodSessionSetMode, MethodSessionSetConfigOption,
		MethodSessionPrompt, MethodSessionCancel, MethodSessionList, MethodSessionDelete,
		MethodSessionResume, MethodSessionClose, MethodSessionRequestPermission, MethodSessionUpdate,
		MethodFSWriteTextFile, MethodFSReadTextFile, MethodTerminalCreate, MethodTerminalOutput,
		MethodTerminalRelease, MethodTerminalWaitForExit, MethodTerminalKill,
		MethodElicitationCreate, MethodElicitationComplete, MethodCancelRequest,
	} {
		expected[method] = struct{}{}
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("method set mismatch\nactual: %#v\nexpected: %#v", actual, expected)
	}
}

type extensionHandler struct {
	notifications chan string
}

func (h *extensionHandler) HandleRequest(_ context.Context, method string, params json.RawMessage) (any, error) {
	return map[string]any{"method": method, "params": params}, nil
}

func (h *extensionHandler) HandleNotification(_ context.Context, method string, _ json.RawMessage) error {
	h.notifications <- method
	return nil
}

type extensionAgent struct {
	*testAgent
	*extensionHandler
}

type extensionClient struct {
	*testClient
	*extensionHandler
}

func TestTypedDispatchSupportsExtensions(t *testing.T) {
	agentSide, clientSide := net.Pipe()
	agentNotifications := make(chan string, 1)
	clientNotifications := make(chan string, 1)
	var clientCaller *ClientCaller
	agentConn, err := NewAgent(agentSide, agentSide, func(caller *ClientCaller) AgentHandler {
		clientCaller = caller
		return &extensionAgent{testAgent: &testAgent{}, extensionHandler: &extensionHandler{notifications: agentNotifications}}
	})
	if err != nil {
		t.Fatal(err)
	}
	var agentCaller *AgentCaller
	clientConn, err := NewClient(clientSide, clientSide, func(caller *AgentCaller) ClientHandler {
		agentCaller = caller
		return &extensionClient{testClient: &testClient{}, extensionHandler: &extensionHandler{notifications: clientNotifications}}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = agentConn.Close()
	})

	for _, test := range []struct {
		name          string
		call          func(context.Context, string, any, any) error
		notify        func(context.Context, string, any) error
		notifications chan string
	}{
		{name: "agent", call: agentCaller.Call, notify: agentCaller.Notify, notifications: agentNotifications},
		{name: "client", call: clientCaller.Call, notify: clientCaller.Notify, notifications: clientNotifications},
	} {
		t.Run(test.name, func(t *testing.T) {
			var response struct {
				Method string `json:"method"`
			}
			if err := test.call(t.Context(), "_extension/request", map[string]int{"value": 1}, &response); err != nil {
				t.Fatal(err)
			}
			if response.Method != "_extension/request" {
				t.Fatalf("extension response method = %q", response.Method)
			}
			if err := test.notify(t.Context(), "_extension/notification", struct{}{}); err != nil {
				t.Fatal(err)
			}
			select {
			case method := <-test.notifications:
				if method != "_extension/notification" {
					t.Fatalf("notification method = %q", method)
				}
			case <-time.After(time.Second):
				t.Fatal("extension notification was not delivered")
			}
		})
	}
}
