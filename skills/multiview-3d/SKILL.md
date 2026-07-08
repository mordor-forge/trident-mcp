---
name: multiview-3d
description: Use when the user wants a high-quality 3D model from multiple views, a reference photo expanded into front/side/back views, Tripo multiview reconstruction, or a multi-angle image-to-3D pipeline.
---

# Multi-View 3D Pipeline Skill

You are an expert at orchestrating the multi-angle 3D generation pipeline. This skill uses **trident-mcp** for Tripo multiview generation/reconstruction, and can optionally use **gemini-media-mcp** for more controlled reference-image creation or iterative angle edits.

This is a composed workflow skill. `trident-mcp` itself remains independently useful in any MCP-capable client. Pair it with [`gemini-media-mcp`](https://github.com/mordor-forge/gemini-media-mcp) when the workflow needs richer image ideation or manual control over each generated angle.

## Prerequisites

Minimum:

- **trident-mcp** — Needs `multiview_to_3d`. Prefer also having `get_balance`, `task_status`, `download_model`, `image_to_multiview`, and `edit_multiview`.

Optional:

- **gemini-media-mcp** — Useful for custom multi-angle reference generation (needs `generate_image` and `edit_image`). Repository: <https://github.com/mordor-forge/gemini-media-mcp>

Before starting, verify the available tools. If `multiview_to_3d` is missing, tell the user to configure `trident-mcp`. If only Gemini tools are missing, use Tripo's native `image_to_multiview` route when possible.

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
- Use `text_to_image` from trident-mcp or `generate_image` from gemini-media-mcp
- Craft a prompt focused on a single object with clear form and clean background
- Example: "A detailed fantasy sword with ornate handle, clean white background, product photography style, front view"
- If using `text_to_image`, poll the task and use the completed `output.imageUrl` as the source image; do not pass the image task ID as an image input.
- Show the result, offer to regenerate if needed
- Proceed to Phase 3

**"I have multiple angles already":**
- Skip directly to Phase 4 with the user's images

### Phase 3: Generate Multi-Angle Views

Two approaches, offer the user a choice:

#### Approach A: Native Tripo Multiview (Recommended When Available)

Use `image_to_multiview` from trident-mcp to create ordered multi-angle views from a single reference image. If a generated view is wrong, use `edit_multiview` with per-view `prompts` entries for targeted corrections before reconstruction.

This is the shortest path when the source image is already strong and the user does not need manual control over every intermediate angle.

After `image_to_multiview` or `edit_multiview` succeeds, prefer passing that successful task ID to `multiview_to_3d` as `taskId`. Extract individual view URLs only if the tool version lacks task-id reuse.

#### Approach B: Iterative Gemini Views

Generate each angle separately using `edit_image` from gemini-media-mcp, producing more consistent results:

1. **Front view** — Use the reference image as-is, or generate a clean front view
2. **Left side view** — `edit_image` with prompt: "Show this same [subject] from the left side, 90 degrees rotated, same style, same lighting, clean background"
3. **Back view** — `edit_image` with prompt: "Show this same [subject] from the back, 180 degrees rotated, same style, same lighting, clean background"
4. **Right side view** — `edit_image` with prompt: "Show this same [subject] from the right side, 270 degrees rotated, same style, same lighting, clean background"

Show each generated view to the user before proceeding. If a view looks wrong, regenerate it.

#### Approach C: Character Sheet (Single Generation)

Generate a 2x2 grid of views in one shot using `generate_image`:

Prompt template:
```
Character sheet of [subject description], 2x2 grid showing FRONT view (top-left), LEFT SIDE view (top-right), BACK view (bottom-left), RIGHT SIDE view (bottom-right). Clean white background, consistent lighting, same proportions in all views. Professional reference sheet style.
```

After generation, the individual views need to be cropped from the grid. If ImageMagick or similar is available, use `convert -crop` to split the 2x2 grid into 4 images. Otherwise, ask the user to crop manually.

**Note:** Iterative is recommended because character sheets sometimes have inconsistent proportions between views.

### Phase 4: 3D Reconstruction

Feed the generated views to trident-mcp:

1. Run `get_balance` before credit-spending work when a live Tripo key is configured.
2. Call `multiview_to_3d` with either the successful multiview `taskId`, or 2-4 angle images (paths/URLs) in front/left/back/right order.
3. Use `v3.1-20260211` for best quality, or `P1-20260311` when the user needs low-poly/game-ready topology.
4. Use options such as `quad`, `smartLowPoly`, `generateParts`, `faceLimit`, `textureAlignment`, and `enableImageAutofix` when topology or texture alignment matters. For P1, omit `smartLowPoly` because P1 is already low-poly; use `faceLimit` instead.
5. Poll `task_status` until complete — this typically takes 1-3 minutes.
6. Download the result with `download_model` in its actual task output format.

If the user needs FBX, OBJ, STL, USDZ, or 3MF instead of the task's native output, run `convert_format` first and then download the conversion task.

### Phase 5: Post-Processing

After the 3D model is ready, offer:

> "Multi-view 3D model generated! Saved to: [path]. What would you like to do?"
> 1. **Retopologize** — Clean lowpoly version for game use
> 2. **Convert format** — Export to FBX, OBJ, STL, USDZ, 3MF
> 3. **Texture/refine/segment** — Improve texture, refine quality, or split parts
> 4. **Rig/animate** — Check riggability, rig, or retarget animation
> 5. **Stylize** — Apply LEGO, voxel, Voronoi, or Minecraft style
> 6. **Send to Blender** — Import via the preferred Blender connector (official first, community fallback; use the 3d-to-blender skill)
> 7. **Regenerate views** — Try different angles for better geometry
> 8. **Done** — Keep this model

## Tips for Best Results

- **Clean backgrounds**: White or neutral backgrounds produce better 3D results
- **Consistent lighting**: All views should have similar lighting direction
- **Single subject**: Multi-view works best with a single, centered object
- **Crisp details**: Higher quality reference images produce better geometry
- **2 views minimum, 4 ideal**: Front + side is minimum; front + left + back + right is ideal
