# Operations Guide

## Runtime configuration

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `TRIPO_API_KEY` | yes | none | Create one at <https://platform.tripo3d.ai>. The server refuses to start without it. |
| `TRIDENT_TRIPO_API_KEY` | no | none | Set by the Claude Code plugin. It takes precedence, so an empty plugin setting falls back to `TRIPO_API_KEY`. |
| `TRIPO_BASE_URL` | no | `https://openapi.tripo3d.ai/v3` | Keys are valid in one region only. China-region keys need `https://openapi.tripo3d.com/v3`. |
| `MODEL_OUTPUT_DIR` | no | `~/generated_models` | Artifacts go to `<dir>/<task_id>/<output>.<ext>`. `~` is expanded, and the directory is created if missing. |

The server speaks MCP over stdio and opens no port.

## Install

- **Claude Code plugin.** Put the `trident-mcp` binary on your `PATH` (a release archive or `go install`), then:

  ```text
  /plugin marketplace add mordor-forge/claude-plugins
  /plugin install trident@mordor-forge
  ```

  The plugin asks for the API key (stored in the system keychain) and the output directory, starts the server, and adds the `tripo-3d` skill.
- **Any MCP client:**

  ```json
  {
    "mcpServers": {
      "trident": {
        "command": "trident-mcp",
        "env": { "TRIPO_API_KEY": "tsk_...", "MODEL_OUTPUT_DIR": "/abs/path/generated_models" }
      }
    }
  }
  ```

- **From source:** `CGO_ENABLED=0 go install github.com/mordor-forge/trident-mcp/cmd/trident-mcp@latest`.

## Live e2e tests

```bash
TRIPO_API_KEY=tsk_... go test -tags=e2e -run TestE2E_ ./internal/tripo/ -v
TRIPO_API_KEY=tsk_... TRIPO_E2E_SPEND=1 go test -tags=e2e -run TestE2E_SpendingLifecycle ./internal/tripo/ -v -timeout 30m
```

- **Default run:** free. It reads balance and usage, uploads a 1-pixel PNG and looks up an unknown task.
- **Spending run:** about 80 credits, and skips itself below a 150-credit balance. It generates an image (seedream and banana) and an untextured mesh. On those it runs multiview, rig check, a v3.5 texture, an FBX convert, a v3.0 rig and a walk retarget. Every Output is logged, so any drift from Tripo's documented response shapes shows up in the test output.
- **Keeping files:** set `TRIPO_E2E_OUTPUT_DIR` to keep the downloaded files.
- **CI:** CI runs only the free run, on pushes to `main`.

## Failure handling

- **Rejected before a task exists** (bad parameters, insufficient credits, invalid key): the tool errors with Tripo's code, message and suggestion, and nothing is charged.
- **Task ends `failed`, `cancelled`, `banned` or `expired`:** `wait_for_task` reports it with `error_code` and `error_message`. Credits frozen for a failed task are refunded.
- **Rate limited (HTTP 429):** calls are retried automatically. Image generation allows one concurrent task per account, so run image tasks one at a time.
- **Download fails with HTTP 403/404:** the signed URL expired. `download_task` always fetches fresh URLs; call it again.
- **Wrong output format:** run `convert_model` and download that task. Downloads never convert.

## Release

1. Run `scripts/sync-version.sh X.Y.Z`, commit, then tag `vX.Y.Z` and push.
2. `release.yml` verifies tests and that the manifest versions match the tag.
3. GoReleaser publishes linux, darwin and windows binaries for amd64 and arm64. Each archive also holds `LICENSE`, `README.md`, `.claude-plugin/plugin.json` and `skills/`, so an extracted archive is a plugin directory.
4. Build provenance is attested.
5. `ghcr.io/mordor-forge/trident-mcp` gets a multi-arch distroless image with SBOM and provenance.

`main.version` is stamped through `-ldflags`.
