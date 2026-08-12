package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"strings"
	"sync/atomic"
	"testing"
)

var (
	benchmarkMessage     message
	errBenchmarkRPC      *Error
	benchmarkIDKey       string
	benchmarkIDValid     bool
	benchmarkBlock       ContentBlock
	benchmarkUpdate      SessionUpdate
	benchmarkElicitation CreateElicitationRequest
	benchmarkJSON        []byte
	benchmarkFrame       []byte
)

type repeatingReader struct {
	data   []byte
	offset int
}

func (r *repeatingReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.data[r.offset]
		r.offset++
		if r.offset == len(r.data) {
			r.offset = 0
		}
	}
	return len(p), nil
}

func BenchmarkReadFrame(b *testing.B) {
	cases := map[string][]byte{
		"small":      []byte(`{"jsonrpc":"2.0","method":"test"}` + "\n"),
		"fragmented": []byte(strings.Repeat("x", 16<<10) + "\n"),
	}
	for name, frame := range cases {
		b.Run(name, func(b *testing.B) {
			reader := bufio.NewReaderSize(&repeatingReader{data: frame}, 4096)
			b.ReportAllocs()
			b.SetBytes(int64(len(frame) - 1))
			for b.Loop() {
				var err error
				benchmarkFrame, err = readFrame(reader, len(frame))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeMessage(b *testing.B) {
	largeText := strings.Repeat("x", 64<<10)
	cases := map[string][]byte{
		"request":      []byte(`{"jsonrpc":"2.0","id":1,"method":"session/prompt","params":{"sessionId":"s","prompt":[{"type":"text","text":"hello"}]}}`),
		"response":     []byte(`{"jsonrpc":"2.0","id":1,"result":{"stopReason":"end_turn"}}`),
		"notification": []byte(`{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"s","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"hello"}}}}`),
		"large_request": []byte(`{"jsonrpc":"2.0","id":"request-1","method":"session/prompt","params":{"sessionId":"s","prompt":[{"type":"text","text":"` +
			largeText + `"}]}}`),
	}

	for name, frame := range cases {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(frame)))
			for b.Loop() {
				benchmarkMessage, errBenchmarkRPC = decodeMessage(frame)
			}
		})
	}
}

func BenchmarkIDKey(b *testing.B) {
	cases := map[string]json.RawMessage{
		"integer":             json.RawMessage(`42`),
		"equivalent_decimal":  json.RawMessage(`42.000`),
		"equivalent_exponent": json.RawMessage(`4.2e1`),
		"string":              json.RawMessage(`"request-42"`),
	}

	for name, id := range cases {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkIDKey, benchmarkIDValid = idKey(id)
			}
		})
	}
}

func BenchmarkUnionUnmarshal(b *testing.B) {
	b.Run("content_block", func(b *testing.B) {
		data := []byte(`{"type":"text","text":"hello","annotations":{"audience":["assistant"],"priority":0.8}}`)
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for b.Loop() {
			var block ContentBlock
			if err := json.Unmarshal(data, &block); err != nil {
				b.Fatal(err)
			}
			benchmarkBlock = block
		}
	})

	b.Run("session_update", func(b *testing.B) {
		data := []byte(`{"sessionUpdate":"tool_call_update","toolCallId":"call-1","status":"in_progress","title":"read file","content":[{"type":"content","content":{"type":"text","text":"working"}}]}`)
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for b.Loop() {
			var update SessionUpdate
			if err := json.Unmarshal(data, &update); err != nil {
				b.Fatal(err)
			}
			benchmarkUpdate = update
		}
	})

	b.Run("elicitation_request", func(b *testing.B) {
		data := []byte(`{"sessionId":"session-1","mode":"form","message":"Choose","requestedSchema":{"type":"object","properties":{"strategy":{"type":"string","enum":["a","b"]}}}}`)
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for b.Loop() {
			var request CreateElicitationRequest
			if err := json.Unmarshal(data, &request); err != nil {
				b.Fatal(err)
			}
			benchmarkElicitation = request
		}
	})
}

