package acp

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
	"testing"
	"time"
)

type elicitationClient struct {
	completed chan ElicitationID
}

func (*elicitationClient) RequestPermission(context.Context, *RequestPermissionRequest) (*RequestPermissionResponse, error) {
	return &RequestPermissionResponse{Outcome: CanceledRequestPermissionOutcome()}, nil
}

func (*elicitationClient) Update(context.Context, *SessionNotification) error {
	return nil
}

func (*elicitationClient) CreateElicitation(_ context.Context, _ *CreateElicitationRequest) (*CreateElicitationResponse, error) {
	content := map[string]ElicitationContentValue{"strategy": ElicitationContentValue(`"balanced"`)}
	response := AcceptCreateElicitationResponse()
	response.Content = &content
	return &response, nil
}

func (c *elicitationClient) CompleteElicitation(_ context.Context, notification *CompleteElicitationNotification) error {
	c.completed <- notification.ElicitationID
	return nil
}

func TestElicitationRoundTrip(t *testing.T) {
	agentSide, clientSide := net.Pipe()
	var clientCaller *ClientCaller
	agentConn, err := NewAgent(agentSide, agentSide, func(client *ClientCaller) AgentHandler {
		clientCaller = client
		return &testAgent{client: client}
	})
	if err != nil {
		t.Fatal(err)
	}
	completed := make(chan ElicitationID, 1)
	clientConn, err := NewClient(clientSide, clientSide, func(*AgentCaller) ClientHandler {
		return &elicitationClient{completed: completed}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = agentConn.Close()
	})

	strategy := StringElicitationPropertySchema()
	strategy.Enum = &[]string{"conservative", "balanced"}
	request := SessionFormCreateElicitationRequest("Choose a strategy", ElicitationSchema{
		Type: ElicitationSchemaTypeObject,
		Properties: map[string]ElicitationPropertySchema{
			"strategy": strategy,
		},
	}, "session-1")
	response, err := clientCaller.CreateElicitation(t.Context(), &request)
	if err != nil {
		t.Fatal(err)
	}
	if response.Action != CreateElicitationResponseTypeAccept || response.Content == nil || string((*response.Content)["strategy"]) != `"balanced"` {
		t.Fatalf("elicitation response = %#v", response)
	}

	const elicitationID ElicitationID = "oauth-1"
	if err := clientCaller.CompleteElicitation(t.Context(), &CompleteElicitationNotification{ElicitationID: elicitationID}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-completed:
		if got != elicitationID {
			t.Fatalf("completed elicitation id = %q, want %q", got, elicitationID)
		}
	case <-time.After(time.Second):
		t.Fatal("elicitation completion was not delivered")
	}
}

func TestElicitationRejectsMalformedKnownVariants(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		value any
	}{
		{name: "form without schema", input: `{"mode":"form","message":"missing"}`, value: &CreateElicitationRequest{}},
		{name: "url without id", input: `{"mode":"url","message":"missing","url":"https://example.com","sessionId":"s"}`, value: &CreateElicitationRequest{}},
		{name: "request without scope", input: `{"mode":"form","message":"missing","requestedSchema":{}}`, value: &CreateElicitationRequest{}},
		{name: "response without action", input: `{}`, value: &CreateElicitationResponse{}},
		{name: "array property without items", input: `{"type":"array"}`, value: &ElicitationPropertySchema{}},
		{name: "property without type", input: `{}`, value: &ElicitationPropertySchema{}},
		{name: "string items without enum", input: `{"type":"string"}`, value: &MultiSelectItems{}},
		{name: "multi-select null type", input: `{"type":null}`, value: &MultiSelectItems{}},
		{name: "titled items without anyOf", input: `{}`, value: &MultiSelectItems{}},
		{name: "enum option without const", input: `{"title":"Title"}`, value: &EnumOption{}},
		{name: "enum option without title", input: `{"const":"value"}`, value: &EnumOption{}},
		{name: "unknown string format", input: `"hostname"`, value: new(StringFormat)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(test.input), test.value); err == nil {
				t.Fatalf("json.Unmarshal(%s) succeeded: %#v", test.input, test.value)
			}
		})
	}
}

