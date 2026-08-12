package acp

import (
	"io"
	"strings"
	"testing"
)

type nilReadCloser struct{}

func (*nilReadCloser) Read([]byte) (int, error) { return 0, io.EOF }
func (*nilReadCloser) Close() error             { return nil }

type nilWriter struct{}

func (*nilWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestConstructorsRejectTypedNilDependencies(t *testing.T) {
	var input *nilReadCloser
	var output *nilWriter
	var agent *testAgent
	var client *testClient

	tests := []struct {
		name string
		new  func() (*Conn, error)
	}{
		{
			name: "input",
			new: func() (*Conn, error) {
				return New(input, io.Discard, nil)
			},
		},
		{
			name: "output",
			new: func() (*Conn, error) {
				return New(io.NopCloser(strings.NewReader("")), output, nil)
			},
		},
		{
			name: "agent handler",
			new: func() (*Conn, error) {
				return NewAgent(io.NopCloser(strings.NewReader("")), io.Discard, func(*ClientCaller) AgentHandler { return agent })
			},
		},
		{
			name: "client handler",
			new: func() (*Conn, error) {
				return NewClient(io.NopCloser(strings.NewReader("")), io.Discard, func(*AgentCaller) ClientHandler { return client })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := tt.new()
			if conn != nil {
				_ = conn.Close()
			}
			if err == nil {
				t.Fatal("constructor accepted a typed nil dependency")
			}
		})
	}
}

func TestNewRejectsInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{name: "nil option", opts: []Option{nil}},
		{name: "nil logger", opts: []Option{WithLogger(nil)}},
		{name: "nonpositive frame limit", opts: []Option{WithMaxFrameBytes(0)}},
		{name: "nonpositive buffer limit", opts: []Option{WithMaxBufferedBytes(0)}},
		{name: "nonpositive request limit", opts: []Option{WithMaxConcurrentRequests(0)}},
		{name: "nonpositive notification backlog", opts: []Option{WithNotificationBacklog(0)}},
		{
			name: "buffer smaller than frame",
			opts: []Option{WithMaxFrameBytes(2), WithMaxBufferedBytes(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := New(io.NopCloser(strings.NewReader("")), io.Discard, nil, tt.opts...)
			if conn != nil {
				_ = conn.Close()
			}
			if err == nil {
				t.Fatal("New accepted invalid options")
			}
		})
	}
}
