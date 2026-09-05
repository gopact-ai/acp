package acp

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPromptResponseUsageRoundTrip(t *testing.T) {
	thought, read, write := uint64(30), uint64(40), uint64(50)
	tests := []struct {
		name string
		wire string
		want PromptResponse
	}{
		{
			name: "all counters and metadata",
			wire: `{"stopReason":"end_turn","usage":{"totalTokens":300,"inputTokens":100,"outputTokens":110,"thoughtTokens":30,"cachedReadTokens":40,"cachedWriteTokens":50,"_meta":{"provider":"test"}},"_meta":{"model_usage":[{"model":"test-model"}]}}`,
			want: PromptResponse{
				StopReason: StopReasonEndTurn,
				Meta:       Meta{"model_usage": []any{map[string]any{"model": "test-model"}}},
				Usage: &Usage{
					TotalTokens: 300, InputTokens: 100, OutputTokens: 110,
					ThoughtTokens: &thought, CachedReadTokens: &read, CachedWriteTokens: &write,
					Meta: Meta{"provider": "test"},
				},
			},
		},
		{
			name: "required zero counters",
			wire: `{"stopReason":"end_turn","usage":{"totalTokens":0,"inputTokens":0,"outputTokens":0}}`,
			want: PromptResponse{StopReason: StopReasonEndTurn, Usage: &Usage{}},
		},
		{
			name: "uint64 precision",
			wire: `{"stopReason":"end_turn","usage":{"totalTokens":18446744073709551615,"inputTokens":9007199254740993,"outputTokens":0}}`,
			want: PromptResponse{StopReason: StopReasonEndTurn, Usage: &Usage{
				TotalTokens: ^uint64(0), InputTokens: 9007199254740993,
			}},
		},
		{
			name: "legacy response",
			wire: `{"stopReason":"end_turn"}`,
			want: PromptResponse{StopReason: StopReasonEndTurn},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got PromptResponse
			if err := json.Unmarshal([]byte(test.wire), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("decoded response = %#v (usage %#v), want %#v", got, got.Usage, test.want)
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if !sameJSONShape(encoded, []byte(test.wire)) {
				t.Fatalf("encoded response = %s, want %s", encoded, test.wire)
			}
			var roundTrip PromptResponse
			if err := json.Unmarshal(encoded, &roundTrip); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(roundTrip, test.want) {
				t.Fatalf("round trip = %#v, want %#v", roundTrip, test.want)
			}
		})
	}
}

func TestPromptResponseRequiresStopReason(t *testing.T) {
	for _, input := range []string{
		`null`, `{}`, `{"usage":{"totalTokens":1,"inputTokens":1,"outputTokens":0}}`,
		`{"stopReason":null}`, `{"stopReason":""}`, `{"stopReason":1}`,
	} {
		t.Run(input, func(t *testing.T) {
			response := PromptResponse{StopReason: StopReasonEndTurn}
			if err := json.Unmarshal([]byte(input), &response); err == nil {
				t.Fatalf("json.Unmarshal(%s) succeeded, want a required stopReason error", input)
			}
			if response.StopReason != StopReasonEndTurn {
				t.Fatalf("failed decode changed the response: %#v", response)
			}
		})
	}
}

func TestPromptResponseUsageDefaults(t *testing.T) {
	for _, usage := range []string{
		`null`, `7`, `{}`, `{"totalTokens":1,"inputTokens":1}`,
		`{"totalTokens":null,"inputTokens":1,"outputTokens":0}`,
		`{"totalTokens":-1,"inputTokens":1,"outputTokens":0}`,
		`{"totalTokens":1.5,"inputTokens":1,"outputTokens":0}`,
		`{"totalTokens":18446744073709551616,"inputTokens":1,"outputTokens":0}`,
	} {
		t.Run(usage, func(t *testing.T) {
			response := PromptResponse{Usage: &Usage{TotalTokens: 10}}
			if err := json.Unmarshal([]byte(`{"stopReason":"end_turn","usage":`+usage+`}`), &response); err != nil {
				t.Fatal(err)
			}
			if response.StopReason != StopReasonEndTurn || response.Usage != nil {
				t.Fatalf("response = %#v, want completed turn without usage", response)
			}
		})
	}
}

func TestUsageOptionalCounters(t *testing.T) {
	for _, value := range []string{`null`, `"invalid"`, `-1`, `1.5`, `18446744073709551616`, `0`} {
		t.Run(value, func(t *testing.T) {
			wire := `{"totalTokens":3,"inputTokens":1,"outputTokens":2,"thoughtTokens":` + value + `,"cachedReadTokens":` + value + `,"cachedWriteTokens":` + value + `}`
			var usage Usage
			if err := json.Unmarshal([]byte(wire), &usage); err != nil {
				t.Fatal(err)
			}
			if usage.TotalTokens != 3 || usage.InputTokens != 1 || usage.OutputTokens != 2 {
				t.Fatalf("required counters changed: %#v", usage)
			}
			for _, counter := range []*uint64{usage.ThoughtTokens, usage.CachedReadTokens, usage.CachedWriteTokens} {
				if value == "0" {
					if counter == nil || *counter != 0 {
						t.Fatalf("zero counter = %v, want reported zero", counter)
					}
				} else if counter != nil {
					t.Fatalf("optional counter = %v, want nil", counter)
				}
			}
			assertStableJSONRoundTrip[Usage](t, []byte(wire))
		})
	}
}
