# AGENTS.md

## What this is

A Go MCP server (stdio, single binary) for the Tripo 3D API v3, shipped with an agent skill as a Claude Code plugin. Everything Tripo-specific lives in the deep module `internal/tripo`. `internal/server` generates MCP tools from it. Read [GLOSSARY.md](GLOSSARY.md) for the domain words: Operation, Task, Input, Output, Artifact, Model version, Catalog.

## Build and test

```bash
go build ./cmd/trident-mcp
go test ./... -count=1        # unit tests; no API key, no network
go vet ./...
golangci-lint run             # CI pins v2.11.4, default config (no .golangci.yml)
go run golang.org/x/vuln/cmd/govulncheck@latest ./...   # advisory in CI
```

Install the pinned linter with `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4`. CI gates build → tests → vet → lint.

Go checks packages, not files. For one changed file, run `gofmt -w <file>` and `go test ./<its package> -count=1`. Validate Mermaid blocks in Markdown with `mmdc -i block.mmd -o /tmp/out.svg`.

### Live API

```bash
TRIPO_API_KEY=tsk_... go test -tags=e2e -run TestE2E_ ./internal/tripo/ -v               # free calls only
TRIPO_API_KEY=tsk_... TRIPO_E2E_SPEND=1 go test -tags=e2e -run TestE2E_SpendingLifecycle ./internal/tripo/ -v -timeout 30m   # ~80 credits
```

The spending run checks the balance first. Run it after changing request shapes, output parsing or the Catalog, and only with the user's go-ahead, because it spends their credits. CI runs only the free e2e calls, on pushes to `main`.

## Change recipes

- **New Tripo endpoint:**
  1. Add a request type and an `Operation` in `internal/tripo/operations.go`.
  2. Add `addOperation(s, tripo.X)` in `internal/server/server.go`.
  3. Add a row in `docs/MCP_TOOLS.md`; `TestDocs_DescribeEveryTool` fails until it exists.
  4. Add a table case in `internal/tripo/operations_test.go` and in `internal/server/server_test.go`; `TestServer_ListsTheWholeSurface` lists every tool name.
- **New parameter:** add one field. Its `json` tag is Tripo's exact snake_case field name (ADR-0005). Add a `jsonschema` description for agents, and an `enum:"a,b"` tag when Tripo documents a fixed set. Pointer types (`*bool`, `*int`) keep `false` and `0` sendable.
- **Cross-field rule** (e.g. "fast needs v3.5"): add it to the Operation's `check` func, only when Tripo documents it. Local checks exist to fail before credits are spent.
- **New or retired model version:** edit `catalog` in `internal/tripo/catalog.go`. `Rejects` lists the fields the version refuses, `DefaultFor` names the Operations it is the default for, and `CatalogDate` must be bumped. Retired entries are deleted rather than kept with a warning.
- **New output kind:** usually nothing to do. Any `*_url` field in a Task's Output becomes a downloadable Artifact. Add the extension to `knownFormats` in `download.go` if detection falls back to `.bin`.
- **Skill change:** edit `skills/tripo-3d/`. Keep model and price facts in its `references/` with a date. The skill reads `list_models` for anything the Catalog already knows.

## Conventions

- Tests cross the real seams. `tripotest` is the fake Tripo API; server tests use an in-memory MCP client in front of the real `tripo.Client`. Mocks of Go interfaces are not used (ADR-0004).
- `main.version` is stamped by GoReleaser via `-ldflags`; dev builds report `dev`. Builds use `CGO_ENABLED=0`.
- Design notes: changing the task lifecycle, uploads, downloads, error handling or the Catalog's semantics means updating [docs/design/tripo-module.md](docs/design/tripo-module.md) and the relevant ADR in `docs/adr/`.

## Environment

| Variable | Default | Notes |
|---|---|---|
| `TRIPO_API_KEY` | required | `TRIDENT_TRIPO_API_KEY` (set by the plugin) takes precedence |
| `TRIPO_BASE_URL` | `https://openapi.tripo3d.ai/v3` | China-region keys need `https://openapi.tripo3d.com/v3` |
| `MODEL_OUTPUT_DIR` | `~/generated_models` | Artifacts go to `<dir>/<task_id>/`; created if missing |

## Release

Tag `vX.Y.Z` and push. CI runs the test job, then GoReleaser builds linux, darwin and windows binaries for amd64 and arm64. Bump `version` in `.claude-plugin/plugin.json` to match.
