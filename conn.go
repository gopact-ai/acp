package acp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	defaultMaxFrameBytes         = 16 << 20
	defaultMaxBufferedBytes      = 64 << 20
	defaultMaxConcurrentRequests = 64
	defaultNotificationBacklog   = 1024
)

// A completed response has no remaining sender, so its channel can be reused.
// This removes two allocations and about 160 bytes from the successful call path.
var responseChannelPool = sync.Pool{New: func() any { return make(chan response, 1) }}

// Handler handles an incoming request or notification. Requests may be handled
// concurrently; notifications are handled serially in wire order. Omitted
// params are passed as an empty JSON object. Returning *Error controls a request
// error response; other errors become Internal Error responses.
type Handler func(context.Context, string, json.RawMessage) (any, error)

// ExtensionRequestHandler handles extension requests not defined by stable ACP v1.
type ExtensionRequestHandler interface {
	HandleRequest(context.Context, string, json.RawMessage) (any, error)
}

// ExtensionNotificationHandler handles extension notifications not defined by stable ACP v1.
type ExtensionNotificationHandler interface {
	HandleNotification(context.Context, string, json.RawMessage) error
}

// Option configures a Conn.
type Option func(*connConfig) error

type connConfig struct {
	logger                *slog.Logger
	maxFrameBytes         int
	maxBufferedBytes      int
	maxConcurrentRequests int
	notificationBacklog   int
}

// WithLogger enables structured runtime diagnostics. Protocol payloads are
// never logged.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *connConfig) error {
		if logger == nil {
			return errors.New("acp: logger is nil")
		}
		cfg.logger = logger
		return nil
	}
}

// WithMaxFrameBytes sets the maximum NDJSON frame size, including the trailing
// newline on writes.
func WithMaxFrameBytes(n int) Option {
	return func(cfg *connConfig) error {
		if n <= 0 {
			return errors.New("acp: max frame bytes must be positive")
		}
		cfg.maxFrameBytes = n
		return nil
	}
}

// WithMaxBufferedBytes bounds payloads retained by active inbound handlers and
// the notification queue. It must be at least the maximum frame size.
func WithMaxBufferedBytes(n int) Option {
	return func(cfg *connConfig) error {
		if n <= 0 {
			return errors.New("acp: max buffered bytes must be positive")
		}
		cfg.maxBufferedBytes = n
		return nil
	}
}

// WithNotificationBacklog sets the maximum number of queued notifications.
func WithNotificationBacklog(n int) Option {
	return func(cfg *connConfig) error {
		if n <= 0 {
			return errors.New("acp: notification backlog must be positive")
		}
		cfg.notificationBacklog = n
		return nil
	}
}

// WithMaxConcurrentRequests sets the maximum number of active inbound
// requests.
func WithMaxConcurrentRequests(n int) Option {
	return func(cfg *connConfig) error {
		if n <= 0 {
			return errors.New("acp: max concurrent requests must be positive")
		}
		cfg.maxConcurrentRequests = n
		return nil
	}
}

// Conn is a bidirectional newline-delimited JSON-RPC 2.0 connection. Its
// methods are safe for concurrent use.
type Conn struct {
	input   io.ReadCloser
	output  io.Writer
	handler Handler
	logger  *slog.Logger

	maxFrameBytes    int
	maxBufferedBytes int64
	notifications    chan notification
	requestSlots     chan struct{}

	ctx    context.Context
	cancel context.CancelCauseFunc
	done   chan struct{}

	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup
	writeGate chan struct{}

	nextID   atomic.Uint64
	buffered atomic.Int64
	mu       sync.Mutex
	pending  map[string]chan response
	inflight map[string]context.CancelCauseFunc
}

// New starts a generic bidirectional connection for ACP extensions. A nil
// handler accepts notifications and responds to requests with Method Not Found.
func New(input io.ReadCloser, output io.Writer, handler Handler, opts ...Option) (*Conn, error) {
	c, err := newConn(input, output, opts...)
	if err != nil {
		return nil, err
	}
	c.start(handler)
	return c, nil
}

