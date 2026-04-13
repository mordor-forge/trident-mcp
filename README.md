# trident-mcp

`trident-mcp` is a Go MCP server for AI-assisted 3D model generation and post-processing.

The server is client-agnostic and works independently with any MCP-compatible client. You do not need any companion skills or extra MCP servers to use the core 3D generation, polling, download, and post-processing tools.

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
go build ./cmd/trident-mcp
```

Or install with Go:

```bash
go install github.com/mordor-forge/trident-mcp/cmd/trident-mcp@latest
```

## Configuration

The server reads configuration from environment variables:

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `TRIPO_API_KEY` | Yes | none | Tripo API key used for generation and editing calls |
| `MODEL_OUTPUT_DIR` | No | `~/generated_models` | Directory where downloaded models are written |

## Running

The server speaks MCP over stdio:

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

### Status and Download

- `task_status`
- `download_model`

`download_model` saves the task's actual output format. If you need a different format, run `convert_format` first and then download the conversion task.

### Post-processing

- `retopologize`
- `convert_format`
- `stylize`

### Introspection

- `list_models`
- `get_config`

`get_config` reports the active backend, output directory, and server version.

## Skills

The repo also includes companion agent skills under `skills/`:

- `skills/3d-gen/SKILL.md`
- `skills/multiview-3d/SKILL.md`
- `skills/3d-to-blender/SKILL.md`

These skills are optional. The MCP server itself works fine on its own in any MCP client.

Some of the companion skills are designed to compose `trident-mcp` with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp) for a fuller automated pipeline. In that setup, `gemini-media-mcp` can help with ideation, reference image generation, and multi-angle image creation, while `trident-mcp` handles reconstruction and post-processing. That pairing enables a more complete flow from idea to finished 3D model.

## Development

Run the local checks:

```bash
go test ./...
go vet ./...
```

There is also an opt-in E2E smoke test for live Tripo uploads:

```bash
TRIPO_API_KEY=tsk_your_key_here go test -tags=e2e -run "TestE2E_" ./internal/provider/tripo/ -v
```

## Release

GitHub Actions runs unit checks on pushes and pull requests, and runs the E2E smoke test on `main` and version tags. Tagged releases are built with GoReleaser.

Artifacts are stamped with the release version so the binary and MCP implementation metadata stay aligned.
