# Development Guide

## Prerequisites

- Go 1.26+
- `golangci-lint` v2.14.0: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.14.0`
- `TRIPO_API_KEY`, only to run the server or the live e2e tests
- Optional: `mmdc` (mermaid-cli) for diagram checks

Bootstrap and smoke check in one line:

```bash
go mod download && go test ./... -count=1
```

## Verification gate

```bash
go build ./cmd/trident-mcp
go test ./... -count=1
go vet ./...
golangci-lint run
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
python3 scripts/validate-skills.py skills
scripts/validate-packaging.sh
```

CI runs all of these, with `-race` on Linux and unit tests on macOS and Windows too.

## Where things go

[AGENTS.md](../AGENTS.md#change-recipes) has the recipes, and [patterns/common-changes.md](patterns/common-changes.md) has worked examples. In short:

- **Tripo behaviour** goes in `internal/tripo`, one module (ADR-0004).
- **MCP surface** goes in `internal/server`: one `addOperation` line per Operation, plus the task, account and config tools.
- **Tests** run against `tripotest` (the fake API) and an in-memory MCP client. There are no Go-interface mocks.

## Trying the server locally

```bash
go build -o ./trident-mcp ./cmd/trident-mcp
TRIPO_API_KEY=tsk_... claude --plugin-dir .    # loads .claude-plugin/plugin.json and skills/ from this checkout
```

The plugin manifest runs `trident-mcp` from your `PATH`. Put the freshly built binary first on `PATH`, or run `go install ./cmd/trident-mcp`.

## Documentation checks

- Every MCP tool must appear in `docs/MCP_TOOLS.md`. `TestDocs_DescribeEveryTool` enforces this.
- Validate each Mermaid block with `mmdc`.
- Domain words come from [GLOSSARY.md](../GLOSSARY.md). Lasting trade-offs go in `docs/adr/`.
