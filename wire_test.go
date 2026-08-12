package acp

import (
	"encoding/json"
	"testing"
)

func TestIDKeyMatchesACPRequestIDSchema(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{name: "string", id: `"request"`, valid: true},
		{name: "null", id: `null`, valid: true},
		{name: "integer", id: `42`, valid: true},
		{name: "integer decimal", id: `42.0`, valid: true},
		{name: "integer exponent", id: `42e1`, valid: true},
		{name: "minimum int64", id: `-9223372036854775808`, valid: true},
		{name: "maximum int64", id: `9223372036854775807`, valid: true},
		{name: "fraction", id: `1.5`, valid: false},
		{name: "below int64", id: `-9223372036854775809`, valid: false},
		{name: "above int64", id: `9223372036854775808`, valid: false},
		{name: "leading zero", id: `01`, valid: false},
		{name: "negative leading zero", id: `-01`, valid: false},
		{name: "leading plus", id: `+1`, valid: false},
		{name: "boolean", id: `true`, valid: false},
		{name: "array", id: `[]`, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, valid := idKey(json.RawMessage(tt.id))
			if valid != tt.valid {
				t.Fatalf("idKey(%s) valid = %t, want %t", tt.id, valid, tt.valid)
			}
		})
	}
}

func TestDecodeMessageMatchesJSONRPCFieldNamesExactly(t *testing.T) {
	if _, rpcErr := decodeMessage([]byte(`{"JSONRPC":"2.0","ID":1,"Method":"x"}`)); rpcErr == nil {
		t.Fatal("decodeMessage accepted non-canonical field names")
	}

	msg, rpcErr := decodeMessage([]byte(`{"jsonrpc":"2.0","JSONRPC":"1.0","id":1,"method":"x","Method":7}`))
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	if msg.kind != messageRequest || msg.method != "x" {
		t.Fatalf("decodeMessage() = %#v, want canonical request fields", msg)
	}
}
