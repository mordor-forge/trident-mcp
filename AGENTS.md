# AGENTS.md

## What this is

Go MCP server for AI-assisted 3D model generation. Single binary (`cmd/trident-mcp/main.go`), stdio transport, currently backed by the Tripo API. The provider is behind capability interfaces in `internal/provider/interfaces.go` — generation, image generation, status/common APIs, model/mesh processing, animation, and catalog listing — so adding a new backend means implementing the relevant interfaces without touching the server layer.

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

### Single-file verification

Go type-checks packages rather than standalone files. For a single changed Go file, verify the containing package before running the full gate:

```bash
gofmt -w internal/provider/tripo/generation.go
golangci-lint run internal/provider/tripo/generation.go
go test ./internal/provider/tripo -run '^$'       # package type-check/compile
go test ./internal/provider/tripo -run TestTextToModel_WithOptions -count=1
```

For Markdown files with Mermaid diagrams, extract each fenced `mermaid` block to a temporary `.mmd` file and run:

```bash
mmdc -i /tmp/diagram.mmd -o /tmp/diagram.svg
```

### E2E tests

E2E tests are behind a build tag and require a live Tripo key:

```bash
TRIPO_API_KEY=tsk_... go test -tags=e2e -run "TestE2E_" ./internal/provider/tripo/ -v -timeout 10m
TRIPO_API_KEY=tsk_... TRIPO_E2E_GENERATE=1 go test -tags=e2e -run TestE2E_TextToModelLifecycle ./internal/provider/tripo/ -v -timeout 10m
TRIPO_API_KEY=tsk_... TRIPO_E2E_FULL=1 go test -tags=e2e -run TestE2E_FullV3CreationMatrix ./internal/provider/tripo/ -v -timeout 10m
```

The default e2e path checks balance/usage and uploads a tiny file. `TRIPO_E2E_GENERATE=1` and `TRIPO_E2E_FULL=1` spend credits after balance gating. E2E only runs in CI on pushes to `main` (not on PRs).

## Project structure

```
cmd/trident-mcp/main.go    – entrypoint, wires config → provider → server
internal/
  config/                   – env-var loading (TRIPO_API_KEY, MODEL_OUTPUT_DIR)
  provider/
    interfaces.go           – provider capability interfaces
    types.go                – shared request/response types
    tripo/                  – Tripo v3 API implementation + unit tests + e2e_test.go
  server/
    server.go               – MCP server setup, tool registration, stdio runner
    tools_generation.go     – generation, status/download, batch task, account, and file upload handlers
    tools_image.go          – text_to_image, image_to_image, multiview image, splat handlers
    tools_postprocess.go    – retopologize, convert_format, stylize handlers
    tools_v3_processing.go  – import/refine/texture, mesh, rigging, animation handlers
    tools_config.go         – list_models, get_config handlers
skills/                     – optional companion agent skills (not part of the server)
```

## Key conventions

- **Version stamping**: `main.version` is set via `-ldflags` by GoReleaser at release time. During dev builds it defaults to `"dev"`. Don't hardcode version strings.
- **No golangci-lint config file**: CI uses `golangci-lint run` with default settings (v2.11.4). No `.golangci.yml` in the repo.
- **CGO disabled**: both Dockerfile and GoReleaser set `CGO_ENABLED=0`.
- **Tool registration is conditional**: tools only register when their backing interface is non-nil (see `server.go:66-76`). Tests can pass `nil` for unused interfaces.
- **Design docs track boundaries**: when changing provider interfaces, task lifecycle semantics, model catalog behavior, upload/download behavior, or tool registration rules, review and update `docs/design/provider-boundaries.md` and the relevant ADR.

## Pattern references

- **New MCP tool**: follow the pattern in `internal/server/tools_generation.go` for registration/handlers and `internal/provider/tripo/generation.go` for provider request mapping.
- **New async Tripo endpoint**: see `internal/provider/tripo/v3_features.go` for task creation and `internal/server/tools_v3_processing.go` for tool handlers.
- **Generation request option**: follow the pattern in `internal/provider/tripo/generation.go`; translate MCP camelCase fields to Tripo snake_case payload keys in the provider layer.
- **Format or post-processing option**: follow the pattern in `internal/provider/tripo/postprocess.go`; validate known enums locally before making credit-spending API calls.
- **Model catalog update**: see `internal/provider/tripo/models.go` for aliases and static catalog entries, then update `docs/MCP_TOOLS.md`.
- **Agent skill update**: use `skills/3d-gen/SKILL.md` as a template when the preferred workflow, tool sequence, or live API caveat changes.
- **Longer examples**: see `docs/patterns/common-changes.md` for copy-modify checklists.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `TRIPO_API_KEY` | Yes (runtime) | — | Server refuses to start without it |
| `TRIPO_BASE_URL` | No | `https://openapi.tripo3d.ai/v3` | Override for private deployments/tests |
| `MODEL_OUTPUT_DIR` | No | `~/generated_models` | Created automatically if missing |

## Release

Tag with `vX.Y.Z` and push. CI runs the full test job (build, unit tests, vet, lint, plus advisory `govulncheck`), then GoReleaser builds cross-platform binaries (linux/darwin/windows, amd64/arm64).
