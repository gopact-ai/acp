package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func awaitReceipt[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("receipt fixture did not settle")
		var zero T
		return zero
	}
}

func TestReceiptSequenceTracksActualInboundOrder(t *testing.T) {
	for _, responseFirst := range []bool{false, true} {
		t.Run(fmt.Sprint(responseFirst), func(t *testing.T) {
			input, writer := io.Pipe()
			requests, output := io.Pipe()
			sequence := make(chan uint64, 1)
			conn, err := New(input, output, func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
				seq, ok := NotificationSequence(ctx)
				if !ok {
					t.Error("received notification lacks accepted ingress sequence")
				}
				sequence <- seq
				return nil, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = conn.Close(); _ = writer.Close(); _ = requests.Close(); _ = output.Close() })
			peer := make(chan error, 1)
			go func() {
				frame, err := bufio.NewReader(requests).ReadBytes('\n')
				if err != nil {
					peer <- err
					return
				}
				var envelope struct {
					ID json.RawMessage `json:"id"`
				}
				if err := json.Unmarshal(frame, &envelope); err != nil {
					peer <- err
					return
				}
				response := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{\"ok\":true}}\n", envelope.ID)
				notification := "{\"jsonrpc\":\"2.0\",\"method\":\"test/update\",\"params\":{}}\n"
				frames := notification + response
				if responseFirst {
					frames = response + notification
				}
				_, err = io.WriteString(writer, frames)
				peer <- err
			}()
			var result struct {
				OK bool `json:"ok"`
			}
			receipt, err := conn.CallWithReceipt(t.Context(), "test/call", nil, &result)
			if err != nil || !result.OK {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			notify := awaitReceipt(t, sequence)
			if err := awaitReceipt(t, peer); err != nil {
				t.Fatal(err)
			}
			if receipt.Sequence == 0 || notify == 0 {
				t.Fatalf("missing receipt: response=%d notification=%d", receipt.Sequence, notify)
			}
			if responseFirst && receipt.Sequence >= notify || !responseFirst && receipt.Sequence <= notify {
				t.Fatalf("responseFirst=%v response=%d notification=%d", responseFirst, receipt.Sequence, notify)
			}
		})
	}
}

func TestReceiptNotificationContextDoesNotLeakToRequest(t *testing.T) {
	if _, ok := NotificationSequence(context.Background()); ok {
		t.Fatal("plain context had notification sequence")
	}
	a, b := net.Pipe()
	server, err := New(a, a, func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		if _, ok := NotificationSequence(ctx); ok {
			t.Error("request had notification-only metadata")
		}
		return map[string]bool{"ok": true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(b, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close(); _ = client.Close() })
	receipt, err := client.CallWithReceipt(t.Context(), "test/request", nil, nil)
	if err != nil || receipt.Sequence == 0 {
		t.Fatalf("request response receipt=%+v err=%v", receipt, err)
	}
}

func TestReceiptAbsenceBeforeDispatch(t *testing.T) {
	input, writer := io.Pipe()
	conn, err := New(input, io.Discard, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(); _ = writer.Close() })
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, tc := range []struct {
		ctx    context.Context
		method string
	}{{nil, "test"}, {t.Context(), ""}, {ctx, "test"}} {
		receipt, err := conn.CallWithReceipt(tc.ctx, tc.method, nil, nil)
		if err == nil || receipt.Sequence != 0 {
			t.Fatalf("undispatched receipt=%+v err=%v", receipt, err)
		}
	}
}