func newConn(input io.ReadCloser, output io.Writer, opts ...Option) (*Conn, error) {
	if isNil(input) {
		return nil, errors.New("acp: input is nil")
	}
	if isNil(output) {
		return nil, errors.New("acp: output is nil")
	}

	cfg := connConfig{
		logger:                slog.New(slog.DiscardHandler),
		maxFrameBytes:         defaultMaxFrameBytes,
		maxBufferedBytes:      defaultMaxBufferedBytes,
		maxConcurrentRequests: defaultMaxConcurrentRequests,
		notificationBacklog:   defaultNotificationBacklog,
	}
	for _, opt := range opts {
		if opt == nil {
			return nil, errors.New("acp: option is nil")
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	if cfg.maxBufferedBytes < cfg.maxFrameBytes {
		return nil, errors.New("acp: max buffered bytes must be at least max frame bytes")
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	return &Conn{
		input:            input,
		output:           output,
		logger:           cfg.logger,
		maxFrameBytes:    cfg.maxFrameBytes,
		maxBufferedBytes: int64(cfg.maxBufferedBytes),
		notifications:    make(chan notification, cfg.notificationBacklog),
		requestSlots:     make(chan struct{}, cfg.maxConcurrentRequests),
		ctx:              ctx,
		cancel:           cancel,
		done:             make(chan struct{}),
		writeGate:        make(chan struct{}, 1),
		pending:          make(map[string]chan response),
		inflight:         make(map[string]context.CancelCauseFunc),
	}, nil
}

func (c *Conn) start(handler Handler) {
	c.startOnce.Do(func() {
		c.handler = handler
		c.wg.Add(2)
		go c.readLoop()
		go c.notificationLoop()
		go func() {
			c.wg.Wait()
			close(c.done)
		}()
	})
}

// Done is closed after the reader, notification dispatcher, and active request
// handlers stop.
func (c *Conn) Done() <-chan struct{} {
	return c.done
}

// Err returns the terminal connection error, or nil while the connection is
// running.
func (c *Conn) Err() error {
	select {
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	default:
		return nil
	}
}

// Close stops the connection and closes its input. It does not separately
// close output; closing an input that shares the same transport may affect it.
func (c *Conn) Close() error {
	return c.shutdown(ErrClosed)
}

// Call sends a request and decodes its result. Cancelling ctx makes a
// best-effort $/cancel_request notification to the peer.
func (c *Conn) Call(ctx context.Context, method string, params, result any) error {
	if ctx == nil {
		return errors.New("acp: context is nil")
	}
	if method == "" {
		return errors.New("acp: method is empty")
	}

	idNumber := c.nextID.Add(1)
	paramsJSON, err := marshalParams(params)
	if err != nil {
		return fmt.Errorf("acp: marshal params: %w", err)
	}
	return c.callWithID(ctx, idNumber, method, paramsJSON, result)
}

func (c *Conn) callMarshaled(ctx context.Context, method string, params json.RawMessage, result any) error {
	return c.callWithID(ctx, c.nextID.Add(1), method, params, result)
}

func (c *Conn) callWithID(ctx context.Context, idNumber uint64, method string, params json.RawMessage, result any) error {
	id := rawID(idNumber)
	key := "n:" + strconv.FormatUint(idNumber, 10)
	responses, ok := responseChannelPool.Get().(chan response)
	if !ok {
		responses = make(chan response, 1)
	}
	c.mu.Lock()
	c.pending[key] = responses
	c.mu.Unlock()

	if err := c.write(ctx, requestEnvelope{JSONRPC: "2.0", ID: id, Method: method, Params: params}); err != nil {
		c.removePending(key)
		return err
	}

	var resp response
	select {
	case resp = <-responses:
	default:
		select {
		case resp = <-responses:
		case <-ctx.Done():
			select {
			case resp = <-responses:
			default:
				c.removePending(key)
				c.tryCancel(id)
				return context.Cause(ctx)
			}
		case <-c.ctx.Done():
			select {
			case resp = <-responses:
			default:
				c.removePending(key)
				return context.Cause(c.ctx)
			}
		}
	}

	responseChannelPool.Put(responses)
	scope, _ := ctx.Value(notificationContextKey{}).(*notificationCallScope)
	isReentrantNotificationCall := scope != nil && scope.conn == c && scope.active.Load()
	if resp.barrier != nil && !isReentrantNotificationCall {
		select {
		case <-resp.barrier:
		default:
			select {
			case <-resp.barrier:
			case <-ctx.Done():
				select {
				case <-resp.barrier:
				default:
					return context.Cause(ctx)
				}
			case <-c.ctx.Done():
				select {
				case <-resp.barrier:
				default:
					return context.Cause(c.ctx)
				}
			}
		}
	}
	if resp.err != nil {
		return resp.err
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(resp.result, result); err != nil {
		return fmt.Errorf("acp: decode result: %w", err)
	}
	return nil
}

// Notify sends a notification.
func (c *Conn) Notify(ctx context.Context, method string, params any) error {
	if ctx == nil {
		return errors.New("acp: context is nil")
	}
	if method == "" {
		return errors.New("acp: method is empty")
	}
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	default:
	}

	paramsJSON, err := marshalParams(params)
	if err != nil {
		return fmt.Errorf("acp: marshal params: %w", err)
	}
	return c.notifyMarshaled(ctx, method, paramsJSON)
}

func (c *Conn) notifyMarshaled(ctx context.Context, method string, params json.RawMessage) error {
	return c.write(ctx, notificationEnvelope{JSONRPC: "2.0", Method: method, Params: params})
}

func (c *Conn) shutdown(cause error) error {
	var closeErr error
	c.stopOnce.Do(func() {
		if cause == nil {
			cause = ErrClosed
		}
		c.cancel(cause)
		closeErr = c.input.Close()
	})
	return closeErr
}

func (c *Conn) removePending(key string) {
	c.mu.Lock()
	delete(c.pending, key)
	c.mu.Unlock()
}

func (c *Conn) reserveBufferedBytes(n int) bool {
	bytes := int64(n)
	for {
		current := c.buffered.Load()
		if bytes > c.maxBufferedBytes-current {
			return false
		}
		if c.buffered.CompareAndSwap(current, current+bytes) {
			return true
		}
	}
}

func (c *Conn) releaseBufferedBytes(n int) {
	c.buffered.Add(-int64(n))
}

func (c *Conn) hasPending(id json.RawMessage) bool {
	key, ok := idKey(id)
	if !ok {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok = c.pending[key]
	return ok
}
