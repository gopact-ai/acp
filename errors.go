package acp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var (
	// ErrClosed indicates an explicitly closed connection.
	ErrClosed = errors.New("acp: connection closed")
	// ErrBufferFull indicates that retained inbound payloads reached their byte limit.
	ErrBufferFull = errors.New("acp: buffered message limit exceeded")
	// ErrFrameTooLarge indicates an inbound or outbound frame over its size limit.
	ErrFrameTooLarge = errors.New("acp: frame too large")
	// ErrInvalidResponse indicates a malformed or schema-invalid peer response.
	ErrInvalidResponse = errors.New("acp: invalid response")
	// ErrQueueFull indicates that the ordered notification backlog is full.
	ErrQueueFull = errors.New("acp: notification queue full")
	// ErrTooManyRequests indicates that the inbound request concurrency limit was reached.
	ErrTooManyRequests = errors.New("acp: too many concurrent requests")
)

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("acp: rpc error %d: %s", e.Code, e.Message)
}

func newRPCError(code ErrorCode, message string, data any) *Error {
	return &Error{Code: code, Message: message, Data: data}
}

func parseError(err error) *Error {
	return newRPCError(ErrorCodeParseError, "Parse error", errorData(err))
}

func invalidRequest(err error) *Error {
	return newRPCError(ErrorCodeInvalidRequest, "Invalid request", errorData(err))
}

func methodNotFound(method string) *Error {
	return newRPCError(ErrorCodeMethodNotFound, "Method not found", map[string]any{"method": method})
}

func invalidParams(err error) *Error {
	return newRPCError(ErrorCodeInvalidParams, "Invalid params", errorData(err))
}

func internalError() *Error {
	return newRPCError(ErrorCodeInternalError, "Internal error", nil)
}

func requestCanceled() *Error {
	return newRPCError(ErrorCodeRequestCanceled, "Request cancelled", nil)
}

func errorData(err error) any {
	if err == nil {
		return nil
	}
	return map[string]any{"error": err.Error()}
}

func asRPCError(err error) *Error {
	if err == nil {
		return nil
	}
	var rpcErr *Error
	if errors.As(err, &rpcErr) && rpcErr != nil {
		return rpcErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrClosed) {
		return requestCanceled()
	}
	return internalError()
}

func rawID(v uint64) json.RawMessage {
	return strconv.AppendUint(nil, v, 10)
}
