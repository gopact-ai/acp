# acp

[![CI](https://github.com/gopact-ai/acp/actions/workflows/ci.yml/badge.svg)](https://github.com/gopact-ai/acp/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopact-ai/acp.svg)](https://pkg.go.dev/github.com/gopact-ai/acp)
[![License](https://img.shields.io/github/license/gopact-ai/acp)](LICENSE)

`acp` implements stable Agent Client Protocol v1 for Go. It provides typed
client and agent APIs on top of a bidirectional newline-delimited JSON-RPC 2.0
runtime. The typed surface covers all 25 methods and notifications in the
pinned official stable-v1 schema.

The package tracks the stable ACP v1 schema pinned from the protocol main
branch and wire protocol version `1`. It requires Go 1.24. Its only runtime
dependency is `github.com/valyala/fastjson`.

## Install

```sh
go get github.com/gopact-ai/acp
```

## Quick start

Start the agent side over stdio:

```go
conn, err := acp.NewAgent(os.Stdin, os.Stdout, func(client *acp.ClientCaller) acp.AgentHandler {
	return newAgent(client)
})
if err != nil {
	return err
}
defer conn.Close()

<-conn.Done()
if err := conn.Err(); err != nil && !errors.Is(err, io.EOF) {
	return err
}
```

Use `NewClient` for the client side. A complete in-memory agent/client round
trip is available in [`example_test.go`](example_test.go).

The handler factories receive the reverse-direction caller and must return
without performing protocol I/O; the connection starts after the factory
returns. `AgentHandler` and `ClientHandler` contain the baseline methods.
Optional ACP capabilities are enabled by implementing the corresponding small
handler interfaces.

`AgentCaller` and `ClientCaller` expose `Call`/`Notify` for outbound ACP
extensions. Typed handlers can implement `ExtensionRequestHandler` and
`ExtensionNotificationHandler` for inbound extensions. `New` exposes the same
runtime without the stable-v1 typed layer.

## Runtime contract

- `Conn` calls `Close` on its input and never closes the output independently.
  If both sides share one transport, closing the input may close that transport.
- Inbound requests run concurrently, with a default maximum of 64.
  Notifications are processed in wire order. For ordinary calls, notifications
  received before a response finish before the corresponding call returns.
- Notification handlers may call the peer synchronously when they propagate
  the handler context. Reentrant calls bypass the notification barrier to avoid
  an ordering cycle; queued notifications then continue in wire order.
- Frames are limited to 16 MiB, retained request and notification payloads to
  64 MiB in total, and the ordered notification backlog to 1024. `Option`
  values configure these limits. Exceeding an inbound limit closes the
  connection with a sentinel error.
- Cancelling a call makes a best-effort `$/cancel_request` notification.
  Inbound cancellation reaches the request context. Handlers must observe their
  context and return; Go cannot forcibly stop a handler.
- Writes are serialized. Context cancellation can interrupt waiting for the
  writer, but it cannot interrupt an `io.Writer.Write` already in progress.
  Transports must not block forever.
- Protocol payloads are never logged. Logging is disabled by default and can be
  enabled with `WithLogger`.

ACP requires the client to call `initialize` and finish version and capability
negotiation before using sessions. The package does not duplicate that protocol
lifecycle as private transport state; callers and handlers enforce it.

## Schema provenance

The checked-in schema comes from the official stable v1 schema at protocol
commit [`af41b25f57a79c5629b3164e23fb4e8650badeeb`](https://github.com/agentclientprotocol/agent-client-protocol/tree/af41b25f57a79c5629b3164e23fb4e8650badeeb):

- `schema/v1/schema.json`
- `schema/v1/meta.json`

That commit includes stable elicitation, which was newer than the latest
`schema-v1.20.0` tag when this snapshot was taken. `types_gen.go` was generated with
`github.com/spachava753/acp-sdk/internal/schemagen` at commit
`ea76600dde1bd490a2fc6c0c4a44f05383a8abc9`, then corrected for stable-v1
required fields, default-on-error semantics, idiomatic Go names, and concrete
union decoding. Schema consistency and union behavior are covered by tests; do
not replace the file with unmodified upstream output.

The schema's Apache-2.0 license is preserved in
`third_party/agent-client-protocol.LICENSE`. Required upstream Apache-2.0, MIT,
and CC-BY-4.0 notices are preserved in `third_party/acp-sdk.LICENSE`; those
notices do not change this project's Apache-2.0 license.

## Contributing and security

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for development instructions. Report
vulnerabilities according to [`SECURITY.md`](SECURITY.md).

## License

Licensed under the [Apache License 2.0](LICENSE).
