package acp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/valyala/fastjson"
)

// Required-field checks should not build a second reflection-backed object map.
const maxFastJSONValidationBytes = 4 << 10

var jsonParserPool fastjson.ParserPool

// UnmarshalJSON implements json.Unmarshaler.
func (m *Meta) UnmarshalJSON(data []byte) error {
	raw, err := decodeJSONValue(data)
	value, ok := raw.(map[string]any)
	if err != nil || !ok {
		*m = nil
		return nil //nolint:nilerr // ACP specifies malformed _meta as a nil default.
	}
	*m = value
	return nil
}

func unmarshalEnum[T ~string](data []byte, field string, dst *T, allowed ...T) error {
	if len(data) > maxFastJSONValidationBytes {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		return setEnum([]byte(raw), field, dst, allowed)
	}
	parser := jsonParserPool.Get()
	defer jsonParserPool.Put(parser)
	value, err := parser.ParseBytes(data)
	if err != nil {
		return err
	}
	if value.Type() == fastjson.TypeNull {
		return setEnum(nil, field, dst, allowed)
	}
	raw, err := value.StringBytes()
	if err != nil {
		return err
	}
	return setEnum(raw, field, dst, allowed)
}

func setEnum[T ~string](raw []byte, field string, dst *T, allowed []T) error {
	for _, candidate := range allowed {
		if bytes.Equal(raw, []byte(candidate)) {
			*dst = candidate
			return nil
		}
	}
	return invalidDiscriminator(field, string(raw))
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *ContentBlockType) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "content block type", t, ContentBlockTypeText, ContentBlockTypeImage, ContentBlockTypeAudio, ContentBlockTypeResourceLink, ContentBlockTypeResource)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *MCPServerType) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "mcp server type", t, "", MCPServerTypeHTTP, MCPServerTypeSSE)
}

// UnmarshalJSON implements json.Unmarshaler.
func (k *PermissionOptionKind) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "permission option kind", k, PermissionOptionKindAllowOnce, PermissionOptionKindAllowAlways, PermissionOptionKindRejectOnce, PermissionOptionKindRejectAlways)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PlanEntryPriority) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "plan entry priority", p, PlanEntryPriorityHigh, PlanEntryPriorityMedium, PlanEntryPriorityLow)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *PlanEntryStatus) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "plan entry status", s, PlanEntryStatusPending, PlanEntryStatusInProgress, PlanEntryStatusCompleted)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *RequestPermissionOutcomeType) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "permission outcome", t, RequestPermissionOutcomeTypeCanceled, RequestPermissionOutcomeTypeSelected)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Role) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "role", r, RoleAssistant, RoleUser)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *SessionConfigOptionType) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "session config option type", t, SessionConfigOptionTypeSelect, SessionConfigOptionTypeBoolean)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *SessionUpdateType) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "session update type", t,
		SessionUpdateTypeUserMessageChunk, SessionUpdateTypeAgentMessageChunk, SessionUpdateTypeAgentThoughtChunk,
		SessionUpdateTypeToolCall, SessionUpdateTypeToolCallUpdate, SessionUpdateTypePlan,
		SessionUpdateTypeAvailableCommandsUpdate, SessionUpdateTypeCurrentModeUpdate, SessionUpdateTypeConfigOptionUpdate,
		SessionUpdateTypeSessionInfoUpdate, SessionUpdateTypeUsageUpdate)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *ToolCallContentType) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "tool call content type", t, ToolCallContentTypeContent, ToolCallContentTypeDiff, ToolCallContentTypeTerminal)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *ToolCallStatus) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "tool call status", s, ToolCallStatusPending, ToolCallStatusInProgress, ToolCallStatusCompleted, ToolCallStatusFailed)
}

