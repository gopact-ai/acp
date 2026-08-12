package acp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// AgentHandler defines the baseline methods every ACP agent implements.
type AgentHandler interface {
	Initialize(context.Context, *InitializeRequest) (*InitializeResponse, error)
	NewSession(context.Context, *NewSessionRequest) (*NewSessionResponse, error)
	Prompt(context.Context, *PromptRequest) (*PromptResponse, error)
	Cancel(context.Context, *CancelNotification) error
}

// AuthenticateHandler optionally handles authenticate requests.
type AuthenticateHandler interface {
	Authenticate(context.Context, *AuthenticateRequest) (*AuthenticateResponse, error)
}

// LogoutHandler optionally handles logout requests.
type LogoutHandler interface {
	Logout(context.Context, *LogoutRequest) (*LogoutResponse, error)
}

// LoadSessionHandler optionally loads an existing session.
type LoadSessionHandler interface {
	LoadSession(context.Context, *LoadSessionRequest) (*LoadSessionResponse, error)
}

// SetSessionModeHandler optionally changes a session mode.
type SetSessionModeHandler interface {
	SetSessionMode(context.Context, *SetSessionModeRequest) (*SetSessionModeResponse, error)
}

// SetSessionConfigOptionHandler optionally changes a session configuration.
type SetSessionConfigOptionHandler interface {
	SetSessionConfigOption(context.Context, *SetSessionConfigOptionRequest) (*SetSessionConfigOptionResponse, error)
}

// ListSessionsHandler optionally lists sessions.
type ListSessionsHandler interface {
	ListSessions(context.Context, *ListSessionsRequest) (*ListSessionsResponse, error)
}

// DeleteSessionHandler optionally deletes a session.
type DeleteSessionHandler interface {
	DeleteSession(context.Context, *DeleteSessionRequest) (*DeleteSessionResponse, error)
}

// ResumeSessionHandler optionally resumes a session.
type ResumeSessionHandler interface {
	ResumeSession(context.Context, *ResumeSessionRequest) (*ResumeSessionResponse, error)
}

// CloseSessionHandler optionally closes a session.
type CloseSessionHandler interface {
	CloseSession(context.Context, *CloseSessionRequest) (*CloseSessionResponse, error)
}

// ClientCaller invokes client-side ACP methods from an agent handler. Values
// are supplied by NewAgent; the zero value is not usable.
type ClientCaller struct {
	conn *Conn
}

// NewAgent starts the agent side of an ACP connection. The factory receives the
// caller used for reverse client requests and notifications. It must return
// without performing protocol I/O; the connection starts after it returns.
func NewAgent(input io.ReadCloser, output io.Writer, factory func(*ClientCaller) AgentHandler, opts ...Option) (*Conn, error) {
	if factory == nil {
		return nil, errors.New("acp: agent factory is nil")
	}
	c, err := newConn(input, output, opts...)
	if err != nil {
		return nil, err
	}
	agent := factory(&ClientCaller{conn: c})
	if isNil(agent) {
		_ = c.Close()
		return nil, errors.New("acp: agent handler is nil")
	}
	c.start(dispatchAgent(agent))
	return c, nil
}

func dispatchAgent(agent AgentHandler) Handler {
	return func(ctx context.Context, method string, params json.RawMessage) (any, error) {
		switch method {
		case MethodInitialize:
			return invoke(ctx, params, method, agent.Initialize)
		case MethodAuthenticate:
			return invokeOptional(ctx, params, method, agent, AuthenticateHandler.Authenticate)
		case MethodLogout:
			return invokeOptional(ctx, params, method, agent, LogoutHandler.Logout)
		case MethodSessionNew:
			return invoke(ctx, params, method, agent.NewSession)
		case MethodSessionLoad:
			return invokeOptional(ctx, params, method, agent, LoadSessionHandler.LoadSession)
		case MethodSessionSetMode:
			return invokeOptional(ctx, params, method, agent, SetSessionModeHandler.SetSessionMode)
		case MethodSessionSetConfigOption:
			return invokeOptional(ctx, params, method, agent, SetSessionConfigOptionHandler.SetSessionConfigOption)
		case MethodSessionPrompt:
			return invoke(ctx, params, method, agent.Prompt)
		case MethodSessionCancel:
			if err := validateRequestFields(params, method); err != nil {
				return nil, invalidParams(err)
			}
			var req CancelNotification
			if err := json.Unmarshal(params, &req); err != nil {
				return nil, invalidParams(err)
			}
			return nil, agent.Cancel(ctx, &req)
		case MethodSessionList:
			return invokeOptional(ctx, params, method, agent, ListSessionsHandler.ListSessions)
		case MethodSessionDelete:
			return invokeOptional(ctx, params, method, agent, DeleteSessionHandler.DeleteSession)
		case MethodSessionResume:
			return invokeOptional(ctx, params, method, agent, ResumeSessionHandler.ResumeSession)
		case MethodSessionClose:
			return invokeOptional(ctx, params, method, agent, CloseSessionHandler.CloseSession)
		default:
			return dispatchExtension(ctx, method, params, agent)
		}
	}
}

// Call invokes an extension request on the client.
func (c *ClientCaller) Call(ctx context.Context, method string, params, result any) error {
	return c.conn.Call(ctx, method, params, result)
}

// Notify sends an extension notification to the client.
func (c *ClientCaller) Notify(ctx context.Context, method string, params any) error {
	return c.conn.Notify(ctx, method, params)
}

