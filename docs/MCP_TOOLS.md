# MCP Tool Reference

This document describes the stable MCP surface exposed by `trident-mcp`. The server is client-agnostic; the companion skills under `skills/` are optional workflow guidance for agent clients.

## Tool Contract

- Most Tripo operations are asynchronous. Generation and processing tools return a task ID.
- Use `task_status` or `get_task` until the task reaches `success`, `failed`, `cancelled`, or `expired`.
- Use `download_model` only after a model-producing task succeeds.
- `download_model` saves the task's actual output format. To request another format, run `convert_format` first, poll the conversion task, then download that task.
- Use `get_balance` before credit-spending work and `get_usage` when auditing credit consumption.

## Generation And Image Prep

| Tool | Purpose | Notes |
| --- | --- | --- |
| `text_to_3d` | Generate a 3D model from text | Default model is H3.1 unless a model is specified. |
| `image_to_3d` | Generate a 3D model from one reference image | Accepts local path, public URL, or uploaded file token depending on request field. |
| `multiview_to_3d` | Generate from multiple ordered views | Accepts 2-4 image paths/URLs in front/left/back/right order, or `taskId` from `image_to_multiview` / `edit_multiview`. |
| `text_to_image` | Generate a reference image | Poll the image task and use `output.imageUrl`; do not pass the task ID as an image input. |
| `image_to_image` | Edit or derive an image from an input image | Use for reference cleanup before 3D reconstruction. |
| `image_to_multiview` | Create ordered multiview reference images from one source image | The resulting task ID can feed `multiview_to_3d` as `taskId`. |
| `edit_multiview` | Correct an existing multiview image set | Prefer `prompts` entries with `prompt` and `view` (`front`, `left`, `back`, `right`), then feed the edited task ID into `multiview_to_3d`. |
| `image_to_splat` | Generate a Gaussian splat task | Treat output as a splat artifact, not a conventional mesh. |

## Status, Account, Files, And Download

| Tool | Purpose |
| --- | --- |
| `task_status` / `get_task` | Poll a single task and return output metadata. |
| `get_tasks` | Batch query up to 100 task IDs. |
| `get_balance` | Return available and frozen credits. |
| `get_usage` | Return per-task credit usage history. |
| `upload_file` | Upload a local file to Tripo and return a reusable file token. |
| `create_file_upload` | Create a presigned upload URL and file token for direct upload flows. |
| `download_model` | Download a completed model artifact into `MODEL_OUTPUT_DIR`. |

## Post-Processing

| Tool | Purpose | Notes |
| --- | --- | --- |
| `retopologize` | Create lower-density triangle or quad mesh output | Use `targetFaces` and `quad` to shape topology. |
| `convert_format` | Convert to GLTF, FBX, OBJ, STL, USDZ, or 3MF | `exportOrientation` accepts `+x`, `-x`, `+y`, `-y`; `x_up` and `y_up` are aliases. |
| `stylize` | Apply LEGO, voxel, Voronoi, or Minecraft style | Use `blockSize` for Minecraft scale control. |
| `import_model` | Import an external model URL or file token into Tripo | Useful before texture, segment, rig, or conversion workflows. |
| `texture_model` | Generate or regenerate textures | Supports PBR, texture quality, text prompts, image prompts, style images, part names, and baking controls. |
| `refine_model` | Refine an existing model | May be account-tier gated even when credits are available. Surface the API error and keep valid earlier task IDs. |
| `segment_mesh` | Split a model into editable parts | Poll before using the result in `complete_mesh`. |
| `complete_mesh` | Complete a segmented model or selected parts | Use `partNames` only when the segmentation result exposes reliable names. |
| `rig_check` | Check whether a model can be rigged | Run before `rig_model`. |
| `rig_model` | Auto-rig a model | Supports `rig-v2.0` / `rig-v1.0`, rig type, skeleton spec, and GLB/FBX output. |
| `retarget_animation` | Apply preset animations to a rigged model | Poll and download the resulting animated artifact. |

## Models

| Namespace | IDs |
| --- | --- |
| 3D generation | `v3.1-20260211`, `v3.0-20250812`, `v2.5-20250123`, `tripo-v2.0`, `tripo-turbo`, `P1-20260311` |
| Image generation | `seedream_v4`, `seedream_v5`, `gemini-2.5-flash`, `gemini-3-pro`, `gemini-3.1-flash`, `chat_image_1`, `chat_image_1.5`, `chat_image_2` |
| Animation | `rig-v2.0`, `rig-v1.0` |

`v3.1-20260211` is the default high-quality 3D generation model. Use `P1-20260311` for low-poly/game-ready topology. P1 is already optimized for low-poly output, so the provider omits unsupported `smartLowPoly` when P1 is selected.

`list_models` is intentionally static. It provides the server's supported compatibility catalog and aliases without making live discovery calls.

## Live Behavior Notes

Manual Claude Code testing on July 8, 2026 observed:

- P1 low-poly text-to-3D completed in about 76 seconds and consumed 40 credits.
- H3.1 detailed text-to-3D completed in about 150 seconds and consumed 50 credits.
- Native multiview from one generated image consumed 45 credits total: 5 for `text_to_image`, 10 for `image_to_multiview`, and 30 for `multiview_to_3d`.
- Native multiview produced better-aligned textures than a single H3.1 text prompt for the tested object.
