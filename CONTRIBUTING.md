# Contributing

Contributions should keep the public API small, preserve ACP v1 wire semantics,
and include tests for observable behavior.

## Development setup

Install Go 1.24 or newer and the pinned quality tools:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go install golang.org/x/vuln/cmd/govulncheck@v1.5.0
```

Clone the repository and run the same checks used by CI:

```sh
git clone https://github.com/gopact-ai/acp.git
cd acp
make check
```

## Changes

- Keep changes focused and use the standard library unless a dependency has a
  clear correctness or security advantage.
- Add or update tests before changing protocol behavior. Use `go test -race`
  for concurrency changes and fuzz parsers or union decoders when appropriate.
- Document exported API changes and update the README when the runtime contract
  changes.
- Run `make check` before opening a pull request.

`types_gen.go` is derived from the checked-in ACP schema with the generator and
commit named in its header. It also contains intentional stable-v1 corrections.
Schema updates must update the schema provenance, generated types, required-field
maps, and their tests together.

By contributing, you agree that your contribution is licensed under the
Apache License 2.0.
