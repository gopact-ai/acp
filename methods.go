package acp

// ProtocolVersionV1 is the stable ACP wire protocol version implemented by
// this package.
const ProtocolVersionV1 ProtocolVersion = 1

// CancelRequestNotification carries the JSON-RPC request ID to cancel.
type CancelRequestNotification struct {
	Meta      Meta      `json:"_meta,omitzero"`
	RequestID RequestID `json:"requestId"`
}

const (
	// MethodInitialize names the protocol negotiation request.
	MethodInitialize = "initialize"
	// MethodAuthenticate names the agent authentication request.
	MethodAuthenticate = "authenticate"
	// MethodLogout names the agent logout request.
	MethodLogout = "logout"
	// MethodSessionNew names the session creation request.
	MethodSessionNew = "session/new"
	// MethodSessionLoad names the session load request.
	MethodSessionLoad = "session/load"
	// MethodSessionSetMode names the session mode change request.
	MethodSessionSetMode = "session/set_mode"
	// MethodSessionSetConfigOption names the session configuration request.
	MethodSessionSetConfigOption = "session/set_config_option"
	// MethodSessionPrompt names the prompt request.
	MethodSessionPrompt = "session/prompt"
	// MethodSessionCancel names the prompt cancellation notification.
	MethodSessionCancel = "session/cancel"
	// MethodSessionList names the session listing request.
	MethodSessionList = "session/list"
	// MethodSessionDelete names the session deletion request.
	MethodSessionDelete = "session/delete"
	// MethodSessionResume names the session resume request.
	MethodSessionResume = "session/resume"
	// MethodSessionClose names the session close request.
	MethodSessionClose = "session/close"

	// MethodSessionRequestPermission names the client permission request.
	MethodSessionRequestPermission = "session/request_permission"
	// MethodSessionUpdate names the client session update notification.
	MethodSessionUpdate = "session/update"
	// MethodFSWriteTextFile names the client file-write request.
	MethodFSWriteTextFile = "fs/write_text_file"
	// MethodFSReadTextFile names the client file-read request.
	MethodFSReadTextFile = "fs/read_text_file"
	// MethodTerminalCreate names the client terminal creation request.
	MethodTerminalCreate = "terminal/create"
	// MethodTerminalOutput names the client terminal output request.
	MethodTerminalOutput = "terminal/output"
	// MethodTerminalRelease names the client terminal release request.
	MethodTerminalRelease = "terminal/release"
	// MethodTerminalWaitForExit names the client terminal wait request.
	MethodTerminalWaitForExit = "terminal/wait_for_exit"
	// MethodTerminalKill names the client terminal kill request.
	MethodTerminalKill = "terminal/kill"
	// MethodElicitationCreate names the client elicitation request.
	MethodElicitationCreate = "elicitation/create"
	// MethodElicitationComplete names the client elicitation completion notification.
	MethodElicitationComplete = "elicitation/complete"

	// MethodCancelRequest is the protocol-level cancellation notification.
	MethodCancelRequest = "$/cancel_request"
)
