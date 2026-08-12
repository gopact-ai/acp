package acp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type invalidCountWriter struct {
	count func(int) int
}

func (w invalidCountWriter) Write(p []byte) (int, error) {
	return w.count(len(p)), nil
}

func TestWriteRejectsInvalidWriterCounts(t *testing.T) {
	tests := []struct {
		name  string
		count func(int) int
	}{
		{name: "negative", count: func(int) int { return -1 }},
		{name: "too large", count: func(n int) int { return n + 1 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, keepOpen := io.Pipe()
			conn, err := New(input, invalidCountWriter{count: tt.count}, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = keepOpen.Close()
				_ = conn.Close()
			})

			err = conn.Notify(t.Context(), "test/notification", nil)
			if err == nil || !strings.Contains(err.Error(), "invalid write count") {
				t.Fatalf("Notify error = %v, want invalid write count", err)
			}
			if conn.Err() == nil || !strings.Contains(conn.Err().Error(), "invalid write count") {
				t.Fatalf("connection error = %v, want invalid write count", conn.Err())
			}
		})
	}
}

func TestOversizedHandlerResponseFallsBackToInternalError(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	server, err := New(serverSide, serverSide, func(_ context.Context, method string, _ json.RawMessage) (any, error) {
		switch method {
		case "test/large":
			return map[string]string{"data": strings.Repeat("x", 512)}, nil
		case "test/small":
			return map[string]bool{"ok": true}, nil
		default:
			return nil, methodNotFound(method)
		}
	}, WithMaxFrameBytes(128))
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

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	var largeResult map[string]string
	err = client.Call(ctx, "test/large", nil, &largeResult)
	var rpcErr *Error
	if !errors.As(err, &rpcErr) || rpcErr.Code != ErrorCodeInternalError {
		t.Fatalf("large Call error = %v, want Internal Error", err)
	}
	if server.Err() != nil {
		t.Fatalf("server closed after fallback response: %v", server.Err())
	}

	var smallResult map[string]bool
	if err := client.Call(t.Context(), "test/small", nil, &smallResult); err != nil {
		t.Fatal(err)
	}
	if !smallResult["ok"] {
		t.Fatalf("small result = %#v", smallResult)
	}
}

func TestOversizedFallbackResponseClosesConnection(t *testing.T) {
	request := requestEnvelope{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`"` + strings.Repeat("i", 40) + `"`),
		Method:  "x",
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	fallbackJSON, err := json.Marshal(responseEnvelope{
		JSONRPC: "2.0",
		ID:      request.ID,
		Error:   internalError(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fallbackJSON)+1 <= len(requestJSON) {
		t.Fatal("test requires a fallback response larger than its request")
	}

	serverSide, peer := net.Pipe()
	server, err := New(serverSide, serverSide, func(context.Context, string, json.RawMessage) (any, error) {
		return map[string]string{"data": strings.Repeat("x", 512)}, nil
	}, WithMaxFrameBytes(len(requestJSON)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = server.Close()
	})

	if err := peer.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := peer.Write(append(requestJSON, '\n')); err != nil {
		t.Fatal(err)
	}
	if _, err := peer.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("peer read error = %v, want EOF", err)
	}
	<-server.Done()
	if !errors.Is(server.Err(), ErrFrameTooLarge) {
		t.Fatalf("server error = %v, want ErrFrameTooLarge", server.Err())
	}
}