func TestReceiptNotificationReentrantCallKeepsWireOrdering(t *testing.T) {
	a, b := net.Pipe()
	var server, client *Conn
	var err error
	server, err = New(a, a, func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		if err := server.Notify(ctx, "test/followup", nil); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	completed := make(chan error, 1)
	followup := make(chan uint64, 1)
	client, err = New(b, b, func(ctx context.Context, method string, _ json.RawMessage) (any, error) {
		seq, ok := NotificationSequence(ctx)
		if !ok {
			t.Error("reentrant notification has no ingress sequence")
		}
		if method == "test/followup" {
			followup <- seq
			return nil, nil
		}
		receipt, err := client.CallWithReceipt(ctx, "test/nested", nil, nil)
		if err == nil && receipt.Sequence <= seq {
			err = errors.New("nested response lacks later ingress receipt")
		}
		completed <- err
		return nil, err
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close(); _ = server.Close() })
	if err := server.Notify(t.Context(), "test/start", nil); err != nil {
		t.Fatal(err)
	}
	if err := awaitReceipt(t, completed); err != nil {
		t.Fatal(err)
	}
	if seq := awaitReceipt(t, followup); seq == 0 {
		t.Fatal("followup has no ingress sequence")
	}
}

func TestReceiptSurvivesMatchedPeerErrorAndResultDecodeError(t *testing.T) {
	for _, peerError := range []bool{false, true} {
		t.Run(fmt.Sprint(peerError), func(t *testing.T) {
			a, b := net.Pipe()
			server, err := New(a, a, func(context.Context, string, json.RawMessage) (any, error) {
				if peerError {
					return nil, &Error{Code: ErrorCodeInvalidParams, Message: "fixture refusal"}
				}
				return map[string]bool{"ok": true}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			client, err := New(b, b, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = server.Close(); _ = client.Close() })
			var impossible int
			receipt, err := client.CallWithReceipt(t.Context(), "test/error", nil, &impossible)
			if err == nil || receipt.Sequence == 0 {
				t.Fatalf("matched frame receipt=%+v err=%v", receipt, err)
			}
			var rpc *Error
			if peerError && (!errors.As(err, &rpc) || rpc.Code != ErrorCodeInvalidParams) {
				t.Fatalf("peer error lost: %v", err)
			}
		})
	}
}

func TestReceiptPriorNotificationBarrierCancellationRetainsMatchedFrame(t *testing.T) {
	input, writer := io.Pipe()
	requests, output := io.Pipe()
	entered, release := make(chan struct{}), make(chan struct{})
	conn, err := New(input, output, func(context.Context, string, json.RawMessage) (any, error) {
		close(entered)
		<-release
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
		_ = conn.Close()
		_ = writer.Close()
		_ = requests.Close()
		_ = output.Close()
	})
	written := make(chan error, 1)
	go func() {
		frame, err := bufio.NewReader(requests).ReadBytes('\n')
		if err != nil {
			written <- err
			return
		}
		var env struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(frame, &env); err != nil {
			written <- err
			return
		}
		_, err = fmt.Fprintf(writer, "{\"jsonrpc\":\"2.0\",\"method\":\"test/barrier\"}\n{\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":{}}\n", env.ID)
		written <- err
	}()
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	type outcome struct {
		receipt ResponseReceipt
		err     error
	}
	done := make(chan outcome, 1)
	go func() { r, e := conn.CallWithReceipt(ctx, "test/wait", nil, nil); done <- outcome{r, e} }()
	awaitReceipt(t, entered)
	if err := awaitReceipt(t, written); err != nil {
		t.Fatal(err)
	}
	// io.Pipe write has reached the response frame; wait for the real pending
	// match before canceling, without manufacturing an ingress ordinal.
	until := time.After(time.Second)
	for {
		conn.mu.Lock()
		matched := len(conn.pending) == 0
		conn.mu.Unlock()
		if matched {
			break
		}
		select {
		case <-until:
			t.Fatal("response did not match")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cause := errors.New("local barrier canceled")
	cancel(cause)
	result := awaitReceipt(t, done)
	if !errors.Is(result.err, cause) || result.receipt.Sequence == 0 {
		t.Fatalf("barrier error lost receipt: %+v", result)
	}
	close(release)
}

func TestReceiptTypedConfigOptionRetainsOriginalValidation(t *testing.T) {
	for _, result := range []string{`{"configOptions":[]}`, `{}`} {
		t.Run(result, func(t *testing.T) {
			input, writer := io.Pipe()
			requests, output := io.Pipe()
			var caller *AgentCaller
			conn, err := NewClient(input, output, func(c *AgentCaller) ClientHandler { caller = c; return &testClient{} })
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = conn.Close(); _ = writer.Close(); _ = requests.Close(); _ = output.Close() })
			peer := make(chan error, 1)
			go func() {
				frame, err := bufio.NewReader(requests).ReadBytes('\n')
				if err != nil {
					peer <- err
					return
				}
				var env struct {
					ID     json.RawMessage `json:"id"`
					Method string          `json:"method"`
					Params json.RawMessage `json:"params"`
				}
				if err := json.Unmarshal(frame, &env); err != nil {
					peer <- err
					return
				}
				var req SetSessionConfigOptionRequest
				if err := json.Unmarshal(env.Params, &req); err != nil {
					peer <- err
					return
				}
				if env.Method != MethodSessionSetConfigOption || req.Value != false {
					peer <- fmt.Errorf("typed wire changed: %s", frame)
					return
				}
				_, err = fmt.Fprintf(writer, "{\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":%s}\n", env.ID, result)
				peer <- err
			}()
			req := BooleanSetSessionConfigOptionRequest("session", "toggle", false)
			resp, receipt, err := caller.SetSessionConfigOptionWithReceipt(t.Context(), &req)
			if e := awaitReceipt(t, peer); e != nil {
				t.Fatal(e)
			}
			if receipt.Sequence == 0 {
				t.Fatal("typed matching response lacks receipt")
			}
			if result == `{}` {
				if !errors.Is(err, ErrInvalidResponse) || resp != nil {
					t.Fatalf("typed required field validation relaxed: resp=%+v err=%v", resp, err)
				}
			} else if err != nil || resp == nil || resp.ConfigOptions == nil || len(resp.ConfigOptions) != 0 {
				t.Fatalf("confirmed [] lost: resp=%+v err=%v", resp, err)
			}
		})
	}
}

func TestReceiptTypedInvalidInputIsUndispatched(t *testing.T) {
	input, writer := io.Pipe()
	var caller *AgentCaller
	conn, err := NewClient(input, io.Discard, func(c *AgentCaller) ClientHandler { caller = c; return &testClient{} })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(); _ = writer.Close() })
	resp, receipt, err := caller.SetSessionConfigOptionWithReceipt(t.Context(), nil)
	if err == nil || resp != nil || receipt.Sequence != 0 {
		t.Fatalf("invalid request reached wire: resp=%+v receipt=%+v err=%v", resp, receipt, err)
	}
}

