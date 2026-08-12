package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func (c *Conn) writeResponse(method string, id json.RawMessage, result any, rpcErr *Error) error {
	resp := responseEnvelope{JSONRPC: "2.0", ID: id, Error: rpcErr}
	if rpcErr != nil {
		if _, err := json.Marshal(rpcErr); err != nil {
			c.logger.Error("marshal handler error failed", "method", method, "error", err)
			resp.Error = internalError()
		}
		return c.writeResponseEnvelope(method, resp)
	}

	b, err := json.Marshal(result)
	if err != nil {
		c.logger.Error("marshal handler result failed", "method", method, "error", err)
		resp.Error = internalError()
		return c.writeResponseEnvelope(method, resp)
	}

	fields, isTypedResponse := responseRequiredFields[method]
	isEmptyResponse := isTypedResponse && len(fields) == 0
	if isEmptyResponse && bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		b = []byte("{}")
	}
	if err := validateResponseFields(b, method); err != nil {
		c.logger.Error("handler returned invalid result", "method", method, "error", err)
		resp.Error = internalError()
		return c.writeResponseEnvelope(method, resp)
	}
	if result != nil {
		if err := validateTypedJSON(b, result); err != nil {
			c.logger.Error("handler returned invalid result", "method", method, "error", err)
			resp.Error = internalError()
			return c.writeResponseEnvelope(method, resp)
		}
	}
	resp.Result = b
	return c.writeResponseEnvelope(method, resp)
}

func (c *Conn) writeResponseEnvelope(method string, resp responseEnvelope) error {
	err := c.write(c.ctx, resp)
	if !errors.Is(err, ErrFrameTooLarge) {
		return err
	}
	c.logger.Error("handler response exceeded frame limit", "method", method, "error", err)
	resp.Result = nil
	resp.Error = internalError()
	if fallbackErr := c.write(c.ctx, resp); fallbackErr != nil {
		_ = c.shutdown(fallbackErr)
		return fmt.Errorf("acp: write internal error response: %w", fallbackErr)
	}
	return nil
}

func (c *Conn) write(ctx context.Context, value any) error {
	return c.writeValue(ctx, value, true)
}

func (c *Conn) writeValue(ctx context.Context, value any, wait bool) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	default:
	}

	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("acp: encode message: %w", err)
	}
	if len(b)+1 > c.maxFrameBytes {
		return ErrFrameTooLarge
	}
	b = append(b, '\n')

	if wait {
		select {
		case c.writeGate <- struct{}{}:
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-c.ctx.Done():
			return context.Cause(c.ctx)
		}
	} else {
		select {
		case c.writeGate <- struct{}{}:
		default:
			return nil
		}
	}
	defer func() { <-c.writeGate }()

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	default:
	}
	for len(b) > 0 {
		n, writeErr := c.output.Write(b)
		if n < 0 || n > len(b) {
			writeErr = fmt.Errorf("acp: invalid write count %d for %d-byte buffer", n, len(b))
			_ = c.shutdown(writeErr)
			return writeErr
		}
		if writeErr != nil {
			_ = c.shutdown(writeErr)
			return fmt.Errorf("acp: write message: %w", writeErr)
		}
		if n == 0 {
			_ = c.shutdown(io.ErrShortWrite)
			return fmt.Errorf("acp: write message: %w", io.ErrShortWrite)
		}
		b = b[n:]
	}
	return nil
}

func (c *Conn) tryCancel(id json.RawMessage) {
	params, err := json.Marshal(CancelRequestNotification{RequestID: id})
	if err != nil {
		return
	}
	_ = c.writeValue(c.ctx, notificationEnvelope{JSONRPC: "2.0", Method: MethodCancelRequest, Params: params}, false)
}
