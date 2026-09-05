package acp

import (
	"bytes"
	"encoding/json"
	"testing"
)

func FuzzDecodeMessage(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{"jsonrpc":"2.0","id":1,"method":"session/list","params":{}}`),
		[]byte(`{"jsonrpc":"2.0","method":"session/cancel","params":{"sessionId":"s"}}`),
		[]byte(`{"jsonrpc":"2.0","id":"a","result":{}}`),
		[]byte(`{"jsonrpc":"2.0","id":"a","error":{"code":-32603,"message":"error"}}`),
		[]byte(`{"jsonrpc":"2.0"}`),
		[]byte(`{]`),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, frame []byte) {
		if len(frame) > 1<<20 {
			t.Skip()
		}
		msg, rpcErr := decodeMessage(frame)
		if rpcErr != nil {
			return
		}
		switch msg.kind {
		case messageRequest:
			if !validID(msg.id) {
				t.Fatal("decoded request has an invalid id")
			}
		case messageNotification:
			if len(msg.id) != 0 {
				t.Fatalf("decoded notification has id %s", msg.id)
			}
		case messageResponse:
			if !validID(msg.id) {
				t.Fatal("decoded response has an invalid id")
			}
			if (len(msg.result) == 0) == (msg.err == nil) {
				t.Fatal("decoded response must contain exactly one of result or error")
			}
		default:
			t.Fatalf("decoded message has invalid kind %d", msg.kind)
		}
	})
}

func FuzzPromptResponse(f *testing.F) {
	for _, seed := range []string{
		`{}`,
		`null`,
		`{"stopReason":null}`,
		`{"stopReason":"end_turn"}`,
		`{"stopReason":"cancelled","usage":null}`,
		`{"stopReason":"end_turn","usage":{"totalTokens":300,"inputTokens":100,"outputTokens":110,"thoughtTokens":30,"cachedReadTokens":40,"cachedWriteTokens":50,"_meta":{"provider":"test"}}}`,
		`{"stopReason":"end_turn","usage":{"totalTokens":0,"inputTokens":0,"outputTokens":0,"thoughtTokens":0,"cachedReadTokens":null}}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		assertStableJSONRoundTrip[PromptResponse](t, data)
	})
}

func FuzzIDKey(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`1`),
		[]byte(`1.0`),
		[]byte(`1e0`),
		[]byte(`"request"`),
		[]byte(`"\u0072equest"`),
		[]byte(`null`),
		[]byte(`true`),
		[]byte(`[]`),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, id []byte) {
		if len(id) > 1<<20 {
			t.Skip()
		}
		key, ok := idKey(id)
		if !ok {
			return
		}
		if key == "" {
			t.Fatal("valid request id has an empty key")
		}
		padded := make([]byte, 0, len(id)+4)
		padded = append(padded, ' ', '\n')
		padded = append(padded, id...)
		padded = append(padded, '\t', ' ')
		paddedKey, paddedOK := idKey(padded)
		if !paddedOK || paddedKey != key {
			t.Fatalf("whitespace changed id key from %q to %q", key, paddedKey)
		}
	})
}

func FuzzUnionUnmarshal(f *testing.F) {
	for _, seed := range []struct {
		kind uint8
		json string
	}{
		{kind: 0, json: `{"type":"text","text":"hello"}`},
		{kind: 0, json: `{"type":"resource","resource":{"uri":"file:///a","text":"a"}}`},
		{kind: 1, json: `{"uri":"file:///a","blob":"YQ=="}`},
		{kind: 2, json: `{"type":"diff","path":"a","newText":"b"}`},
		{kind: 3, json: `{"sessionUpdate":"usage_update","used":1,"size":2}`},
		{kind: 4, json: `{"sessionId":"s","mode":"form","message":"Choose","requestedSchema":{"type":"object","properties":{}}}`},
		{kind: 5, json: `{"action":"_defer","retryAfterMs":1000}`},
		{kind: 6, json: `{"type":"_location","precision":"city"}`},
		{kind: 7, json: `{"type":"_token","format":"workspace"}`},
		{kind: 0, json: `{}`},
	} {
		f.Add(seed.kind, []byte(seed.json))
	}

	f.Fuzz(func(t *testing.T, kind uint8, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		switch kind % 8 {
		case 0:
			assertStableJSONRoundTrip[ContentBlock](t, data)
		case 1:
			assertStableJSONRoundTrip[EmbeddedResourceContents](t, data)
		case 2:
			assertStableJSONRoundTrip[ToolCallContent](t, data)
		case 3:
			assertStableJSONRoundTrip[SessionUpdate](t, data)
		case 4:
			assertStableJSONRoundTrip[CreateElicitationRequest](t, data)
		case 5:
			assertStableJSONRoundTrip[CreateElicitationResponse](t, data)
		case 6:
			assertStableJSONRoundTrip[ElicitationPropertySchema](t, data)
		case 7:
			assertStableJSONRoundTrip[MultiSelectItems](t, data)
		}
	})
}

func assertStableJSONRoundTrip[T any](t *testing.T, data []byte) {
	t.Helper()
	var first T
	if err := json.Unmarshal(data, &first); err != nil {
		return
	}
	normalized, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	var second T
	if unmarshalErr := json.Unmarshal(normalized, &second); unmarshalErr != nil {
		t.Fatalf("normalized value does not decode: %v", unmarshalErr)
	}
	remarshaled, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !sameJSONShape(normalized, remarshaled) || !bytes.Equal(normalized, remarshaled) {
		t.Fatalf("round trip is unstable: %s != %s", normalized, remarshaled)
	}
}
