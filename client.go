package acp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
)

// ClientHandler defines the baseline methods every ACP client implements.
type ClientHandler interface {
	RequestPermission(context.Context, *RequestPermissionRequest) (*RequestPermissionResponse, error)
	Update(context.Context, *SessionNotification) error
}

// WriteTextFileHandler optionally writes text files for an agent.
type WriteTextFileHandler interface {
	WriteTextFile(context.Context, *WriteTextFileRequest) (*WriteTextFileResponse, error)
}

// ReadTextFileHandler optionally reads text files for an agent.
type ReadTextFileHandler interface {
	ReadTextFile(context.Context, *ReadTextFileRequest) (*ReadTextFileResponse, error)
}

// TerminalHandler implements the complete ACP terminal capability.
type TerminalHandler interface {
	CreateTerminal(context.Context, *CreateTerminalRequest) (*CreateTerminalResponse, error)
	TerminalOutput(context.Context, *TerminalOutputRequest) (*TerminalOutputResponse, error)
	ReleaseTerminal(context.Context, *ReleaseTerminalRequest) (*ReleaseTerminalResponse, error)
	WaitForTerminalExit(context.Context, *WaitForTerminalExitRequest) (*WaitForTerminalExitResponse, error)
	KillTerminal(context.Context, *KillTerminalRequest) (*KillTerminalResponse, error)
}

// CreateElicitationHandler optionally collects structured input from the user.
type CreateElicitationHandler interface {
	CreateElicitation(context.Context, *CreateElicitationRequest) (*CreateElicitationResponse, error)
}

// CompleteElicitationHandler optionally handles completion of a URL elicitation.
type CompleteElicitationHandler interface {
	CompleteElicitation(context.Context, *CompleteElicitationNotification) error
}

// AgentCaller invokes agent-side ACP methods from a client handler. Values are
// supplied by NewClient; the zero value is not usable.
type AgentCaller struct {
	conn *Conn
}

// NewClient starts the client side of an ACP connection. The factory receives
// the caller used for agent requests and notifications. It must return without
// performing protocol I/O; the connection starts after it returns.
func NewClient(input io.ReadCloser, output io.Writer, factory func(*AgentCaller) ClientHandler, opts ...Option) (*Conn, error) {
	if factory == nil {
		return nil, errors.New("acp: client factory is nil")
	}
	c, err := newConn(input, output, opts...)
	if err != nil {
		return nil, err
	}
	client := factory(&AgentCaller{conn: c})
	if isNil(client) {
		_ = c.Close()
		return nil, errors.New("acp: client handler is nil")
	}
	c.start(dispatchClient(client))
	return c, nil
}

func dispatchClient(client ClientHandler) Handler {
	return func(ctx context.Context, method string, params json.RawMessage) (any, error) {
		switch method {
		case MethodSessionRequestPermission:
			return invoke(ctx, params, method, client.RequestPermission)
		case MethodSessionUpdate:
			if err := validateRequestFields(params, method); err != nil {
				return nil, invalidParams(err)
			}
			var notification SessionNotification
			if err := json.Unmarshal(params, &notification); err != nil {
				return nil, invalidParams(err)
			}
			return nil, client.Update(ctx, &notification)
		case MethodFSWriteTextFile:
			return invokeOptional(ctx, params, method, client, WriteTextFileHandler.WriteTextFile)
		case MethodFSReadTextFile:
			return invokeOptional(ctx, params, method, client, ReadTextFileHandler.ReadTextFile)
		case MethodTerminalCreate:
			return invokeOptional(ctx, params, method, client, TerminalHandler.CreateTerminal)
		case MethodTerminalOutput:
			return invokeOptional(ctx, params, method, client, TerminalHandler.TerminalOutput)
		case MethodTerminalRelease:
			return invokeOptional(ctx, params, method, client, TerminalHandler.ReleaseTerminal)
		case MethodTerminalWaitForExit:
			return invokeOptional(ctx, params, method, client, TerminalHandler.WaitForTerminalExit)
		case MethodTerminalKill:
			return invokeOptional(ctx, params, method, client, TerminalHandler.KillTerminal)
		case MethodElicitationCreate:
			return invokeOptional(ctx, params, method, client, CreateElicitationHandler.CreateElicitation)
		case MethodElicitationComplete:
			handler, ok := client.(CompleteElicitationHandler)
			if !ok {
				return nil, methodNotFound(method)
			}
			if err := validateRequestFields(params, method); err != nil {
				return nil, invalidParams(err)
			}
			var notification CompleteElicitationNotification
			if err := json.Unmarshal(params, &notification); err != nil {
				return nil, invalidParams(err)
			}
			return nil, handler.CompleteElicitation(ctx, &notification)
		default:
			return dispatchExtension(ctx, method, params, client)
		}
	}
}