// UnmarshalJSON implements json.Unmarshaler.
func (k *ToolKind) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "tool kind", k,
		ToolKindRead, ToolKindEdit, ToolKindDelete, ToolKindMove, ToolKindSearch,
		ToolKindExecute, ToolKindThink, ToolKindFetch, ToolKindSwitchMode, ToolKindOther)
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *StringFormat) UnmarshalJSON(data []byte) error {
	return unmarshalEnum(data, "string format", f, StringFormatEmail, StringFormatURI, StringFormatDate, StringFormatDateTime)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *ElicitationContentValue) UnmarshalJSON(data []byte) error {
	if err := validateElicitationContentValue(data); err != nil {
		return err
	}
	*v = append((*v)[:0], data...)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (v ElicitationContentValue) MarshalJSON() ([]byte, error) {
	if err := validateElicitationContentValue(v); err != nil {
		return nil, err
	}
	return v, nil
}

// NewElicitationContentValue converts a supported Go value to its wire representation.
func NewElicitationContentValue(value any) (ElicitationContentValue, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if err := validateElicitationContentValue(data); err != nil {
		return nil, err
	}
	return ElicitationContentValue(data), nil
}

func validateElicitationContentValue(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || !json.Valid(data) {
		return errors.New("acp: invalid elicitation content value")
	}
	switch data[0] {
	case '"':
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("acp: invalid elicitation content value: %w", err)
		}
		return nil
	case 't', 'f':
		var value bool
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("acp: invalid elicitation content value: %w", err)
		}
		return nil
	case '[':
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil || values == nil {
			return errors.New("acp: elicitation content value must be a string array")
		}
		for _, raw := range values {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return errors.New("acp: elicitation content value must be a string array")
			}
		}
		return nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if _, err := strconv.ParseFloat(string(data), 64); err != nil {
			return fmt.Errorf("acp: invalid elicitation content value: %w", err)
		}
		return nil
	default:
		return errors.New("acp: elicitation content value must be a string, number, boolean, or string array")
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *CompleteElicitationNotification) UnmarshalJSON(data []byte) error {
	type wire CompleteElicitationNotification
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "elicitationId"); err != nil {
		return err
	}
	*n = CompleteElicitationNotification(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *ElicitationRequestScope) UnmarshalJSON(data []byte) error {
	fields, err := requireJSONKeys(data, "requestId")
	if err != nil {
		return err
	}
	if !validID(fields["requestId"]) {
		return errors.New("acp: invalid elicitation request id")
	}
	s.RequestID = append(s.RequestID[:0], fields["requestId"]...)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *StringMultiSelectItems) UnmarshalJSON(data []byte) error {
	type wire StringMultiSelectItems
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "enum"); err != nil {
		return err
	}
	*i = StringMultiSelectItems(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *TitledMultiSelectItems) UnmarshalJSON(data []byte) error {
	type wire TitledMultiSelectItems
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "anyOf"); err != nil {
		return err
	}
	*i = TitledMultiSelectItems(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *HTTPHeader) UnmarshalJSON(data []byte) error {
	type wire HTTPHeader
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "name", "value"); err != nil {
		return err
	}
	*h = HTTPHeader(value)
	return nil
}

