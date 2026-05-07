---
name: multiview-3d
description: Multi-angle 3D model generation pipeline using gemini-media-mcp for multi-view image generation and trident-mcp for 3D reconstruction. Use when the user wants to create a high-quality 3D model from a reference photo using multiple generated angles, or when they mention "multi-angle 3D", "multiview 3D", "photo to 3D with multiple views", "generate 3D from reference". This skill orchestrates the full pipeline from reference image through multi-angle generation to 3D reconstruction.
---

# Multi-View 3D Pipeline Skill

You are an expert at orchestrating the multi-angle 3D generation pipeline. This skill combines **gemini-media-mcp** (multi-angle image generation via Gemini) with **trident-mcp** (3D reconstruction via Tripo) to produce high-quality 3D models from reference photos or generated images.

This is a composed workflow skill. `trident-mcp` itself remains independently useful in any MCP-capable client, but this skill is specifically meant to couple it with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp) for a fuller automated pipeline from idea to 3D model.

## Prerequisites

This skill requires **two MCP servers** to be configured:

1. **gemini-media-mcp** — For generating consistent multi-angle views (needs `generate_image` and `edit_image` tools). Repository: <https://github.com/mordor-forge/gemini-media-mcp>
2. **trident-mcp** — For 3D reconstruction from multi-view images (needs `multiview_to_3d` tool)

Before starting, verify both are available by checking for the required tools. If either is missing, tell the user which MCP server to configure.

## The Pipeline Workflow

### Phase 1: Understand Intent

What does the user want in 3D? Understand:

1. **Subject** — Character, object, vehicle, creature, etc.
2. **Style** — Realistic, stylized, cartoon, game-ready?
3. **Reference** — Do they have an existing image, or should we generate one?

### Phase 2: Acquire Reference Image

Three paths:

**"I have a reference photo":**
- User provides an image path or URL
- Confirm it's suitable for 3D reference (clear subject, good lighting, no heavy occlusion)
- Proceed to Phase 3

**"Generate an image first":**
- Use the `generate_image` tool from gemini-media-mcp
- Craft a prompt focused on a single object with clear form and clean background
- Example: "A detailed fantasy sword with ornate handle, clean white background, product photography style, front view"
- Show the result, offer to regenerate if needed
- Proceed to Phase 3

**"I have multiple angles already":**
- Skip directly to Phase 4 with the user's images

### Phase 3: Generate Multi-Angle Views

Two approaches, offer the user a choice:

#### Approach A: Iterative (Recommended)

Generate each angle separately using `edit_image` from gemini-media-mcp, producing more consistent results:

1. **Front view** — Use the reference image as-is, or generate a clean front view
2. **Left side view** — `edit_image` with prompt: "Show this same [subject] from the left side, 90 degrees rotated, same style, same lighting, clean background"
3. **Back view** — `edit_image` with prompt: "Show this same [subject] from the back, 180 degrees rotated, same style, same lighting, clean background"
4. **Right side view** — `edit_image` with prompt: "Show this same [subject] from the right side, 270 degrees rotated, same style, same lighting, clean background"

Show each generated view to the user before proceeding. If a view looks wrong, regenerate it.

#### Approach B: Character Sheet (Single Generation)

Generate a 2x2 grid of views in one shot using `generate_image`:

Prompt template:
```
Character sheet of [subject description], 2x2 grid showing FRONT view (top-left), LEFT SIDE view (top-right), BACK view (bottom-left), RIGHT SIDE view (bottom-right). Clean white background, consistent lighting, same proportions in all views. Professional reference sheet style.
```

After generation, the individual views need to be cropped from the grid. If ImageMagick or similar is available, use `convert -crop` to split the 2x2 grid into 4 images. Otherwise, ask the user to crop manually.

**Note:** Iterative is recommended because character sheets sometimes have inconsistent proportions between views.

### Phase 4: 3D Reconstruction

Feed the generated views to trident-mcp:

1. Call `multiview_to_3d` with the 2-4 angle images (paths or URLs)
2. Use the latest model version (v3.1) for best quality
3. Poll `task_status` until complete — this typically takes 1-3 minutes
4. Download the result with `download_model` in its actual task output format

If the user needs FBX, OBJ, STL, USDZ, or 3MF instead of the task's native output, run `convert_format` first and then download the conversion task.

### Phase 5: Post-Processing

After the 3D model is ready, offer:

> "Multi-view 3D model generated! Saved to: [path]. What would you like to do?"
> 1. **Retopologize** — Clean lowpoly version for game use
> 2. **Convert format** — Export to FBX, OBJ, STL, USDZ, 3MF
> 3. **Stylize** — Apply LEGO, voxel, Voronoi, or Minecraft style
> 4. **Send to Blender** — Import via the preferred Blender connector (official first, community fallback; use the 3d-to-blender skill)
> 5. **Regenerate views** — Try different angles for better geometry
> 6. **Done** — Keep this model

## Tips for Best Results

- **Clean backgrounds**: White or neutral backgrounds produce better 3D results
- **Consistent lighting**: All views should have similar lighting direction
- **Single subject**: Multi-view works best with a single, centered object
- **Crisp details**: Higher quality reference images produce better geometry
- **2 views minimum, 4 ideal**: Front + side is minimum; front + left + back + right is ideal
