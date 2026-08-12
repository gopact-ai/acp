package acp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"sync"
	"unicode/utf8"
)

// A map preserves JSON-RPC's case-sensitive member names; pooling keeps that
// exact-key behavior without allocating a new object map for every frame.
var messageFieldsPool = sync.Pool{New: func() any {
	return make(map[string]json.RawMessage, 6)
}}

type messageKind uint8

const (
	messageRequest messageKind = iota + 1
	messageNotification
	messageResponse
)

type message struct {
	kind   messageKind
	id     json.RawMessage
	method string
	params json.RawMessage
	result json.RawMessage
	err    *Error
}

type requestEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type notificationEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type responseEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func marshalParams(params any) (json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	if !validParams(b) {
		return nil, errors.New("params must be an object or array")
	}
	return b, nil
}

func paramsOrEmpty(params json.RawMessage) json.RawMessage {
	if len(params) == 0 {
		return json.RawMessage("{}")
	}
	return params
}

func readFrame(reader *bufio.Reader, limit int) ([]byte, error) {
	fragment, isPrefix, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	if len(fragment) > limit {
		return nil, ErrFrameTooLarge
	}
	if !isPrefix {
		// readLoop consumes the frame before the reader advances, so keep the
		// reader-owned slice instead of allocating for the common small-frame case.
		return fragment, nil
	}

	frame := make([]byte, 0, min(limit, 4096))
	frame = append(frame, fragment...)
	for {
		fragment, isPrefix, err = reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if len(frame)+len(fragment) > limit {
			return nil, ErrFrameTooLarge
		}
		frame = append(frame, fragment...)
		if !isPrefix {
			return frame, nil
		}
	}
}

func decodeMessage(frame []byte) (message, *Error) {
	if !utf8.Valid(frame) {
		return message{}, parseError(errors.New("invalid JSON"))
	}

	pooledFields, ok := messageFieldsPool.Get().(map[string]json.RawMessage)
	if !ok {
		pooledFields = make(map[string]json.RawMessage, 6)
	}
	fields := pooledFields
	defer func() {
		if len(pooledFields) <= 16 {
			clear(pooledFields)
			messageFieldsPool.Put(pooledFields)
		}
	}()
	if err := json.Unmarshal(frame, &fields); err != nil {
		var syntaxError *json.SyntaxError
		if errors.As(err, &syntaxError) {
			return message{}, parseError(errors.New("invalid JSON"))
		}
		return message{}, invalidRequest(err)
	}
	if fields == nil {
		return message{}, invalidRequest(errors.New("message must be an object"))
	}

	id, hasID := fields["id"]
	methodJSON, hasMethod := fields["method"]
	result, hasResult := fields["result"]
	errorJSON, hasError := fields["error"]
	msg := message{id: id}
	if hasMethod {
		msg.kind = messageNotification
		if hasID {
			msg.kind = messageRequest
		}
	} else if hasResult || hasError {
		msg.kind = messageResponse
	}
	invalid := func(err error) (message, *Error) {
		return msg, invalidRequest(err)
	}

	if hasID {
		if _, ok := idKey(id); !ok {
			return invalid(errors.New("invalid request id"))
		}
	}

	var version string
	if err := json.Unmarshal(fields["jsonrpc"], &version); err != nil || version != "2.0" {
		return invalid(errors.New("jsonrpc must be \"2.0\""))
	}

	params := fields["params"]
	if len(params) > 0 && !validParams(params) {
		return invalid(errors.New("params must be an object or array"))
	}

	if hasMethod {
		if hasResult || hasError {
			return invalid(errors.New("request contains response fields"))
		}
		var method string
		if err := json.Unmarshal(methodJSON, &method); err != nil {
			return invalid(errors.New("method must be a string"))
		}
		msg.method = method
		msg.params = params
		return msg, nil
	}

	if !hasID || hasResult == hasError {
		return invalid(errors.New("invalid response shape"))
	}
	msg.result = result
	if hasError {
		if _, err := requireJSONFields(errorJSON, "code", "message"); err != nil {
			return invalid(errors.New("invalid error response"))
		}
		var rpcErr Error
		if err := json.Unmarshal(errorJSON, &rpcErr); err != nil {
			return invalid(errors.New("invalid error response"))
		}
		msg.err = &rpcErr
	}
	return msg, nil
}

func validID(id json.RawMessage) bool {
	_, ok := idKey(id)
	return ok
}

func idKey(id json.RawMessage) (string, bool) {
	id = bytes.TrimSpace(id)
	if len(id) == 0 {
		return "", false
	}
	switch id[0] {
	case '"':
		var value string
		if err := json.Unmarshal(id, &value); err != nil {
			return "", false
		}
		return "s:" + value, true
	case 'n':
		if bytes.Equal(id, []byte("null")) {
			return "z:", true
		}
		return "", false
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if len(id) > 128 {
			return "", false
		}
		markerIndex := bytes.IndexAny(id, ".eE")
		if markerIndex < 0 {
			start := 0
			if id[0] == '-' {
				start = 1
			}
			if start == len(id) || id[start] == '0' && start+1 != len(id) {
				return "", false
			}
			value, err := strconv.ParseInt(string(id), 10, 64)
			if err != nil {
				return "", false
			}
			return "n:" + strconv.FormatInt(value, 10), true
		}
		if !json.Valid(id) {
			return "", false
		}
		exponentIndex := markerIndex
		if id[markerIndex] == '.' {
			exponentIndex = bytes.IndexAny(id[markerIndex+1:], "eE")
			if exponentIndex >= 0 {
				exponentIndex += markerIndex + 1
			}
		}
		if exponentIndex >= 0 {
			exponent, err := strconv.Atoi(string(id[exponentIndex+1:]))
			if err != nil || exponent < -1024 || exponent > 1024 {
				return "", false
			}
		}
		var value big.Rat
		if _, ok := value.SetString(string(id)); !ok || !value.IsInt() || !value.Num().IsInt64() {
			return "", false
		}
		return "n:" + value.Num().String(), true
	default:
		return "", false
	}
}

func validParams(params json.RawMessage) bool {
	params = bytes.TrimSpace(params)
	if len(params) == 0 || !json.Valid(params) {
		return false
	}
	return params[0] == '{' || params[0] == '['
}
