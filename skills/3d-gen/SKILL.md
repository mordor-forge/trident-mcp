---
name: 3d-gen
description: Use when the user asks to generate, create, import, refine, texture, rig, animate, convert, stylize, segment, or inspect a 3D model, mesh, asset, Gaussian splat, or game-ready object with trident-mcp.
---

# 3D Generation Skill

You are an expert 3D model generation assistant. Your job is to translate the user's creative vision into high-quality 3D models using the trident-mcp tools, which connect to 3D generation providers (currently Tripo).

`trident-mcp` works independently with any MCP-capable client. This skill is an optional companion workflow layer on top of the MCP server.

For a more automated end-to-end pipeline, some related skills can be paired with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp) to help turn an idea into reference imagery and then into a 3D model. In particular, use the `multiview-3d` skill when the workflow benefits from Gemini-generated reference or multi-angle images before 3D reconstruction.

## Available Models

| Version | ID | Best For | Multi-view? |
|---------|-----|----------|-------------|
| **Tripo Turbo** | `tripo-turbo` (`turbo` alias, docs-listed) | Quick iterations if accepted by the account/API tier | No |
| **Tripo v2.5** | `v2.5-20250123` (`tripo-v2.5`, `v2.5` aliases) | Legacy compatibility | Yes |
| **Tripo H3.0** | `v3.0-20250812` (`tripo-v3.0`, `v3.0` aliases) | High quality generation | Yes |
| **Tripo H3.1** (Default) | `v3.1-20260211` (`tripo-v3.1`, `v3.1` aliases) | Best overall quality | Yes |
| **Tripo P1** | `P1-20260311` (`tripo-p1`, `p1` aliases) | Low-poly/game-ready topology | Yes |

**Default to `v3.1-20260211`** unless the user needs low-poly/game-ready output (`P1-20260311`). P1 is already the low-poly model, so do not set `smartLowPoly` with P1; control the budget with `faceLimit`. Use `tripo-turbo` only when you have confirmed it is accepted by the live API/account tier.

## Current Tool Surface

Use `get_balance` before credit-spending work when a live key is configured. Use `get_usage` when the user asks where credits went.

Generation and prep:
- `text_to_3d`, `image_to_3d`, `multiview_to_3d`
- `text_to_image`, `image_to_image`, `image_to_multiview`, `edit_multiview`
- `image_to_splat`
- `upload_file`, `create_file_upload`, `import_model`

Processing:
- `texture_model`, `refine_model`, `segment_mesh`, `complete_mesh`
- `retopologize`, `convert_format`, `stylize`
- `rig_check`, `rig_model`, `retarget_animation`

Status and retrieval:
- `task_status` or `get_task`, `get_tasks`, `download_model`

## The Interactive Workflow

Most Tripo operations are **asynchronous** — you start a task, poll for completion, then download or feed the result into the next task. The user should not have to manage task IDs manually.

### Phase 1: Understand Intent

Read the user's request carefully. You need to understand:

1. **Subject** — What 3D model does the user want? (character, object, vehicle, architecture)
2. **Input type** — Text description, reference image, or multiple angle photos?
3. **Quality needs** — Quick draft or production-ready asset?
4. **Post-processing** — Will they need texture, refinement, retopology, segmentation, rigging, animation, format conversion, or stylization?

If the request is clear (e.g., "make a 3D model of a medieval sword"), skip to generation. If vague, ask 1-2 focused questions.

### Phase 2: Select Input Mode

Choose the input mode that matches the user's source material:

**Text-to-3D** (`text_to_3d`):
- Best for: Conceptual objects, quick prototypes, when no reference exists
- Quality: Good geometry, AI-generated textures
- Prompt tips: Be specific about shape, material, and style. "A weathered wooden treasure chest with iron bands and a rusty lock" beats "a chest"
- Useful options: `quad`, `smartLowPoly`, `generateParts`, `faceLimit`, `textureQuality`, `geometryQuality`, and seed fields. For P1, omit `smartLowPoly`.

**Image-to-3D** (`image_to_3d`):
- Best for: When the user has a reference photo or concept art
- Quality: Better geometry that follows the reference, texture aligned to input
- Accepts: Local file path or public URL
- Useful options: `textureAlignment`, `enableImageAutofix`, `orientation`, `quad`, `smartLowPoly`, and `generateParts`. For P1, omit `smartLowPoly`.

**Multi-view-to-3D** (`multiview_to_3d`):
- Best for: Highest quality reconstruction from multiple angles
- Quality: Best geometry and texture consistency
- Requires: 2-4 images from different angles (front, side, back, etc.) or a successful `image_to_multiview` / `edit_multiview` task ID via `taskId`
- Use the `multiview-3d` skill for the automated multi-angle pipeline with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp)

**Image preparation** (`text_to_image`, `image_to_image`, `image_to_multiview`, `edit_multiview`):
- Best for: Creating or repairing reference images before 3D reconstruction
- After `text_to_image`, poll the image task and use `output.imageUrl` / `output.imageUrls` as image input. Do not pass a text/image generation task ID where an image URL or file token is required.
- Use `image_to_multiview` when one good reference image should become ordered multi-angle inputs
- Use `edit_multiview` with per-view `prompts` entries when generated views need targeted correction before `multiview_to_3d`
- After `image_to_multiview` or `edit_multiview` succeeds, prefer passing its task ID as `taskId` to `multiview_to_3d` instead of extracting individual view URLs manually.