func TestElicitationPreservesUnknownVariants(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		value any
	}{
		{
			name:  "request mode",
			input: `{"requestId":42,"mode":"_browser","message":"Open login","target":{"provider":"github"}}`,
			value: &CreateElicitationRequest{},
		},
		{
			name:  "response action",
			input: `{"action":"_defer","reason":"waiting","retryAfterMs":1000}`,
			value: &CreateElicitationResponse{},
		},
		{
			name:  "property type",
			input: `{"type":"_location","title":"Location","precision":"city"}`,
			value: &ElicitationPropertySchema{},
		},
		{
			name:  "multi-select item type",
			input: `{"type":"_token","format":"workspace","anyOf":[{"const":"repo","title":"Repository"}]}`,
			value: &MultiSelectItems{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(test.input), test.value); err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			var gotValue, expectedValue any
			if err := json.Unmarshal(got, &gotValue); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.input), &expectedValue); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotValue, expectedValue) {
				t.Fatalf("round trip = %s, want %s", got, test.input)
			}
		})
	}
}

func TestElicitationContentValueValidation(t *testing.T) {
	for _, input := range []string{`"text"`, `42`, `3.5`, `true`, `["a","b"]`} {
		t.Run("accept "+input, func(t *testing.T) {
			var response CreateElicitationResponse
			if err := json.Unmarshal([]byte(`{"action":"accept","content":{"value":`+input+`}}`), &response); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, input := range []string{`null`, `{}`, `["a",1]`, `1e9999`} {
		t.Run("reject "+input, func(t *testing.T) {
			var response CreateElicitationResponse
			if err := json.Unmarshal([]byte(`{"action":"accept","content":{"value":`+input+`}}`), &response); err == nil {
				t.Fatalf("accepted invalid content value %s", input)
			}
		})
	}
}

func TestElicitationReaderDefaults(t *testing.T) {
	var schema ElicitationSchema
	if err := json.Unmarshal([]byte(`{"type":null,"title":42,"description":false}`), &schema); err != nil {
		t.Fatal(err)
	}
	if schema.Type != ElicitationSchemaTypeObject || schema.Properties == nil || len(schema.Properties) != 0 {
		t.Fatalf("schema defaults = %#v", schema)
	}
	encoded, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["type"]) != `"object"` || string(fields["properties"]) != `{}` {
		t.Fatalf("schema JSON = %s", encoded)
	}

	var capabilities ClientCapabilities
	if err := json.Unmarshal([]byte(`{"elicitation":false,"fs":false,"session":false,"terminal":"yes"}`), &capabilities); err != nil {
		t.Fatal(err)
	}
	if capabilities.Elicitation != nil || capabilities.Fs != nil || capabilities.Session != nil || capabilities.Terminal {
		t.Fatalf("capability defaults = %#v", capabilities)
	}
}

func TestElicitationPropertyDefaultTypes(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		expected any
	}{
		{name: "string", input: `{"type":"string","default":1}`, expected: nil},
		{name: "number", input: `{"type":"number","default":"1"}`, expected: nil},
		{name: "integer", input: `{"type":"integer","default":1.5}`, expected: nil},
		{name: "boolean", input: `{"type":"boolean","default":"true"}`, expected: nil},
		{name: "array", input: `{"type":"array","items":{"type":"string","enum":[]},"default":["a",1,"b"]}`, expected: []string{"a", "b"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var schema ElicitationPropertySchema
			if err := json.Unmarshal([]byte(test.input), &schema); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(schema.Default, test.expected) {
				t.Fatalf("default = %#v, want %#v", schema.Default, test.expected)
			}
		})
	}
}

func TestElicitationScopeSelection(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "request scope ignores malformed session field",
			input:    `{"sessionId":7,"requestId":42,"mode":"form","message":"Choose","requestedSchema":{}}`,
			expected: `{"requestId":42,"mode":"form","message":"Choose","requestedSchema":{"type":"object","properties":{}}}`,
		},
		{
			name:     "session scope wins when both are valid",
			input:    `{"sessionId":"","requestId":42,"mode":"form","message":"Choose","requestedSchema":{}}`,
			expected: `{"sessionId":"","mode":"form","message":"Choose","requestedSchema":{"type":"object","properties":{}}}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var request CreateElicitationRequest
			if err := json.Unmarshal([]byte(test.input), &request); err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var gotValue, expectedValue any
			if err := json.Unmarshal(got, &gotValue); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.expected), &expectedValue); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotValue, expectedValue) {
				t.Fatalf("scope normalization = %s, want %s", got, test.expected)
			}
		})
	}
}