func TestReceiptResponseThenEOFStillReturnsQueuedMatch(t *testing.T) {
	input, peer := io.Pipe()
	output := &responseThenEOFWriter{peer: peer}
	conn, err := New(input, output, nil)
	if err != nil {
		t.Fatal(err)
	}
	output.conn = conn
	t.Cleanup(func() { _ = conn.Close() })
	var result struct {
		OK bool `json:"ok"`
	}
	receipt, err := conn.CallWithReceipt(t.Context(), "test/request", nil, &result)
	if err != nil || !result.OK || receipt.Sequence == 0 {
		t.Fatalf("queued response lost to EOF: %+v %+v %v", result, receipt, err)
	}
}

func TestReceiptSessionMethodsRetainTypedWireContracts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		result string
		invoke func(context.Context, *AgentCaller) (ResponseReceipt, error)
	}{
		{"new", MethodSessionNew, `{"sessionId":"session","configOptions":[]}`, func(ctx context.Context, c *AgentCaller) (ResponseReceipt, error) {
			_, r, e := c.NewSessionWithReceipt(ctx, &NewSessionRequest{Cwd: "/fixture"})
			return r, e
		}},
		{"load", MethodSessionLoad, `{"configOptions":[]}`, func(ctx context.Context, c *AgentCaller) (ResponseReceipt, error) {
			_, r, e := c.LoadSessionWithReceipt(ctx, &LoadSessionRequest{SessionID: "session", Cwd: "/fixture"})
			return r, e
		}},
		{"resume", MethodSessionResume, `{"configOptions":[]}`, func(ctx context.Context, c *AgentCaller) (ResponseReceipt, error) {
			_, r, e := c.ResumeSessionWithReceipt(ctx, &ResumeSessionRequest{SessionID: "session", Cwd: "/fixture"})
			return r, e
		}},
		{"mode", MethodSessionSetMode, `{}`, func(ctx context.Context, c *AgentCaller) (ResponseReceipt, error) {
			_, r, e := c.SetSessionModeWithReceipt(ctx, &SetSessionModeRequest{SessionID: "session", ModeID: "read"})
			return r, e
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input, writer := io.Pipe()
			requests, output := io.Pipe()
			var caller *AgentCaller
			conn, err := NewClient(input, output, func(c *AgentCaller) ClientHandler { caller = c; return &testClient{} })
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = conn.Close(); _ = writer.Close(); _ = requests.Close(); _ = output.Close() })
			peer := make(chan error, 1)
			go func() {
				frame, err := bufio.NewReader(requests).ReadBytes('\n')
				if err != nil {
					peer <- err
					return
				}
				var env struct {
					ID     json.RawMessage `json:"id"`
					Method string          `json:"method"`
				}
				if err := json.Unmarshal(frame, &env); err != nil {
					peer <- err
					return
				}
				if env.Method != tc.method {
					peer <- fmt.Errorf("got method=%s want=%s", env.Method, tc.method)
					return
				}
				_, err = fmt.Fprintf(writer, "{\"jsonrpc\":\"2.0\",\"id\":%s,\"result\":%s}\n", env.ID, tc.result)
				peer <- err
			}()
			receipt, err := tc.invoke(t.Context(), caller)
			if e := awaitReceipt(t, peer); e != nil {
				t.Fatal(e)
			}
			if err != nil || receipt.Sequence == 0 {
				t.Fatalf("typed method receipt=%+v err=%v", receipt, err)
			}
		})
	}
}
