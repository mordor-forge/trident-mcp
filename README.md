# trident-mcp

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![trident-mcp MCP server](https://glama.ai/mcp/servers/mordor-forge/trident-mcp/badges/score.svg)](https://glama.ai/mcp/servers/mordor-forge/trident-mcp)

`trident-mcp` is a Go MCP server for AI-assisted 3D model generation and post-processing. It runs as a single stdio binary and currently uses the Tripo v3 API behind provider capability interfaces.

![Potion bottle generated with the Tripo v3 multiview pipeline](docs/assets/potion-bottle-collage.png)

[![trident-mcp MCP server](https://glama.ai/mcp/servers/mordor-forge/trident-mcp/badges/card.svg)](https://glama.ai/mcp/servers/mordor-forge/trident-mcp)

## Quick Start

Build locally:

```bash
go build -o ./trident-mcp ./cmd/trident-mcp
```

Run with a Tripo key:

```bash
TRIPO_API_KEY=tsk_your_key_here ./trident-mcp
```

One-command development bootstrap and smoke check:

```bash
go mod download && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4 && go test ./... -count=1
```

## MCP Client Config

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

## Documentation Map

| Need | Read |
| --- | --- |
| Agent instructions and common change patterns | [AGENTS.md](AGENTS.md) |
| Architecture, module boundaries, and Mermaid diagrams | [ARCHITECTURE.md](ARCHITECTURE.md) |
| MCP tool surface, models, and Tripo v3 behavior notes | [docs/MCP_TOOLS.md](docs/MCP_TOOLS.md) |
| Local setup, verification commands, and release workflow | [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) |
| Runtime config, E2E tests, and operational caveats | [docs/OPERATIONS.md](docs/OPERATIONS.md) |
| Design invariants and provider-boundary rationale | [docs/design/provider-boundaries.md](docs/design/provider-boundaries.md) |
| Architecture decision records | [docs/adr/](docs/adr/) |
| Security policy and threat model | [SECURITY.md](SECURITY.md), [THREAT_MODEL.md](THREAT_MODEL.md) |

## Configuration

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `TRIPO_API_KEY` | Yes | none | Tripo API key used for generation, polling, download, and post-processing calls |
| `TRIPO_BASE_URL` | No | `https://openapi.tripo3d.ai/v3` | Override for private deployments or tests |
| `MODEL_OUTPUT_DIR` | No | `~/generated_models` | Directory where downloaded models are written |

## Tool Overview

The server exposes tools for text/image/multiview 3D generation, image preparation, splats, async task polling, account balance and usage, file uploads, downloads, conversion, retopology, stylization, texture/refine/segment/complete workflows, rig checks, rigging, animation retargeting, and model/config inspection.

See [docs/MCP_TOOLS.md](docs/MCP_TOOLS.md) for the full tool contract and known Tripo v3 quirks.

## Development

The normal local gate is:

```bash
go build ./cmd/trident-mcp
go test ./... -count=1
go vet ./...
golangci-lint run
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Live E2E tests require `TRIPO_API_KEY` and can spend credits when opt-in generation flags are set. See [docs/OPERATIONS.md](docs/OPERATIONS.md#live-e2e-tests).

## License

Apache-2.0. See [LICENSE](LICENSE).
