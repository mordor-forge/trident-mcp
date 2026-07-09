# Development Guide

## Prerequisites

- Go 1.25+
- `TRIPO_API_KEY` only for runtime and live E2E tests
- CI-pinned linter:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4
```

One-command local bootstrap and smoke check:

```bash
go mod download && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4 && go test ./... -count=1
```

## Normal Verification Gate

Run this before committing code changes:

```bash
go build ./cmd/trident-mcp
go test ./... -count=1
go vet ./...
golangci-lint run
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

CI runs build, unit tests, vet, and golangci-lint. `govulncheck` runs in advisory mode.

## Single-File Verification

Go verifies packages, not isolated files, but agents should still keep checks scoped to the package containing the changed file before running the full gate.

For one changed Go file:

```bash
gofmt -w internal/provider/tripo/generation.go
golangci-lint run internal/provider/tripo/generation.go
go test ./internal/provider/tripo -run '^$'
go test ./internal/provider/tripo -run TestTextToModel_WithOptions -count=1
```

For one changed Markdown file with Mermaid diagrams:

```bash
mmdc -i /tmp/diagram.mmd -o /tmp/diagram.svg
```

Extract fenced Mermaid blocks from the changed Markdown file and validate each block with `mmdc`.

## Common Change Patterns

### Add A Tripo Task Endpoint

1. Add or extend the request/response type in `internal/provider/types.go`.
2. Add a capability method to `internal/provider/interfaces.go` only if this is a new capability family.
3. Map the request payload in `internal/provider/tripo/generation.go`, `postprocess.go`, or `v3_features.go`.
4. Add the server tool registration and handler in the matching `internal/server/tools_*.go` file.
5. Add provider `httptest` coverage for request shape and error mapping.
6. Add server tests that verify conditional registration and handler behavior.

Reference examples:

- Async generation endpoint: `internal/provider/tripo/generation.go`
- v3 image/model/mesh/animation endpoint: `internal/provider/tripo/v3_features.go`
- Tool registration: `internal/server/tools_generation.go`

### Add A Model Catalog Entry

1. Update `modelCatalog` in `internal/provider/tripo/models.go`.
2. Add aliases only when they map unambiguously to one API version.
3. Update `multiviewModels` behavior through capabilities, not a separate hand-edited list.
4. Add tests for alias resolution and `list_models` output.
5. Update [docs/MCP_TOOLS.md](MCP_TOOLS.md#models) when the user-facing model matrix changes.

### Add Post-Processing Options

1. Add the field to the relevant provider request type.
2. Convert camelCase MCP input to Tripo snake_case in the provider layer.
3. Validate known enums before calling Tripo when live failures would spend credits.
4. Add focused request-body tests before implementation.
5. Update the companion skill if agents should prefer or avoid the new option.

## Documentation Checks

When changing documentation:

- Keep README short and link to focused docs.
- Update `AGENTS.md` if agent workflow, verification, or common patterns change.
- Update ADRs when changing a boundary or an intentional trade-off.
- Validate every Mermaid block with `mmdc`.
- Re-run `agentready assess . -o .agentready` after agent-readiness docs change.

Generated `.agentready/` reports are local evidence, not source documentation.
