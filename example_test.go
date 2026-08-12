package acp_test

import (
	"context"
	"fmt"
	"net"

	"github.com/gopact-ai/acp"
)

type exampleAgent struct{}

func (*exampleAgent) Initialize(context.Context, *acp.InitializeRequest) (*acp.InitializeResponse, error) {
	return &acp.InitializeResponse{ProtocolVersion: acp.ProtocolVersionV1}, nil
}

func (*exampleAgent) NewSession(context.Context, *acp.NewSessionRequest) (*acp.NewSessionResponse, error) {
	return &acp.NewSessionResponse{SessionID: "session-1"}, nil
}

func (*exampleAgent) Prompt(context.Context, *acp.PromptRequest) (*acp.PromptResponse, error) {
	return &acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
}

func (*exampleAgent) Cancel(context.Context, *acp.CancelNotification) error { return nil }

type exampleClient struct{}

func (*exampleClient) RequestPermission(context.Context, *acp.RequestPermissionRequest) (*acp.RequestPermissionResponse, error) {
	return &acp.RequestPermissionResponse{Outcome: acp.CanceledRequestPermissionOutcome()}, nil
}

func (*exampleClient) Update(context.Context, *acp.SessionNotification) error { return nil }

func ExampleNewAgent() {
	agentTransport, clientTransport := net.Pipe()
	agentConn, err := acp.NewAgent(agentTransport, agentTransport, func(*acp.ClientCaller) acp.AgentHandler { return &exampleAgent{} })
	if err != nil {
		panic(err)
	}
	defer func() { _ = agentConn.Close() }()

	var agent *acp.AgentCaller
	clientConn, err := acp.NewClient(clientTransport, clientTransport, func(caller *acp.AgentCaller) acp.ClientHandler {
		agent = caller
		return &exampleClient{}
	})
	if err != nil {
		panic(err)
	}
	defer func() { _ = clientConn.Close() }()

	response, err := agent.Initialize(context.Background(), &acp.InitializeRequest{ProtocolVersion: acp.ProtocolVersionV1})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.ProtocolVersion)
	// Output: 1
}