func BenchmarkUnionMarshal(b *testing.B) {
	b.Run("content_block", func(b *testing.B) {
		value := TextContentBlock("hello")
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkJSON, err = json.Marshal(value)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("session_update", func(b *testing.B) {
		value := AgentMessageChunkSessionUpdate(TextContentBlock("hello"))
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkJSON, err = json.Marshal(value)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("content_block_required_zero", func(b *testing.B) {
		value := TextContentBlock("")
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkJSON, err = json.Marshal(value)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("session_update_required_zero", func(b *testing.B) {
		value := UsageUpdateSessionUpdate(0, 0)
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkJSON, err = json.Marshal(value)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("elicitation_request", func(b *testing.B) {
		strategy := StringElicitationPropertySchema()
		strategy.Enum = &[]string{"a", "b"}
		value := SessionFormCreateElicitationRequest("Choose", ElicitationSchema{Properties: map[string]ElicitationPropertySchema{"strategy": strategy}}, "session-1")
		b.ReportAllocs()
		for b.Loop() {
			var err error
			benchmarkJSON, err = json.Marshal(value)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRequiredFieldValidation(b *testing.B) {
	cases := map[string][]byte{
		"small": []byte(`{"sessionId":"session-1","prompt":[]}`),
		"wide":  []byte(`{"sessionId":"session-1","prompt":[],"_meta":{"payload":"` + strings.Repeat("x", 8<<10) + `"}}`),
	}
	for name, data := range cases {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for b.Loop() {
				if err := validateRequestFields(data, MethodSessionPrompt); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTypedOutboundValidation(b *testing.B) {
	params := &PromptRequest{SessionID: "session-1", Prompt: []ContentBlock{TextContentBlock("hello")}}
	data, err := json.Marshal(params)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for b.Loop() {
		if err := validateRequestFields(data, MethodSessionPrompt); err != nil {
			b.Fatal(err)
		}
		if err := validateTypedJSON(data, params); err != nil {
			b.Fatal(err)
		}
	}
}

type benchmarkRequest struct {
	Text string `json:"text"`
}

type benchmarkResponse struct {
	OK bool `json:"ok"`
}

type benchmarkAgent struct{}

func (*benchmarkAgent) Initialize(context.Context, *InitializeRequest) (*InitializeResponse, error) {
	return &InitializeResponse{ProtocolVersion: ProtocolVersionV1}, nil
}

func (*benchmarkAgent) NewSession(context.Context, *NewSessionRequest) (*NewSessionResponse, error) {
	return &NewSessionResponse{SessionID: "session-1"}, nil
}

func (*benchmarkAgent) Prompt(context.Context, *PromptRequest) (*PromptResponse, error) {
	return &PromptResponse{StopReason: StopReasonEndTurn}, nil
}

func (*benchmarkAgent) Cancel(context.Context, *CancelNotification) error {
	return nil
}

type benchmarkClient struct{}

func (*benchmarkClient) RequestPermission(context.Context, *RequestPermissionRequest) (*RequestPermissionResponse, error) {
	return &RequestPermissionResponse{Outcome: SelectedRequestPermissionOutcome("allow")}, nil
}

func (*benchmarkClient) Update(context.Context, *SessionNotification) error {
	return nil
}

func benchmarkConnection(b *testing.B) *Conn {
	serverSide, clientSide := net.Pipe()
	server, err := New(serverSide, serverSide, func(context.Context, string, json.RawMessage) (any, error) {
		return benchmarkResponse{OK: true}, nil
	})
	if err != nil {
		b.Fatal(err)
	}
	client, err := New(clientSide, clientSide, nil)
	if err != nil {
		_ = server.Close()
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})
	return client
}

func BenchmarkRoundTrip(b *testing.B) {
	ctx := context.Background()
	request := benchmarkRequest{Text: "hello"}

	b.Run("sequential", func(b *testing.B) {
		client := benchmarkConnection(b)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			var response benchmarkResponse
			if err := client.Call(ctx, "benchmark/echo", request, &response); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("parallel", func(b *testing.B) {
		client := benchmarkConnection(b)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var response benchmarkResponse
				if err := client.Call(ctx, "benchmark/echo", request, &response); err != nil {
					b.Error(err)
					return
				}
			}
		})
	})

	b.Run("full_duplex", func(b *testing.B) {
		leftSide, rightSide := net.Pipe()
		handler := func(context.Context, string, json.RawMessage) (any, error) {
			return benchmarkResponse{OK: true}, nil
		}
		left, err := New(leftSide, leftSide, handler)
		if err != nil {
			b.Fatal(err)
		}
		right, err := New(rightSide, rightSide, handler)
		if err != nil {
			_ = left.Close()
			b.Fatal(err)
		}
		b.Cleanup(func() {
			_ = left.Close()
			_ = right.Close()
		})

		var next atomic.Uint64
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				conn := left
				if next.Add(1)%2 == 0 {
					conn = right
				}
				var response benchmarkResponse
				if err := conn.Call(ctx, "benchmark/echo", request, &response); err != nil {
					b.Error(err)
					return
				}
			}
		})
	})
}

func BenchmarkTypedRoundTrip(b *testing.B) {
	serverSide, clientSide := net.Pipe()
	agent, err := NewAgent(serverSide, serverSide, func(*ClientCaller) AgentHandler { return &benchmarkAgent{} })
	if err != nil {
		b.Fatal(err)
	}
	var caller *AgentCaller
	client, err := NewClient(clientSide, clientSide, func(value *AgentCaller) ClientHandler {
		caller = value
		return &benchmarkClient{}
	})
	if err != nil {
		_ = agent.Close()
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = client.Close()
		_ = agent.Close()
	})

	ctx := context.Background()
	request := &InitializeRequest{ProtocolVersion: ProtocolVersionV1}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := caller.Initialize(ctx, request); err != nil {
			b.Fatal(err)
		}
	}
}
