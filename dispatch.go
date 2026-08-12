package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
)

type notification struct {
	method string
	params json.RawMessage
	done   chan struct{}
	bytes  int
}

type notificationContextKey struct{}

type notificationCallScope struct {
	conn   *Conn
	active atomic.Bool
}

type response struct {
	result  json.RawMessage
	err     *Error
	barrier <-chan struct{}
}

func dispatchExtension(ctx context.Context, method string, params json.RawMessage, handler any) (any, error) {
	if _, isNotification := ctx.Value(notificationContextKey{}).(*notificationCallScope); isNotification {
		extension, ok := handler.(ExtensionNotificationHandler)
		if !ok {
			return nil, methodNotFound(method)
		}
		return nil, extension.HandleNotification(ctx, method, params)
	}
	extension, ok := handler.(ExtensionRequestHandler)
	if !ok {
		return nil, methodNotFound(method)
	}
	return extension.HandleRequest(ctx, method, params)
}

func (c *Conn) readLoop() {
	defer c.wg.Done()
	defer close(c.notifications)

	reader := bufio.NewReader(c.input)
	closedBarrier := make(chan struct{})
	close(closedBarrier)
	var lastBarrier <-chan struct{} = closedBarrier

	for {
		frame, err := readFrame(reader, c.maxFrameBytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				_ = c.shutdown(io.EOF)
			} else {
				_ = c.shutdown(err)
			}
			return
		}

		msg, rpcErr := decodeMessage(frame)
		if rpcErr != nil {
			if msg.kind == messageResponse || msg.kind == 0 && c.hasPending(msg.id) {
				_ = c.shutdown(fmt.Errorf("%w: %s", ErrInvalidResponse, rpcErr.Message))
				return
			}
			id := msg.id
			if !validID(id) {
				id = json.RawMessage("null")
			}
			if err := c.write(c.ctx, responseEnvelope{JSONRPC: "2.0", ID: id, Error: rpcErr}); err != nil {
				_ = c.shutdown(err)
				return
			}
			continue
		}

		switch msg.kind {
		case messageResponse:
			c.handleResponse(msg, lastBarrier)
		case messageNotification:
			if msg.method == MethodCancelRequest {
				c.handleCancelRequest(msg.params)
				continue
			}
			if !c.reserveBufferedBytes(len(frame)) {
				_ = c.shutdown(ErrBufferFull)
				return
			}
			item := notification{
				method: msg.method,
				params: msg.params,
				done:   make(chan struct{}),
				bytes:  len(frame),
			}
			select {
			case c.notifications <- item:
				lastBarrier = item.done
			default:
				c.releaseBufferedBytes(item.bytes)
				_ = c.shutdown(ErrQueueFull)
				return
			}
		case messageRequest:
			if !c.reserveBufferedBytes(len(frame)) {
				_ = c.shutdown(ErrBufferFull)
				return
			}
			if !c.handleRequest(msg, len(frame)) {
				_ = c.shutdown(ErrTooManyRequests)
				return
			}
		}
	}
}

func (c *Conn) notificationLoop() {
	defer c.wg.Done()
	for item := range c.notifications {
		func() {
			defer close(item.done)
			defer c.releaseBufferedBytes(item.bytes)
			scope := &notificationCallScope{conn: c}
			scope.active.Store(true)
			defer scope.active.Store(false)
			defer func() {
				if recovered := recover(); recovered != nil {
					c.logger.Error("notification handler panicked", "method", item.method, "panic", recovered)
				}
			}()
			if c.handler == nil {
				return
			}
			handlerCtx := context.WithValue(c.ctx, notificationContextKey{}, scope)
			if _, err := c.handler(handlerCtx, item.method, paramsOrEmpty(item.params)); err != nil {
				c.logger.Error("notification handler failed", "method", item.method, "error", err)
			}
		}()
	}
}

func (c *Conn) handleRequest(msg message, payloadBytes int) bool {
	select {
	case c.requestSlots <- struct{}{}:
	default:
		c.releaseBufferedBytes(payloadBytes)
		return false
	}

	key, _ := idKey(msg.id)
	reqCtx, cancel := context.WithCancelCause(c.ctx)

	c.mu.Lock()
	if _, exists := c.inflight[key]; exists {
		c.mu.Unlock()
		cancel(nil)
		<-c.requestSlots
		c.releaseBufferedBytes(payloadBytes)
		c.logger.Debug("ignored request with duplicate id")
		return true
	}
	c.inflight[key] = cancel
	c.mu.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() { <-c.requestSlots }()
		defer c.releaseBufferedBytes(payloadBytes)
		defer func() {
			c.mu.Lock()
			delete(c.inflight, key)
			c.mu.Unlock()
			cancel(nil)
		}()

		var result any
		var callErr error
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					c.logger.Error("request handler panicked", "method", msg.method, "panic", recovered)
					callErr = internalError()
				}
			}()
			if c.handler == nil {
				callErr = methodNotFound(msg.method)
				return
			}
			result, callErr = c.handler(reqCtx, msg.method, paramsOrEmpty(msg.params))
		}()

		var rpcErr *Error
		isCancellation := errors.Is(callErr, context.Canceled) || errors.Is(callErr, context.DeadlineExceeded)
		if callErr != nil && !errors.As(callErr, &rpcErr) && !isCancellation {
			c.logger.Error("request handler failed", "method", msg.method, "error", callErr)
		}
		if err := c.writeResponse(msg.method, msg.id, result, asRPCError(callErr)); err != nil &&
			!errors.Is(err, ErrClosed) {
			c.logger.Error("write response failed", "method", msg.method, "error", err)
		}
	}()
	return true
}

func (c *Conn) handleResponse(msg message, barrier <-chan struct{}) {
	key, _ := idKey(msg.id)
	c.mu.Lock()
	responses := c.pending[key]
	delete(c.pending, key)
	c.mu.Unlock()
	if responses == nil {
		c.logger.Debug("ignored response for unknown request")
		return
	}
	responses <- response{result: msg.result, err: msg.err, barrier: barrier}
}

func (c *Conn) handleCancelRequest(params json.RawMessage) {
	var notification CancelRequestNotification
	if err := json.Unmarshal(paramsOrEmpty(params), &notification); err != nil {
		c.logger.Debug("ignored invalid cancel request", "error", err)
		return
	}
	if !validID(notification.RequestID) {
		c.logger.Debug("ignored cancel request with invalid id")
		return
	}

	key, _ := idKey(notification.RequestID)
	c.mu.Lock()
	cancel := c.inflight[key]
	c.mu.Unlock()
	if cancel != nil {
		cancel(context.Canceled)
	}
}
