# MCP Tool Reference

trident-mcp exposes the Tripo 3D API (v3) as MCP tools. Argument names are Tripo's own snake_case field names (ADR-0005), so Tripo's [API docs](https://developers.tripo3d.ai) apply to them directly. Every tool publishes a JSON Schema with descriptions and enums, and the server enforces it.

## Contract

- **Task-creating tools spend credits.** They return a `task_id` immediately. The structured result also carries `operation`, the resolved `model`, what the task `produces`, and any `warnings`.
- **`wait_for_task` is the normal next step.** It polls until the task reaches a terminal status (`success`, `failed`, `cancelled`, `banned`, `expired`) and sends MCP progress notifications. On success it saves every output to `<MODEL_OUTPUT_DIR>/<task_id>/<output>.<ext>`. If `timeout_seconds` (default 120, max 900) passes first, it returns `done: false`; call it again.
- **Chain tasks by ID.** Pass a `task_id` as the next tool's `input`.
- **Inputs.** Any `input` field (and `inputs`, `texture_prompt.image`, `ref_image`, ...) accepts:
  - an absolute local path, uploaded automatically;
  - a public URL;
  - a file token;
  - a task ID.

  Local uploads are limited to images (png, jpg, webp; up to 20 MB) and meshes (glb, gltf, fbx, obj, stl, 3mf, usdz; up to 150 MB). Files over 50 MB use a presigned upload.
- **Output URLs are signed and expire.** Tripo's docs say minutes; the live run below saw 24 hours. `download_task` and `wait_for_task` re-fetch the task before downloading, so the URLs they use are always fresh.
- **Downloads keep the task's real format.** Use `convert_model` to get another format, then download that task.
- **Model versions resolve through a static catalog.** Aliases (`p1`, `v3.1`, `rig-v3.0`, ...) map to dated IDs. A missing `model` becomes the tool's default. A field the chosen version rejects is dropped when unset and refused when set; for example, `quad: true` on P1 fails before any credits are spent. Unknown versions are passed through with a warning.
- **Errors carry Tripo's code, message, suggestion and request ID.** Rate limits (HTTP 429) are retried automatically after the `Retry-After` delay. Nothing else is retried, so a create call is never billed twice.

## Generation

| Tool | Endpoint | Notes |
| --- | --- | --- |
| `text_to_model` | `POST /generation/text-to-model` | `prompt` plus generation options: `model`, `face_limit`, `texture`, `pbr`, `texture_quality` (`fast` needs `texture_version` v3.5), `texture_version`, `delight`, `geometry_quality`, `quad`, `smart_low_poly`, `generate_parts`, `auto_size`, `compress`, `export_uv`, `export_orientation`, seeds. |
| `image_to_model` | `POST /generation/image-to-model` | `input` is one image, or the task ID of `text_to_image` / `image_to_image`. Adds `enable_image_autofix`, `texture_alignment`, `orientation`. |
| `multiview_to_model` | `POST /generation/multiview-to-model` | `inputs: {front, left, back, right}` (front plus at least one more, any subset), or `input` = an `image_to_multiview` / `edit_multiview` task ID. |

## Images

| Tool | Endpoint | Notes |
| --- | --- | --- |
| `text_to_image` | `POST /generation/text-to-image` | `model` defaults to `seedream_v4`. Also `size`, `quality`, `background`, `aspect_ratio`, `output_format`, `watermark`, `template` (`t_pose`, `asset_extraction`, ...). Tripo allows one image task at a time per account. |
| `image_to_image` | `POST /generation/image-to-image` | `input` or up to 16 `inputs` (referenced as `image[1]`, ...). `model` defaults to `seedream_v5`. `template` `3d_enhance` prepares photos for 3D. |
| `image_to_multiview` | `POST /generation/image-to-multiview` | Produces front, left, back and right views. |
| `edit_multiview` | `POST /generation/edit-multiview` | `input` is a multiview task ID; `prompts` holds 1-4 `{view, prompt}` items (5 credits each). |
| `image_to_splat` | `POST /generation/image-to-splat` | Produces a `.splat` Gaussian splat (30 credits). |

## Model processing

| Tool | Endpoint | Notes |
| --- | --- | --- |
| `import_model` | `POST /models/import` | Brings an existing GLB, FBX, OBJ, STL or GLTF (URL only) into Tripo as a task. |
| `texture_model` | `POST /models/texture` | `texture_prompt: {text \| image \| images[4], style_image}`. `model` is one of v3.0 (default), v3.5 or v2.5. Also `texture_quality`, `pbr`, `bake`, `delight`, `part_names`, `compress`. |
| `convert_model` | `POST /models/convert` | `format` is one of GLTF, FBX, OBJ, STL, USDZ or 3MF. Also `quad`, `force_symmetry`, `face_limit`, `flatten_bottom`, `texture_size`, `texture_format`, `pack_uv`, `pivot_to_center_bottom`, `scale_factor`, `with_animation`, `animate_in_place`, `export_orientation`, `fbx_preset` (blender, 3dsmax, mixamo, bake_scale). Non-default geometry or texture options are billed as "advanced" (10 credits instead of 5). |

## Mesh processing

| Tool | Endpoint | Notes |
| --- | --- | --- |
| `retopologize` | `POST /mesh/decimate` | Smart low-poly retopology. `model` v2.0 (default), v3.0 (adds `smart_uv`), v1.0 (plain decimation; `face_limit` required). |
| `smart_uv` | `POST /mesh/uv` | UV unwrap for meshes up to about 80k triangles, plus a UV layout preview (20 credits). |
| `segment_mesh` | `POST /mesh/segment` | `model` v1.0 (geometric) or v2.0 (beta; `segmentation_granularity`, `split_by_connectivity`, `ref_image`). |
| `complete_mesh` | `POST /mesh/complete` | `input` is a `segment_mesh` task. `completion_mode` is `ai_completion` or `quick_cap`; `part_names` is optional. |
| `smart_segment` | `POST /mesh/smartsegment` | `seg_type` `image` (85 credits) or `model` (GLB plus a 16-number `transform`; 55 credits), with a text `hint`. |

## Animation

| Tool | Endpoint | Notes |
| --- | --- | --- |
| `rig_check` | `POST /animations/rig-check` | Free task. Its outputs `riggable` and `rig_type` are shown by `wait_for_task` / `get_task`. |
| `rig_model` | `POST /animations/rig` | `model`: v1.0-20240301 (default, bipeds only), v2.5-20260210 (all creatures), v3.0-20260909 (best humanoids, `rig_type` `others`). Also `spec` (tripo, mixamo, actorcore, unity, unreal, vrm) and `out_format` (glb, fbx). 25 credits. |
| `retarget_animation` | `POST /animations/retarget` | `input` is a `rig_model` task. `animation` or up to 5 `animations` presets (`preset:walk`, `preset:quadruped:walk`, ...). 10 credits each. |

## Tasks

| Tool | Notes |
| --- | --- |
| `wait_for_task` | Waits for completion (with progress notifications), then saves outputs. `download: false` only waits. |
| `get_task` | Status, progress, queue position or ETA, error, output names, credits. |
| `get_tasks` | Up to 100 task IDs at once; unknown IDs are listed under `missed`. |
| `download_task` | Saves the outputs of a successful task. `outputs` limits it to the named ones (`model`, `rendered_image`, `generated_image`, `front_view`, ...). |

## Account, files and configuration

| Tool | Notes |
| --- | --- |
| `get_balance` | Available and frozen credits (1 credit = US$0.01). |
| `get_usage` | Per-task credit usage, with `limit` / `offset` paging. |
| `upload_file` | Uploads one local file and returns a reusable file token. Usually unnecessary, because inputs take paths. |
| `list_models` | The model catalog, optionally for one tool (`operation`), with aliases, defaults, stages and rejected options. |
| `get_config` | Tripo API host, output directory, version, and the date the catalog was last checked. |

## Verified live (2026-10-17)

`TRIPO_E2E_SPEND=1` ran against the global API and spent 80 credits. It confirmed:

- **Model versions.** Tripo rejects unknown 3D versions such as `tripo-turbo` with code 1004 before charging anything, and its error lists exactly the catalog's five 3D versions. `banana` and `seedream_v4` are accepted image model names.
- **text_to_image** (seedream_v4, 2048 px): 5 credits; output `generated_image_url` (PNG).
- **Signed output URLs** (`tripo-data.rg1.data.tripo3d.com`, CloudFront policy) were valid for 24 hours.
- **text_to_model** (H3.1, no texture): 10 credits, about 70 s. Outputs are `model_url` (GLB, about 40 MB untextured), `rendered_image_url` (WebP preview) and `generated_image_url` (the intermediate reference image).
- **image_to_multiview**: 10 credits. The view URLs are nested under `generate_multiview_image` (`front_view_url`, ...); downloads flatten them to `front_view`, `left_view`, and so on.
- **rig_check**: an asynchronous, free task whose output is `{riggable, rig_type}`.
- **texture_model**: the `texture_prompt: {text}` object shape works with `model` v3.5-20260815 and `texture_quality: fast`; 10 credits.
- **Rigging and conversion**: `rig_model` with v3.0-20260909 cost 25 credits, `retarget_animation` with `preset:walk` cost 10, and `convert_model` to FBX cost 5.

## Not exposed

- **Refine and stylize** (`/models/refine`, `/models/stylize`): Tripo has retired them.
- **Mesh local edit** (`/mesh/edit`): documented, but not yet deployed by Tripo.
- **OpenAI-compatible image endpoints** (`/openai/images/*`): they duplicate `text_to_image` / `image_to_image`.
- **Image-to-depth**: undocumented.
- **Presigned uploads**: used internally by uploads over 50 MB.
- **Webhooks**: not applicable to a local stdio server.