// RequestPermission asks the client to approve a tool call.
func (c *ClientCaller) RequestPermission(ctx context.Context, req *RequestPermissionRequest) (*RequestPermissionResponse, error) {
	return call[RequestPermissionResponse](ctx, c.conn, MethodSessionRequestPermission, req)
}

// Update sends a session update notification to the client.
func (c *ClientCaller) Update(ctx context.Context, notification *SessionNotification) error {
	return notify(ctx, c.conn, MethodSessionUpdate, notification)
}

// WriteTextFile asks the client to write a text file.
func (c *ClientCaller) WriteTextFile(ctx context.Context, req *WriteTextFileRequest) (*WriteTextFileResponse, error) {
	return call[WriteTextFileResponse](ctx, c.conn, MethodFSWriteTextFile, req)
}

// ReadTextFile asks the client to read a text file.
func (c *ClientCaller) ReadTextFile(ctx context.Context, req *ReadTextFileRequest) (*ReadTextFileResponse, error) {
	return call[ReadTextFileResponse](ctx, c.conn, MethodFSReadTextFile, req)
}

// CreateTerminal asks the client to create a terminal.
func (c *ClientCaller) CreateTerminal(ctx context.Context, req *CreateTerminalRequest) (*CreateTerminalResponse, error) {
	return call[CreateTerminalResponse](ctx, c.conn, MethodTerminalCreate, req)
}

// TerminalOutput reads terminal output from the client.
func (c *ClientCaller) TerminalOutput(ctx context.Context, req *TerminalOutputRequest) (*TerminalOutputResponse, error) {
	return call[TerminalOutputResponse](ctx, c.conn, MethodTerminalOutput, req)
}

// ReleaseTerminal releases a terminal owned by the client.
func (c *ClientCaller) ReleaseTerminal(ctx context.Context, req *ReleaseTerminalRequest) (*ReleaseTerminalResponse, error) {
	return call[ReleaseTerminalResponse](ctx, c.conn, MethodTerminalRelease, req)
}

// WaitForTerminalExit waits for a client terminal to exit.
func (c *ClientCaller) WaitForTerminalExit(ctx context.Context, req *WaitForTerminalExitRequest) (*WaitForTerminalExitResponse, error) {
	return call[WaitForTerminalExitResponse](ctx, c.conn, MethodTerminalWaitForExit, req)
}

// KillTerminal asks the client to kill a terminal command.
func (c *ClientCaller) KillTerminal(ctx context.Context, req *KillTerminalRequest) (*KillTerminalResponse, error) {
	return call[KillTerminalResponse](ctx, c.conn, MethodTerminalKill, req)
}

// CreateElicitation asks the client to collect structured input from the user.
func (c *ClientCaller) CreateElicitation(ctx context.Context, req *CreateElicitationRequest) (*CreateElicitationResponse, error) {
	return call[CreateElicitationResponse](ctx, c.conn, MethodElicitationCreate, req)
}

// CompleteElicitation tells the client that a URL elicitation has completed.
func (c *ClientCaller) CompleteElicitation(ctx context.Context, notification *CompleteElicitationNotification) error {
	return notify(ctx, c.conn, MethodElicitationComplete, notification)
}

func call[T any](ctx context.Context, conn *Conn, method string, params any) (*T, error) {
	if ctx == nil {
		return nil, errors.New("acp: context is nil")
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("acp: marshal %s params: %w", method, err)
	}
	if err := validateRequestFields(paramsJSON, method); err != nil {
		return nil, fmt.Errorf("acp: invalid %s params: %w", method, err)
	}
	if err := validateTypedJSON(paramsJSON, params); err != nil {
		return nil, fmt.Errorf("acp: invalid %s params: %w", method, err)
	}
	var raw json.RawMessage
	if err := conn.callMarshaled(ctx, method, paramsJSON, &raw); err != nil {
		return nil, err
	}
	if err := validateResponseFields(raw, method); err != nil {
		responseErr := fmt.Errorf("%w: %w", ErrInvalidResponse, err)
		_ = conn.shutdown(responseErr)
		return nil, responseErr
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		responseErr := fmt.Errorf("%w: decode result: %w", ErrInvalidResponse, err)
		_ = conn.shutdown(responseErr)
		return nil, responseErr
	}
	return &result, nil
}

func notify(ctx context.Context, conn *Conn, method string, params any) error {
	if ctx == nil {
		return errors.New("acp: context is nil")
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("acp: marshal %s params: %w", method, err)
	}
	if err := validateRequestFields(paramsJSON, method); err != nil {
		return fmt.Errorf("acp: invalid %s params: %w", method, err)
	}
	if err := validateTypedJSON(paramsJSON, params); err != nil {
		return fmt.Errorf("acp: invalid %s params: %w", method, err)
	}
	return conn.notifyMarshaled(ctx, method, paramsJSON)
}

func invoke[Req, Resp any](ctx context.Context, params json.RawMessage, method string, fn func(context.Context, *Req) (*Resp, error)) (any, error) {
	if err := validateRequestFields(params, method); err != nil {
		return nil, invalidParams(err)
	}
	var req Req
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, invalidParams(err)
	}
	return fn(ctx, &req)
}

func invokeOptional[H, Req, Resp any](ctx context.Context, params json.RawMessage, method string, handler any, fn func(H, context.Context, *Req) (*Resp, error)) (any, error) {
	h, ok := handler.(H)
	if !ok {
		return nil, methodNotFound(method)
	}
	if err := validateRequestFields(params, method); err != nil {
		return nil, invalidParams(err)
	}
	var req Req
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, invalidParams(err)
	}
	return fn(h, ctx, &req)
}
