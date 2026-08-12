package acp

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
)

type schemaDefinition struct {
	Method   string   `json:"x-method"`
	Required []string `json:"required"`
	AnyOf    []struct {
		Required []string `json:"required"`
	} `json:"anyOf"`
}

func TestRequiredFieldMapsMatchSchema(t *testing.T) {
	data, err := os.ReadFile("schema/v1/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Definitions map[string]schemaDefinition `json:"$defs"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}

	expectedRequests := map[string][]string{}
	expectedResponses := map[string][]string{}
	for name, definition := range schema.Definitions {
		if definition.Method == "" {
			continue
		}
		fields := commonRequiredFields(definition)
		switch {
		case strings.HasSuffix(name, "Request"), strings.HasSuffix(name, "Notification"):
			expectedRequests[definition.Method] = fields
		case strings.HasSuffix(name, "Response"):
			expectedResponses[definition.Method] = fields
		default:
			t.Fatalf("method definition %q has an unknown message kind", name)
		}
	}

	actualRequests := normalizedFieldMap(requestRequiredFields)
	actualResponses := normalizedFieldMap(responseRequiredFields)
	if !maps.EqualFunc(expectedRequests, actualRequests, slices.Equal) {
		t.Fatalf("request field map does not match schema\nactual: %#v\nexpected: %#v", actualRequests, expectedRequests)
	}
	if !maps.EqualFunc(expectedResponses, actualResponses, slices.Equal) {
		t.Fatalf("response field map does not match schema\nactual: %#v\nexpected: %#v", actualResponses, expectedResponses)
	}
}

func commonRequiredFields(definition schemaDefinition) []string {
	fields := make(map[string]struct{}, len(definition.Required))
	for _, field := range definition.Required {
		fields[field] = struct{}{}
	}
	if len(definition.AnyOf) > 0 {
		common := make(map[string]struct{}, len(definition.AnyOf[0].Required))
		for _, field := range definition.AnyOf[0].Required {
			common[field] = struct{}{}
		}
		for _, alternative := range definition.AnyOf[1:] {
			alternativeFields := make(map[string]struct{}, len(alternative.Required))
			for _, field := range alternative.Required {
				alternativeFields[field] = struct{}{}
			}
			maps.DeleteFunc(common, func(field string, _ struct{}) bool {
				_, ok := alternativeFields[field]
				return !ok
			})
		}
		maps.Copy(fields, common)
	}
	result := slices.Collect(maps.Keys(fields))
	slices.Sort(result)
	return result
}

func normalizedFieldMap(source map[string][]string) map[string][]string {
	result := make(map[string][]string, len(source))
	for method, fields := range source {
		fields = slices.Clone(fields)
		slices.Sort(fields)
		if fields == nil {
			fields = []string{}
		}
		result[method] = fields
	}
	return result
}

func TestRequiredFieldValidationMatchesJSONKeySemantics(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{name: "escaped key", data: `{"sessio\u006eId":"s","prompt":[]}`},
		{name: "last duplicate is null", data: `{"sessionId":"s","sessionId":null,"prompt":[]}`, wantErr: true},
		{name: "last duplicate is value", data: `{"sessionId":null,"sessionId":"s","prompt":[]}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRequestFields([]byte(test.data), MethodSessionPrompt)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateRequestFields() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}

func TestRequiredFieldValidationAcceptsEncodingJSONDepth(t *testing.T) {
	const depth = 301
	meta := strings.Repeat(`{"value":`, depth) + `null` + strings.Repeat(`}`, depth)
	data := []byte(`{"sessionId":"s","prompt":[],"_meta":` + meta + `}`)
	if err := validateRequestFields(data, MethodSessionPrompt); err != nil {
		t.Fatal(err)
	}
}