func validateJSONFieldsStdlib(data []byte, required, nonNull []string) error {
	fields, err := requireJSONKeys(data, required...)
	if err != nil {
		return err
	}
	return requireFields(fields, nonNull...)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *EnvVariable) UnmarshalJSON(data []byte) error {
	type wire EnvVariable
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "name", "value"); err != nil {
		return err
	}
	*e = EnvVariable(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *AvailableCommandInput) UnmarshalJSON(data []byte) error {
	type wire AvailableCommandInput
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "hint"); err != nil {
		return err
	}
	*i = AvailableCommandInput(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Cost) UnmarshalJSON(data []byte) error {
	type wire Cost
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "amount", "currency"); err != nil {
		return err
	}
	*c = Cost(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *MCPServer) UnmarshalJSON(data []byte) error {
	type wire MCPServer
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value.Type {
	case MCPServerTypeHTTP, MCPServerTypeSSE:
		if err := requireJSONFieldsOnly(data, "type", "name", "url", "headers"); err != nil {
			return err
		}
	case "":
		if err := requireJSONFieldsOnly(data, "name", "command", "args", "env"); err != nil {
			return err
		}
	default:
		return invalidDiscriminator("type", value.Type)
	}
	*s = MCPServer(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *PermissionOption) UnmarshalJSON(data []byte) error {
	type wire PermissionOption
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "optionId", "name", "kind"); err != nil {
		return err
	}
	switch value.Kind {
	case PermissionOptionKindAllowOnce,
		PermissionOptionKindAllowAlways,
		PermissionOptionKindRejectOnce,
		PermissionOptionKindRejectAlways:
	default:
		return invalidDiscriminator("kind", value.Kind)
	}
	*o = PermissionOption(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *PlanEntry) UnmarshalJSON(data []byte) error {
	type wire PlanEntry
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "content", "priority", "status"); err != nil {
		return err
	}
	switch value.Priority {
	case PlanEntryPriorityHigh, PlanEntryPriorityMedium, PlanEntryPriorityLow:
	default:
		return invalidDiscriminator("priority", value.Priority)
	}
	switch value.Status {
	case PlanEntryStatusPending, PlanEntryStatusInProgress, PlanEntryStatusCompleted:
	default:
		return invalidDiscriminator("status", value.Status)
	}
	*e = PlanEntry(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *SetSessionConfigOptionRequest) UnmarshalJSON(data []byte) error {
	type wire SetSessionConfigOptionRequest
	var value wire
	var raw struct {
		Value json.RawMessage `json:"value"`
		*wire
	}
	raw.wire = &value
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if err := requireJSONFieldsOnly(data, "sessionId", "configId", "value"); err != nil {
		return err
	}
	if value.Type == SetSessionConfigOptionRequestTypeBoolean {
		var boolean bool
		if err := json.Unmarshal(raw.Value, &boolean); err != nil {
			return err
		}
		value.Value = boolean
	} else {
		var id SessionConfigValueID
		if err := json.Unmarshal(raw.Value, &id); err != nil {
			return err
		}
		value.Value = id
	}
	*r = SetSessionConfigOptionRequest(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *RequestPermissionOutcome) UnmarshalJSON(data []byte) error {
	type wire RequestPermissionOutcome
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value.Outcome {
	case RequestPermissionOutcomeTypeCanceled:
		if err := requireJSONFieldsOnly(data, "outcome"); err != nil {
			return err
		}
	case RequestPermissionOutcomeTypeSelected:
		if err := requireJSONFieldsOnly(data, "outcome", "optionId"); err != nil {
			return err
		}
	default:
		return invalidDiscriminator("outcome", value.Outcome)
	}
	*o = RequestPermissionOutcome(value)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *StopReason) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch StopReason(value) {
	case StopReasonEndTurn, StopReasonMaxTokens, StopReasonMaxTurnRequests, StopReasonRefusal, StopReasonCanceled:
		*r = StopReason(value)
		return nil
	default:
		return invalidDiscriminator("stopReason", value)
	}
}

var requestRequiredFields = map[string][]string{
	MethodInitialize:               {"protocolVersion"},
	MethodAuthenticate:             {"methodId"},
	MethodLogout:                   nil,
	MethodSessionNew:               {"cwd", "mcpServers"},
	MethodSessionLoad:              {"sessionId", "cwd", "mcpServers"},
	MethodSessionSetMode:           {"sessionId", "modeId"},
	MethodSessionSetConfigOption:   {"sessionId", "configId", "value"},
	MethodSessionPrompt:            {"sessionId", "prompt"},
	MethodSessionCancel:            {"sessionId"},
	MethodSessionList:              nil,
	MethodSessionDelete:            {"sessionId"},
	MethodSessionResume:            {"sessionId", "cwd"},
	MethodSessionClose:             {"sessionId"},
	MethodSessionRequestPermission: {"sessionId", "toolCall", "options"},
	MethodSessionUpdate:            {"sessionId", "update"},
	MethodFSWriteTextFile:          {"sessionId", "path", "content"},
	MethodFSReadTextFile:           {"sessionId", "path"},
	MethodTerminalCreate:           {"sessionId", "command"},
	MethodTerminalOutput:           {"sessionId", "terminalId"},
	MethodTerminalRelease:          {"sessionId", "terminalId"},
	MethodTerminalWaitForExit:      {"sessionId", "terminalId"},
	MethodTerminalKill:             {"sessionId", "terminalId"},
	MethodElicitationCreate:        {"mode", "message"},
	MethodElicitationComplete:      {"elicitationId"},
	MethodCancelRequest:            {"requestId"},
}

var responseRequiredFields = map[string][]string{
	MethodInitialize:               {"protocolVersion"},
	MethodAuthenticate:             nil,
	MethodLogout:                   nil,
	MethodSessionNew:               {"sessionId"},
	MethodSessionLoad:              nil,
	MethodSessionSetMode:           nil,
	MethodSessionSetConfigOption:   {"configOptions"},
	MethodSessionPrompt:            {"stopReason"},
	MethodSessionList:              {"sessions"},
	MethodSessionDelete:            nil,
	MethodSessionResume:            nil,
	MethodSessionClose:             nil,
	MethodSessionRequestPermission: {"outcome"},
	MethodFSWriteTextFile:          nil,
	MethodFSReadTextFile:           {"content"},
	MethodTerminalCreate:           {"terminalId"},
	MethodTerminalOutput:           {"output", "truncated"},
	MethodTerminalRelease:          nil,
	MethodTerminalWaitForExit:      nil,
	MethodTerminalKill:             nil,
	MethodElicitationCreate:        {"action"},
}

var requestNonNullFields = map[string][]string{
	MethodInitialize:               {"protocolVersion"},
	MethodAuthenticate:             {"methodId"},
	MethodSessionNew:               {"cwd"},
	MethodSessionLoad:              {"sessionId", "cwd"},
	MethodSessionSetMode:           {"sessionId", "modeId"},
	MethodSessionSetConfigOption:   {"sessionId", "configId", "value"},
	MethodSessionPrompt:            {"sessionId", "prompt"},
	MethodSessionCancel:            {"sessionId"},
	MethodSessionDelete:            {"sessionId"},
	MethodSessionResume:            {"sessionId", "cwd"},
	MethodSessionClose:             {"sessionId"},
	MethodSessionRequestPermission: {"sessionId", "toolCall", "options"},
	MethodSessionUpdate:            {"sessionId", "update"},
	MethodFSWriteTextFile:          {"sessionId", "path", "content"},
	MethodFSReadTextFile:           {"sessionId", "path"},
	MethodTerminalCreate:           {"sessionId", "command"},
	MethodTerminalOutput:           {"sessionId", "terminalId"},
	MethodTerminalRelease:          {"sessionId", "terminalId"},
	MethodTerminalWaitForExit:      {"sessionId", "terminalId"},
	MethodTerminalKill:             {"sessionId", "terminalId"},
	MethodElicitationCreate:        {"mode", "message"},
	MethodElicitationComplete:      {"elicitationId"},
}

var responseNonNullFields = map[string][]string{
	MethodInitialize:               {"protocolVersion"},
	MethodSessionNew:               {"sessionId"},
	MethodSessionPrompt:            {"stopReason"},
	MethodSessionRequestPermission: {"outcome"},
	MethodFSReadTextFile:           {"content"},
	MethodTerminalCreate:           {"terminalId"},
	MethodTerminalOutput:           {"output", "truncated"},
	MethodElicitationCreate:        {"action"},
}

func requireJSONFields(data []byte, names ...string) (map[string]json.RawMessage, error) {
	fields, err := requireJSONKeys(data, names...)
	if err != nil {
		return nil, err
	}
	if err := requireFields(fields, names...); err != nil {
		return nil, err
	}
	return fields, nil
}

func requireFields(fields map[string]json.RawMessage, names ...string) error {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("acp: missing required field %q", name)
		}
	}
	return nil
}

