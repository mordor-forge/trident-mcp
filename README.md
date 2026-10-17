# trident-mcp

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![trident-mcp MCP server](https://glama.ai/mcp/servers/mordor-forge/trident-mcp/badges/score.svg)](https://glama.ai/mcp/servers/mordor-forge/trident-mcp)

An unofficial MCP server for the [Tripo 3D](https://www.tripo3d.ai) API that covers every documented, deployed v3 workflow. A few endpoints are deliberately left out; they are listed in [docs/MCP_TOOLS.md](docs/MCP_TOOLS.md#not-exposed). Agents can:

- generate meshes from text, one image or several views;
- make reference images and multiview sets;
- texture, retopologize, UV-unwrap, segment, rig, animate and convert meshes;
- produce Gaussian splats.

The outputs land as files on disk. It ships as one Go binary over stdio, plus a Claude Code plugin with a workflow skill.

![Potion bottle generated with the Tripo v3 multiview pipeline](docs/assets/potion-bottle-collage.png)

## Quick start

**Claude Code.** Put the binary on your `PATH`, then install the plugin:

```bash
go install github.com/mordor-forge/trident-mcp/cmd/trident-mcp@latest   # or grab a release archive
```

```text
/plugin marketplace add mordor-forge/claude-plugins
/plugin install trident@mordor-forge
```

The plugin asks for your Tripo key and output folder, and adds the `tripo-3d` skill. Then ask for what you want, e.g. *"make a game-ready low-poly treasure chest and export it as FBX"*.

**Any MCP client:**

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

## How it works

1. A task-creating tool (`text_to_model`, `rig_model`, `convert_model`, …) validates its arguments locally, uploads any local files, and returns a `task_id`.
2. `wait_for_task` waits with progress notifications, then saves the outputs to `<output dir>/<task_id>/`. For example: `model.glb`, `rendered_image.webp`, `generated_image.png`, `front_view.png`.
3. Pass a `task_id` as the next tool's `input` to chain steps: image → mesh → rig → animation → FBX.

Arguments use Tripo's own field names, so [Tripo's docs](https://developers.tripo3d.ai) apply to them directly. A few convenience fields differ from Tripo's wire format, such as multiview `inputs` given as a `{front, left, back, right}` map. See [docs/MCP_TOOLS.md](docs/MCP_TOOLS.md) for all 28 tools.

## Configuration

| Variable | Default | |
| --- | --- | --- |
| `TRIPO_API_KEY` | required | From <https://platform.tripo3d.ai> |
| `MODEL_OUTPUT_DIR` | `~/generated_models` | Created if missing |
| `TRIPO_BASE_URL` | `https://openapi.tripo3d.ai/v3` | China-region keys: `https://openapi.tripo3d.com/v3` |

## Documentation

| Need | Read |
| --- | --- |
| Every tool, argument and behaviour | [docs/MCP_TOOLS.md](docs/MCP_TOOLS.md) |
| Architecture and the `internal/tripo` module | [ARCHITECTURE.md](ARCHITECTURE.md), [docs/design/tripo-module.md](docs/design/tripo-module.md) |
| Domain vocabulary | [GLOSSARY.md](GLOSSARY.md) |
| Contributing, change recipes | [AGENTS.md](AGENTS.md), [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md), [docs/patterns/common-changes.md](docs/patterns/common-changes.md) |
| Running, live tests, failure handling | [docs/OPERATIONS.md](docs/OPERATIONS.md) |
| Decisions | [docs/adr/](docs/adr/) |
| Security | [SECURITY.md](SECURITY.md), [THREAT_MODEL.md](THREAT_MODEL.md) |

## Tripo terms and trademarks

trident-mcp is an independent project. It is not affiliated with or endorsed by Tripo, VAST or Holymolly Ltd, and "Tripo" is a trademark of its owner. Your use of the Tripo API is governed by [Tripo's terms](https://www.tripo3d.ai/terms). Under them:

- Use your own API key, and keep it private.
- Tripo's written authorization is required before offering a shared or hosted deployment to other people.
- Ownership of generated models depends on your Tripo plan.

Credit prices quoted in the docs and skill are dated summaries; [Tripo's pricing page](https://developers.tripo3d.ai/en/pricing) is authoritative.

## License

Apache-2.0. See [LICENSE](LICENSE). The license covers this repository's code, not Tripo's service or the models you generate.
