package acp

import (
	"encoding/json"
	"testing"
)

func TestSessionUpdateMessageContentRoundTrip(t *testing.T) {
	original := AgentMessageChunkSessionUpdate(TextContentBlock("hello"))
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SessionUpdate
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	content, ok := decoded.Content.(ContentBlock)
	if !ok {
		t.Fatalf("content type = %T, want ContentBlock", decoded.Content)
	}
	if content.Type != ContentBlockTypeText || content.Text != "hello" {
		t.Fatalf("content = %#v", content)
	}
}

func TestContentBlockRejectsMalformedKnownVariant(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "missing discriminator", json: `{"text":"hello"}`},
		{name: "unknown discriminator", json: `{"type":"video","data":"abc"}`},
		{name: "missing text", json: `{"type":"text"}`},
		{name: "missing image MIME type", json: `{"type":"image","data":"abc"}`},
		{name: "invalid image MIME type", json: `{"type":"image","data":"abc","mimeType":7}`},
		{name: "missing resource URI", json: `{"type":"resource_link","name":"file"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var block ContentBlock
			if err := json.Unmarshal([]byte(tt.json), &block); err == nil {
				t.Fatalf("json.Unmarshal(%s) succeeded: %#v", tt.json, block)
			}
		})
	}
}

func TestEnumUnmarshalAcceptsEscapedValue(t *testing.T) {
	var block ContentBlock
	if err := json.Unmarshal([]byte(`{"type":"te\u0078t","text":"hello"}`), &block); err != nil {
		t.Fatal(err)
	}
	if block.Type != ContentBlockTypeText {
		t.Fatalf("content block type = %q, want %q", block.Type, ContentBlockTypeText)
	}
}

func TestMCPServerNullTypeUsesDefaultVariant(t *testing.T) {
	var server MCPServer
	if err := json.Unmarshal([]byte(`{"type":null,"name":"server","command":"tool","args":[],"env":[]}`), &server); err != nil {
		t.Fatal(err)
	}
	if server.Type != "" || server.Command != "tool" {
		t.Fatalf("MCP server = %#v, want default stdio variant", server)
	}
}

func TestUnionMarshalKeepsRequiredZeroValues(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		fields []string
	}{
		{name: "empty text", value: TextContentBlock(""), fields: []string{"type", "text"}},
		{name: "empty plan", value: PlanSessionUpdate(nil), fields: []string{"sessionUpdate", "entries"}},
		{name: "zero usage", value: UsageUpdateSessionUpdate(0, 0), fields: []string{"sessionUpdate", "used", "size"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			for _, field := range test.fields {
				if _, ok := fields[field]; !ok {
					t.Fatalf("json.Marshal() = %s, missing required field %q", data, field)
				}
			}
		})
	}
}

func TestEmbeddedResourceRejectsMalformedVariant(t *testing.T) {
	tests := []string{
		`{}`,
		`{"uri":"file:///tmp/a"}`,
		`{"uri":"file:///tmp/a","text":"a","blob":"YQ=="}`,
	}
	for _, input := range tests {
		var resource EmbeddedResourceContents
		if err := json.Unmarshal([]byte(input), &resource); err == nil {
			t.Fatalf("json.Unmarshal(%s) succeeded: %#v", input, resource)
		}
	}
}

func TestToolCallContentRejectsMalformedKnownVariant(t *testing.T) {
	tests := []string{
		`{}`,
		`{"type":"unknown"}`,
		`{"type":"content"}`,
		`{"type":"diff","path":"/tmp/a"}`,
		`{"type":"terminal"}`,
	}
	for _, input := range tests {
		var content ToolCallContent
		if err := json.Unmarshal([]byte(input), &content); err == nil {
			t.Fatalf("json.Unmarshal(%s) succeeded: %#v", input, content)
		}
	}
}

func TestMetaIgnoresMalformedValue(t *testing.T) {
	var request InitializeRequest
	if err := json.Unmarshal([]byte(`{"protocolVersion":1,"_meta":7}`), &request); err != nil {
		t.Fatal(err)
	}
	if request.Meta != nil {
		t.Fatalf("meta = %#v, want nil", request.Meta)
	}
}

func TestDefaultOnErrorDoesNotKeepPartialValue(t *testing.T) {
	var block ContentBlock
	if err := json.Unmarshal([]byte(`{"type":"text","text":"hello","mimeType":7}`), &block); err != nil {
		t.Fatal(err)
	}
	if block.MIMEType != nil {
		t.Fatalf("mime type = %q, want nil", *block.MIMEType)
	}

	var request InitializeRequest
	if err := json.Unmarshal([]byte(`{"protocolVersion":1,"clientInfo":7}`), &request); err != nil {
		t.Fatal(err)
	}
	if request.ClientInfo != nil {
		t.Fatalf("client info = %#v, want nil", request.ClientInfo)
	}
}

func TestDefaultOnErrorDropsUnknownEnum(t *testing.T) {
	var update ToolCallUpdate
	if err := json.Unmarshal([]byte(`{"toolCallId":"call-1","kind":"bogus","status":"bogus"}`), &update); err != nil {
		t.Fatal(err)
	}
	if update.Kind != nil || update.Status != nil {
		t.Fatalf("kind = %v, status = %v; want nil defaults", update.Kind, update.Status)
	}

	var sessionUpdate SessionUpdate
	if err := json.Unmarshal([]byte(`{"sessionUpdate":"tool_call_update","toolCallId":"call-1","content":{"bad":true}}`), &sessionUpdate); err != nil {
		t.Fatal(err)
	}
	if sessionUpdate.Content != nil {
		t.Fatalf("content = %#v, want nil default", sessionUpdate.Content)
	}

	var annotations Annotations
	if err := json.Unmarshal([]byte(`{"audience":["assistant","bogus"]}`), &annotations); err != nil {
		t.Fatal(err)
	}
	if annotations.Audience == nil || len(*annotations.Audience) != 1 || (*annotations.Audience)[0] != RoleAssistant {
		t.Fatalf("audience = %#v, want only assistant", annotations.Audience)
	}
}

func TestArbitraryJSONNumbersRemainLossless(t *testing.T) {
	var request InitializeRequest
	if err := json.Unmarshal([]byte(`{"protocolVersion":1,"_meta":{"large":1e400}}`), &request); err != nil {
		t.Fatal(err)
	}
	large, ok := request.Meta["large"].(json.Number)
	if !ok || large != "1e400" {
		t.Fatalf("large metadata number = %#v, want json.Number(1e400)", request.Meta["large"])
	}
}

func TestMCPServerRejectsMissingTransportFields(t *testing.T) {
	tests := []string{
		`{"type":"http","name":"mcp"}`,
		`{"type":"sse","name":"mcp","url":"https://example.com"}`,
		`{"type":"websocket","name":"mcp","command":"/bin/mcp","args":[],"env":[]}`,
		`{"name":"mcp","command":"/bin/mcp"}`,
		`{"name":"mcp","command":"/bin/mcp","args":[],"env":[{}]}`,
	}
	for _, input := range tests {
		var server MCPServer
		if err := json.Unmarshal([]byte(input), &server); err == nil {
			t.Fatalf("json.Unmarshal(%s) succeeded: %#v", input, server)
		}
	}
}

func TestPermissionOptionAndPlanEntryRejectMalformedFields(t *testing.T) {
	for _, test := range []struct {
		input string
		value any
	}{
		{input: `{"optionId":"yes","name":"Yes"}`, value: &PermissionOption{}},
		{input: `{"optionId":"yes","name":"Yes","kind":"maybe"}`, value: &PermissionOption{}},
		{input: `{"content":"work","priority":"high"}`, value: &PlanEntry{}},
		{input: `{"content":"work","priority":"urgent","status":"pending"}`, value: &PlanEntry{}},
	} {
		if err := json.Unmarshal([]byte(test.input), test.value); err == nil {
			t.Fatalf("json.Unmarshal(%s) succeeded: %#v", test.input, test.value)
		}
	}
}

func TestSetSessionConfigOptionRequestValueRoundTrip(t *testing.T) {
	original := ValueIDSetSessionConfigOptionRequest("s", "model", "fast")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SetSessionConfigOptionRequest
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded.Value.(SessionConfigValueID); !ok {
		t.Fatalf("value type = %T, want SessionConfigValueID", decoded.Value)
	}
}

func TestSessionConfigOptionValueRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		value SessionConfigOption
		check func(*testing.T, any)
	}{
		{
			name:  "select",
			value: SelectSessionConfigOption("model", "Model", "fast", SessionConfigSelectOptions{}),
			check: func(t *testing.T, value any) {
				if _, ok := value.(SessionConfigValueID); !ok {
					t.Fatalf("current value type = %T, want SessionConfigValueID", value)
				}
			},
		},
		{
			name:  "boolean",
			value: BooleanSessionConfigOption("thinking", "Thinking", true),
			check: func(t *testing.T, value any) {
				if _, ok := value.(bool); !ok {
					t.Fatalf("current value type = %T, want bool", value)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			var decoded SessionConfigOption
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatal(err)
			}
			tt.check(t, decoded.CurrentValue)
		})
	}
}