func elicitationScopeIsSession(fields map[string]json.RawMessage) (bool, error) {
	if raw, ok := fields["sessionId"]; ok && !jsonValueIsNull(raw) {
		var sessionID SessionID
		if json.Unmarshal(raw, &sessionID) == nil {
			return true, nil
		}
	}
	if raw, ok := fields["requestId"]; ok && validID(raw) {
		return false, nil
	}
	return false, errors.New("acp: elicitation requires sessionId or requestId")
}

func jsonValueIsNull(data []byte) bool {
	return bytes.Equal(bytes.TrimSpace(data), []byte("null"))
}

func marshalJSONWithFields(value any, fields map[string]json.RawMessage) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil || len(fields) == 0 {
		return data, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	for name, raw := range fields {
		if _, exists := object[name]; exists {
			continue
		}
		if !json.Valid(raw) {
			return nil, fmt.Errorf("acp: invalid raw field %q", name)
		}
		object[name] = raw
	}
	return json.Marshal(object)
}

func requireJSONKeys(data []byte, names ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, errors.New("acp: value must be an object")
	}
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return nil, fmt.Errorf("acp: missing required field %q", name)
		}
	}
	return fields, nil
}

func invalidDiscriminator(field string, value any) error {
	return fmt.Errorf("acp: invalid %s discriminator %q", field, value)
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() { //nolint:exhaustive // Only nilable reflection kinds can report nil.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func unmarshalDefault[T any](data []byte, dst *T) {
	if reflect.TypeFor[T]().Kind() == reflect.Interface {
		decoded, err := decodeJSONValue(data)
		if err != nil {
			return
		}
		value, ok := decoded.(T)
		if ok {
			*dst = value
		}
		return
	}
	var value T
	if json.Unmarshal(data, &value) == nil {
		*dst = value
	}
}

func decodeJSONValue(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	err := decoder.Decode(&value)
	return value, err
}

func validateTypedJSON(data []byte, value any) error {
	t := reflect.TypeOf(value)
	if t == nil {
		return errors.New("acp: value is nil")
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	decoded := reflect.New(t).Interface()
	if err := json.Unmarshal(data, decoded); err != nil {
		return err
	}
	normalized, err := json.Marshal(decoded)
	if err != nil {
		return err
	}
	// Values produced by json.Marshal normally round-trip byte-for-byte. Keep
	// the structural comparison only for RawMessage and other custom JSON.
	if bytes.Equal(data, normalized) {
		return nil
	}
	if !sameJSONShape(data, normalized) {
		return errors.New("acp: typed value loses fields during validation")
	}
	return nil
}

func sameJSONShape(left, right []byte) bool {
	leftValue, err := decodeJSONValue(left)
	if err != nil {
		return false
	}
	rightValue, err := decodeJSONValue(right)
	if err != nil {
		return false
	}
	return sameValueShape(leftValue, rightValue)
}

func sameValueShape(left, right any) bool {
	switch left := left.(type) {
	case map[string]any:
		right, ok := right.(map[string]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for key, leftValue := range left {
			rightValue, ok := right[key]
			if !ok || !sameValueShape(leftValue, rightValue) {
				return false
			}
		}
		return true
	case []any:
		right, ok := right.([]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for i := range left {
			if !sameValueShape(left[i], right[i]) {
				return false
			}
		}
		return true
	case nil:
		return right == nil
	case string:
		_, ok := right.(string)
		return ok
	case json.Number:
		_, ok := right.(json.Number)
		return ok
	case bool:
		_, ok := right.(bool)
		return ok
	default:
		return false
	}
}

func validateRequestFields(data []byte, method string) error {
	return validateRequiredFields(data, method, requestRequiredFields, requestNonNullFields)
}

func validateResponseFields(data []byte, method string) error {
	return validateRequiredFields(data, method, responseRequiredFields, responseNonNullFields)
}

func validateRequiredFields(data []byte, method string, required, nonNull map[string][]string) error {
	names, ok := required[method]
	if !ok {
		return nil
	}
	return validateJSONFields(data, names, nonNull[method])
}

func validateJSONFields(data []byte, required, nonNull []string) error {
	if len(data) > maxFastJSONValidationBytes {
		return validateJSONFieldsStdlib(data, required, nonNull)
	}
	parser := jsonParserPool.Get()
	defer jsonParserPool.Put(parser)
	value, err := parser.ParseBytes(data)
	if err != nil {
		// fastjson deliberately caps nesting below encoding/json's limit. Keep
		// the standard-library semantics for unusually deep, valid values.
		return validateJSONFieldsStdlib(data, required, nonNull)
	}
	fields, err := value.Object()
	if err != nil {
		return errors.New("acp: value must be an object")
	}
	for _, name := range required {
		if lastJSONField(fields, name) == nil {
			return fmt.Errorf("acp: missing required field %q", name)
		}
	}
	for _, name := range nonNull {
		field := lastJSONField(fields, name)
		if field == nil || field.Type() == fastjson.TypeNull {
			return fmt.Errorf("acp: missing required field %q", name)
		}
	}
	return nil
}

func lastJSONField(fields *fastjson.Object, name string) *fastjson.Value {
	var result *fastjson.Value
	fields.Visit(func(key []byte, value *fastjson.Value) {
		if string(key) == name {
			result = value
		}
	})
	return result
}

func requireJSONFieldsOnly(data []byte, names ...string) error {
	return validateJSONFields(data, names, names)
}

func requireJSONKeysOnly(data []byte, names ...string) error {
	return validateJSONFields(data, names, nil)
}
