# trident-mcp

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![trident-mcp MCP server](https://glama.ai/mcp/servers/mordor-forge/trident-mcp/badges/score.svg)](https://glama.ai/mcp/servers/mordor-forge/trident-mcp)

`trident-mcp` is a Go MCP server for AI-assisted 3D model generation and post-processing.

[![trident-mcp MCP server](https://glama.ai/mcp/servers/mordor-forge/trident-mcp/badges/card.svg)](https://glama.ai/mcp/servers/mordor-forge/trident-mcp)

The server is client-agnostic and works independently with any MCP-compatible client. You do not need any companion skills or extra MCP servers to use the core 3D generation, polling, download, and post-processing tools.

For the code-level layout, data flow, and extension boundaries, see [ARCHITECTURE.md](ARCHITECTURE.md).

It currently ships with a Tripo-backed provider and exposes tools for:

- text-to-3D generation
- image-to-3D generation
- multiview-to-3D generation
- async task polling
- model download
- retopology
- format conversion
- stylization
- model catalog and server config inspection

## Requirements

- Go 1.25+
- A Tripo API key in `TRIPO_API_KEY`

## Install

Build locally:

```bash
go build -o ./trident-mcp ./cmd/trident-mcp
```

Or install with Go:

```bash
go install github.com/mordor-forge/trident-mcp/cmd/trident-mcp@latest
```

If you install with `go install`, make sure your Go bin directory is on `PATH`.
By default that is usually `$(go env GOPATH)/bin` (often `~/go/bin`) unless you use `GOBIN`.

## Configuration

The server reads configuration from environment variables:

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `TRIPO_API_KEY` | Yes | none | Tripo API key used for generation, polling, download, and post-processing calls |
| `MODEL_OUTPUT_DIR` | No | `~/generated_models` | Directory where downloaded models are written |

## Running

The server speaks MCP over stdio.

If you built from source in the repo root, run the local binary directly:

```bash
TRIPO_API_KEY=tsk_your_key_here ./trident-mcp
```

If you installed with `go install` and your Go bin directory is on `PATH`, run:

```bash
TRIPO_API_KEY=tsk_your_key_here trident-mcp
```

Example MCP client configuration:

```json
{
  "mcpServers": {
    "trident-mcp": {
      "command": "trident-mcp",
      "env": {
        "TRIPO_API_KEY": "tsk_your_key_here",
        "MODEL_OUTPUT_DIR": "/absolute/path/to/generated_models"
      }
    }
  }
}
```

## Tools

### Generation

- `text_to_3d`
- `image_to_3d`
- `multiview_to_3d`

These tools start asynchronous tasks. Use `task_status` to poll for completion, then `download_model` to retrieve the task output.

For `multiview_to_3d`, supply 2-4 ordered views in Tripo's expected order: front, left, back, right. The server pads missing trailing views to match Tripo's current 4-slot multiview request shape.

### Status and Download

- `task_status`
- `download_model`

`download_model` saves the task's actual output format. If you need a different format, run `convert_format` first and then download the conversion task.

`task_status` reports Tripo's async state and progress. Depending on the upstream task, statuses can include `queued`, `running`, `success`, `failed`, `cancelled`, `expired`, or `unknown`.

### Post-processing

- `retopologize`
- `convert_format`
- `stylize`

### Introspection

- `list_models`
- `get_config`

`get_config` reports the active backend, output directory, and server version.

`list_models` returns the server's built-in compatibility catalog. It is intentionally static so the MCP surface stays predictable and testable; it does not perform live model discovery against Tripo.

## Architecture

The high-level architecture, runtime flow, and extension boundaries live in [ARCHITECTURE.md](ARCHITECTURE.md).

## Skills

The repo also includes companion agent skills under `skills/`:

- `skills/3d-gen/SKILL.md`
- `skills/multiview-3d/SKILL.md`
- `skills/3d-to-blender/SKILL.md`

These skills are optional. The MCP server itself works fine on its own in any MCP client.

Some of the companion skills are designed to compose `trident-mcp` with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp) for a fuller automated pipeline. In that setup, `gemini-media-mcp` can help with ideation, reference image generation, and multi-angle image creation, while `trident-mcp` handles reconstruction and post-processing. That pairing enables a more complete flow from idea to finished 3D model.

## Development

Install the same lint version used in CI:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4
```

Run the local checks:

```bash
go build ./cmd/trident-mcp
go test ./... -count=1
go vet ./...
golangci-lint run
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

There is also an opt-in E2E smoke test for live Tripo uploads:

```bash
TRIPO_API_KEY=tsk_your_key_here go test -tags=e2e -run "TestE2E_" ./internal/provider/tripo/ -v
```

The E2E test hits the live Tripo API, so it should be used sparingly. The normal development loop should rely on unit tests.

## Release

GitHub Actions gates build, unit tests, vet, and lint on pushes and pull requests. CI also runs `govulncheck` in advisory mode. The live E2E smoke test runs only on pushes to `main`. Tagged releases are built with GoReleaser.

Artifacts are stamped with the release version so the binary and MCP implementation metadata stay aligned.