// Call invokes an extension request on the agent.
func (a *AgentCaller) Call(ctx context.Context, method string, params, result any) error {
	return a.conn.Call(ctx, method, params, result)
}

// Notify sends an extension notification to the agent.
func (a *AgentCaller) Notify(ctx context.Context, method string, params any) error {
	return a.conn.Notify(ctx, method, params)
}

// Initialize negotiates the ACP protocol version and capabilities.
func (a *AgentCaller) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	return call[InitializeResponse](ctx, a.conn, MethodInitialize, req)
}

// Authenticate completes an advertised authentication method.
func (a *AgentCaller) Authenticate(ctx context.Context, req *AuthenticateRequest) (*AuthenticateResponse, error) {
	return call[AuthenticateResponse](ctx, a.conn, MethodAuthenticate, req)
}

// Logout clears agent authentication state.
func (a *AgentCaller) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	return call[LogoutResponse](ctx, a.conn, MethodLogout, req)
}

// NewSession creates a session.
func (a *AgentCaller) NewSession(ctx context.Context, req *NewSessionRequest) (*NewSessionResponse, error) {
	return call[NewSessionResponse](ctx, a.conn, MethodSessionNew, req)
}

// LoadSession loads a session.
func (a *AgentCaller) LoadSession(ctx context.Context, req *LoadSessionRequest) (*LoadSessionResponse, error) {
	return call[LoadSessionResponse](ctx, a.conn, MethodSessionLoad, req)
}

// SetSessionMode changes the active session mode.
func (a *AgentCaller) SetSessionMode(ctx context.Context, req *SetSessionModeRequest) (*SetSessionModeResponse, error) {
	return call[SetSessionModeResponse](ctx, a.conn, MethodSessionSetMode, req)
}

// SetSessionConfigOption changes a session configuration value.
func (a *AgentCaller) SetSessionConfigOption(ctx context.Context, req *SetSessionConfigOptionRequest) (*SetSessionConfigOptionResponse, error) {
	return call[SetSessionConfigOptionResponse](ctx, a.conn, MethodSessionSetConfigOption, req)
}

// Prompt starts a prompt turn.
func (a *AgentCaller) Prompt(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	return call[PromptResponse](ctx, a.conn, MethodSessionPrompt, req)
}

// Cancel cancels the active prompt turn for a session.
func (a *AgentCaller) Cancel(ctx context.Context, notification *CancelNotification) error {
	return notify(ctx, a.conn, MethodSessionCancel, notification)
}

// ListSessions lists saved sessions.
func (a *AgentCaller) ListSessions(ctx context.Context, req *ListSessionsRequest) (*ListSessionsResponse, error) {
	return call[ListSessionsResponse](ctx, a.conn, MethodSessionList, req)
}

// DeleteSession deletes a saved session.
func (a *AgentCaller) DeleteSession(ctx context.Context, req *DeleteSessionRequest) (*DeleteSessionResponse, error) {
	return call[DeleteSessionResponse](ctx, a.conn, MethodSessionDelete, req)
}

// ResumeSession resumes an existing session.
func (a *AgentCaller) ResumeSession(ctx context.Context, req *ResumeSessionRequest) (*ResumeSessionResponse, error) {
	return call[ResumeSessionResponse](ctx, a.conn, MethodSessionResume, req)
}

// CloseSession closes an active session.
func (a *AgentCaller) CloseSession(ctx context.Context, req *CloseSessionRequest) (*CloseSessionResponse, error) {
	return call[CloseSessionResponse](ctx, a.conn, MethodSessionClose, req)
}
