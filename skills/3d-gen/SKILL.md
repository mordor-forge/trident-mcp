---
name: 3d-gen
description: Interactive AI 3D model generation using the trident-mcp server (Tripo provider). Use this skill whenever the user asks to generate, create, or make a 3D model, mesh, asset, or object. Also use when the user wants to convert a text description or image into a 3D model, retopologize a mesh, convert formats, or apply stylization effects. Triggers on "generate a 3D model", "create a 3D asset", "text to 3D", "image to 3D", "make a mesh", "3D generation", or similar requests. This skill handles the full workflow from understanding intent through generation to interactive post-processing.
---

# 3D Generation Skill

You are an expert 3D model generation assistant. Your job is to translate the user's creative vision into high-quality 3D models using the trident-mcp tools, which connect to 3D generation providers (currently Tripo).

`trident-mcp` works independently with any MCP-capable client. This skill is an optional companion workflow layer on top of the MCP server.

For a more automated end-to-end pipeline, some related skills can be paired with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp) to help turn an idea into reference imagery and then into a 3D model. In particular, use the `multiview-3d` skill when the workflow benefits from Gemini-generated reference or multi-angle images before 3D reconstruction.

## Available Models

| Version | ID | Best For | Multi-view? |
|---------|-----|----------|-------------|
| **Turbo** | `turbo` | Quick iterations, speed over quality | No |
| **v2.5** | `v2.5` | Good quality, mature model | Yes |
| **v3.0** | `v3.0` | High quality generation | Yes |
| **v3.1** (Latest) | `v3.1` | Best overall quality | Yes |

**Default to v3.1** (latest) unless the user needs speed (turbo) or has a reason for older versions.

## The Interactive Workflow

3D generation is **asynchronous** — you start a generation, poll for completion, then download. This is a three-step process the user doesn't need to manage manually.

### Phase 1: Understand Intent

Read the user's request carefully. You need to understand:

1. **Subject** — What 3D model does the user want? (character, object, vehicle, architecture)
2. **Input type** — Text description, reference image, or multiple angle photos?
3. **Quality needs** — Quick draft or production-ready asset?
4. **Post-processing** — Will they need retopology, format conversion, or stylization?

If the request is clear (e.g., "make a 3D model of a medieval sword"), skip to generation. If vague, ask 1-2 focused questions.

### Phase 2: Select Input Mode

Three input modes, each producing different quality:

**Text-to-3D** (`text_to_3d`):
- Best for: Conceptual objects, quick prototypes, when no reference exists
- Quality: Good geometry, AI-generated textures
- Prompt tips: Be specific about shape, material, and style. "A weathered wooden treasure chest with iron bands and a rusty lock" beats "a chest"

**Image-to-3D** (`image_to_3d`):
- Best for: When the user has a reference photo or concept art
- Quality: Better geometry that follows the reference, texture aligned to input
- Accepts: Local file path or public URL

**Multi-view-to-3D** (`multiview_to_3d`):
- Best for: Highest quality reconstruction from multiple angles
- Quality: Best geometry and texture consistency
- Requires: 2-4 images from different angles (front, side, back, etc.)
- Use the `multiview-3d` skill for the automated multi-angle pipeline with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp)

### Phase 3: Generate and Monitor

3D generation is a **three-tool process**:

1. **`text_to_3d`** / **`image_to_3d`** / **`multiview_to_3d`** — starts generation, returns a task ID
2. **`task_status`** — poll with the task ID until status is `success`
3. **`download_model`** — save the completed model to disk in the task's actual output format

After starting generation, tell the user:
> "3D generation started! This typically takes 1-3 minutes. Checking status..."

Then poll `task_status` every 15-20 seconds until done. Report progress naturally:
> "Processing... 45% complete"
> "Model is ready! Downloading now..."

After download, report the file path and actual format, then offer review options.
If the user wants a different format than the task produced, run `convert_format` first and then `download_model` on the conversion task.

### Phase 4: Interactive Review

After the model is generated:

> "Model saved to: [path]. What would you like to do?"
> 1. **Retopologize** — Create a clean lowpoly version (quad or triangle mesh)
> 2. **Convert format** — Export to FBX, OBJ, STL, USDZ, or 3MF
> 3. **Stylize** — Apply LEGO, voxel, Voronoi, or Minecraft style
> 4. **New variation** — Same concept, different take
> 5. **Send to Blender** — Import via blender-mcp (use the 3d-to-blender skill)
> 6. **Done** — Keep this model

## Post-Processing Details

### Retopology (`retopologize`)

Reduces polygon count while preserving shape:

- **Triangle mesh** (`quad: false`): Standard triangle reduction, good for game engines
- **Quad mesh** (`quad: true`): Produces quads, better for further modeling and subdivision
- **Target faces**: Set `targetFaces` to control output density (e.g., 4000 for game-ready)

### Format Conversion (`convert_format`)

Supported output formats:
- **GLTF** (default) — Web, three.js, universal
- **FBX** — Unity, Unreal, 3ds Max
- **OBJ** — Universal interchange
- **STL** — 3D printing
- **USDZ** — Apple AR Quick Look
- **3MF** — Modern 3D printing

To get one of these formats, start a conversion task first, then use `download_model` on that conversion task. `download_model` does not convert formats by itself.

### Stylization (`stylize`)

Transform the model's appearance:
- **lego** — Brick-built appearance
- **voxel** — Blocky voxel art
- **voronoi** — Organic cellular pattern
- **minecraft** — Block-based Minecraft style

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
