# Operations Guide

## Runtime Configuration

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `TRIPO_API_KEY` | Yes | none | Server refuses to start without it. |
| `TRIPO_BASE_URL` | No | `https://openapi.tripo3d.ai/v3` | Use only for private deployments or tests. |
| `MODEL_OUTPUT_DIR` | No | `~/generated_models` | Created automatically if missing. |

The MCP server speaks stdio. It does not open an HTTP port.

## Install Paths

Local build:

```bash
go build -o ./trident-mcp ./cmd/trident-mcp
```

Go install:

```bash
CGO_ENABLED=0 go install ./cmd/trident-mcp
```

Release artifacts are built with CGO disabled.

## Live E2E Tests

E2E tests are behind the `e2e` build tag and require a live Tripo key:

```bash
TRIPO_API_KEY=tsk_... go test -tags=e2e -run "TestE2E_" ./internal/provider/tripo/ -v -timeout 10m
```

The default e2e path checks balance/usage and uploads a tiny file. Credit-spending generation checks require explicit flags:

```bash
TRIPO_API_KEY=tsk_... TRIPO_E2E_GENERATE=1 go test -tags=e2e -run TestE2E_TextToModelLifecycle ./internal/provider/tripo/ -v -timeout 10m
TRIPO_API_KEY=tsk_... TRIPO_E2E_FULL=1 go test -tags=e2e -run TestE2E_FullV3CreationMatrix ./internal/provider/tripo/ -v -timeout 10m
```

Run `get_balance` or the default e2e balance test before credit-spending runs. Do not run generation E2E in PR CI.

## Claude Code Manual Testing

Manual prompt suites are local operator notes and are intentionally not tracked in
the repository.

For a fresh Claude Code install:

1. Build or install the binary.
2. Configure a stdio MCP server with `TRIPO_API_KEY`.
3. Copy optional skills from `skills/` into the client-specific skills directory.
4. Start a fresh client session so the MCP process and skills reload.

## Release

Tag with `vX.Y.Z` and push. CI runs the full test job, then GoReleaser builds linux/darwin/windows binaries for amd64 and arm64.

`main.version` is stamped by GoReleaser through `-ldflags`. Development builds use `dev`.

## Failure Handling

- If Tripo rejects a request before creating a task, return the API error with context.
- If task polling returns `failed`, `cancelled`, or `expired`, do not call `download_model`.
- If a desired output format differs from the completed task's actual format, use `convert_format`.
- If an API option is account-tier gated, keep the last successful task ID and continue only with steps that can use it.
