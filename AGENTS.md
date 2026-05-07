# AGENTS.md

## What this is

Go MCP server for AI-assisted 3D model generation. Single binary (`cmd/trident-mcp/main.go`), stdio transport, currently backed by the Tripo API. The provider is behind four interfaces in `internal/provider/interfaces.go` — `ModelGenerator`, `ModelStatus`, `ModelPostProcessor`, `ModelLister` — so adding a new backend means implementing those interfaces without touching the server layer.

## Build and test

Install the CI-pinned linter if it is not already available:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4
```

```bash
go build ./cmd/trident-mcp        # build the binary
go test ./... -count=1             # unit tests (no API key needed)
go vet ./...                       # static analysis
golangci-lint run                  # lint (CI uses v2.11.4)
go run golang.org/x/vuln/cmd/govulncheck@latest ./... # security scan
```

CI gates build -> unit tests -> vet -> golangci-lint. CI also runs `govulncheck` in advisory mode.

### E2E tests

E2E tests are behind a build tag and require a live Tripo key:

```bash
TRIPO_API_KEY=tsk_... go test -tags=e2e -run "TestE2E_" ./internal/provider/tripo/ -v -timeout 10m
```

E2E only runs in CI on pushes to `main` (not on PRs).

## Project structure

```
cmd/trident-mcp/main.go    – entrypoint, wires config → provider → server
internal/
  config/                   – env-var loading (TRIPO_API_KEY, MODEL_OUTPUT_DIR)
  provider/
    interfaces.go           – four provider interfaces
    types.go                – shared request/response types
    tripo/                  – Tripo API implementation + unit tests + e2e_test.go
  server/
    server.go               – MCP server setup, tool registration, stdio runner
    tools_generation.go     – text_to_3d, image_to_3d, multiview_to_3d handlers
    tools_postprocess.go    – retopologize, convert_format, stylize handlers
    tools_config.go         – list_models, get_config handlers
skills/                     – optional companion agent skills (not part of the server)
```

## Key conventions

- **Version stamping**: `main.version` is set via `-ldflags` by GoReleaser at release time. During dev builds it defaults to `"dev"`. Don't hardcode version strings.
- **No golangci-lint config file**: CI uses `golangci-lint run` with default settings (v2.11.4). No `.golangci.yml` in the repo.
- **CGO disabled**: both Dockerfile and GoReleaser set `CGO_ENABLED=0`.
- **Tool registration is conditional**: tools only register when their backing interface is non-nil (see `server.go:66-76`). Tests can pass `nil` for unused interfaces.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `TRIPO_API_KEY` | Yes (runtime) | — | Server refuses to start without it |
| `MODEL_OUTPUT_DIR` | No | `~/generated_models` | Created automatically if missing |

## Release

Tag with `vX.Y.Z` and push. CI runs the full test job (build, unit tests, vet, lint, plus advisory `govulncheck`), then GoReleaser builds cross-platform binaries (linux/darwin/windows, amd64/arm64).
