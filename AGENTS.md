# AGENTS

## Project

- Name: `datadiff`
- Module: `github.com/dashmug/datadiff`
- Go version: `1.24`

## Commands

Run these before opening a PR:

```bash
go build ./...
go test -race ./...
go vet ./...
golangci-lint run
```

## Layout

- Keep a flat root package for public API.
- Do not add `pkg/` or `cmd/` for this library.
- Use `internal/` only for private implementation details when the codebase grows.

## Coding conventions

- Format with `gofmt`.
- Follow Effective Go.
- Prefer explicit error returns over panics.
- Avoid package-level mutable global state.