**Existing model import** (`upload_file` / `import_model`):
- Best for: Using a local or remote GLB, GLTF, FBX, OBJ, or STL as input to texture, segment, retopologize, convert, or rig workflows

**Gaussian splat** (`image_to_splat`):
- Best for: Novel-view / scene-like splat output rather than a conventional mesh workflow

### Phase 3: Generate and Monitor

The normal mesh generation flow is:

0. **`get_balance`** — check available credits before starting expensive generation when a live Tripo key is configured
1. **`text_to_3d`** / **`image_to_3d`** / **`multiview_to_3d`** / **`import_model`** — starts or imports a task, returns a task ID
2. **`task_status`** — poll with the task ID until status is `success`
3. **`download_model`** — save the completed model to disk in the task's actual output format

After starting generation, tell the user:
> "3D generation started! This typically takes 1-3 minutes. Checking status..."

Then poll `task_status` every 15-20 seconds until done. Report progress naturally:
> "Processing... 45% complete"
> "Model is ready! Downloading now..."

After download, report the file path and actual format, then offer review options.
If the user wants a different format than the task produced, run `convert_format` first and then `download_model` on the conversion task. Use `get_tasks` when tracking several related tasks.

### Phase 4: Interactive Review

After the model is generated:

> "Model saved to: [path]. What would you like to do?"
> 1. **Retopologize** — Create a clean lowpoly version (quad or triangle mesh)
> 2. **Convert format** — Export to FBX, OBJ, STL, USDZ, or 3MF
> 3. **Stylize** — Apply LEGO, voxel, Voronoi, or Minecraft style
> 4. **Texture/refine/segment/complete** — Improve textures, refine quality, split parts, or complete selected parts
> 5. **Rig/animate** — Check riggability, auto-rig, or retarget animation
> 6. **New variation** — Same concept, different take
> 7. **Send to Blender** — Import via the preferred Blender connector (official first, community fallback; use the 3d-to-blender skill)
> 8. **Done** — Keep this model

## Post-Processing Details

### Retopology (`retopologize`)

Reduces polygon count while preserving shape:

- **Triangle mesh** (`quad: false`): Standard triangle reduction, good for game engines
- **Quad mesh** (`quad: true`): Produces quads, better for further modeling and subdivision
- **Target faces**: Set `targetFaces` to control output density (e.g., 4000 for game-ready)
- **Parts**: Use `partNames` when processing selected segmented parts

### Format Conversion (`convert_format`)

Supported output formats:
- **GLTF** (default) — Web, three.js, universal
- **FBX** — Unity, Unreal, 3ds Max
- **OBJ** — Universal interchange
- **STL** — 3D printing
- **USDZ** — Apple AR Quick Look
- **3MF** — Modern 3D printing

To get one of these formats, start a conversion task first, then use `download_model` on that conversion task. `download_model` does not convert formats by itself.

Conversion can also control export details such as `textureSize`, `textureFormat`, `bake`, `packUV`, `exportVertexColors`, `pivotToCenterBottom`, `scaleFactor`, `partNames`, `exportOrientation`, and `fbxPreset`. Use canonical `exportOrientation` values `+x`, `-x`, `+y`, or `-y`; do not use `z_up`.

### Stylization (`stylize`)

Transform the model's appearance:
- **lego** — Brick-built appearance
- **voxel** — Blocky voxel art
- **voronoi** — Organic cellular pattern
- **minecraft** — Block-based Minecraft style

Use `blockSize` for Minecraft-style block sizing when the user cares about scale.

### Texture and Refinement

- `texture_model`: Generate or regenerate textures; supports text/image/style prompts, PBR, texture quality, part names, and baking controls.
- `refine_model`: Improve an existing model task when the user wants higher quality. If the API rejects refinement for the current account tier despite available credits, report that clearly and continue with other requested processing only when it still has valid input.

### Mesh Part Workflows

- `segment_mesh`: Split a model into editable parts.
- `complete_mesh`: Fill or complete selected segmented parts, using `partNames` when the user identifies specific regions.

### Rigging and Animation

1. Run `rig_check` before rigging to confirm whether the model is suitable.
2. Run `rig_model` with the desired rig type and output format (`glb` or `fbx`).
3. Run `retarget_animation` to apply one or more animation presets.
4. Download the resulting rigged or animated task output.

## Prompt Engineering for Text-to-3D

**Effective prompts:**
- Describe the object's shape, material, and surface detail
- Include scale reference: "a small figurine" vs "a life-size statue"
- Specify art style: "low-poly", "realistic", "stylized", "cartoon"
- Add material keywords: "metallic", "wooden", "glass", "stone"

**Use negative prompts** to exclude unwanted features:
- "blurry, low quality, broken geometry"
- "multiple objects" (when you want a single clean mesh)

**What NOT to do:**
- Don't describe scenes — focus on a single object
- Don't expect text/labels to render on the model
- Don't write overly long prompts — 20-50 words is the sweet spot
