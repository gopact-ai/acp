package acp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
)

func TestTypedCallPreservesResponseDecodeError(t *testing.T) {
	connSide, peer := net.Pipe()
	conn, err := New(connSide, connSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = conn.Close()
	})

	peerDone := make(chan error, 1)
	go func() {
		line, readErr := bufio.NewReader(peer).ReadBytes('\n')
		if readErr != nil {
			peerDone <- readErr
			return
		}
		var req requestEnvelope
		if unmarshalErr := json.Unmarshal(line, &req); unmarshalErr != nil {
			peerDone <- unmarshalErr
			return
		}
		response := fmt.Sprintf(`{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"invalid"}}`+"\n", req.ID)
		_, writeErr := io.WriteString(peer, response)
		peerDone <- writeErr
	}()

	caller := &AgentCaller{conn: conn}
	_, err = caller.Initialize(t.Context(), &InitializeRequest{ProtocolVersion: ProtocolVersionV1})
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("Initialize error = %v, want ErrInvalidResponse", err)
	}
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("Initialize error = %v, want json.UnmarshalTypeError in chain", err)
	}
	if peerErr := <-peerDone; peerErr != nil {
		t.Fatal(peerErr)
	}
}
