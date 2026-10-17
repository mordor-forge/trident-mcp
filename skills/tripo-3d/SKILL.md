---
name: tripo-3d
description: Produces 3D assets with the Tripo 3D API through the trident MCP server - meshes from text, one image or several views; game-ready low-poly (P1/P2), retopology and UVs; textures; rigging and animation; segmentation into parts; 3D-print, AR and engine exports (GLB, FBX, OBJ, STL, USDZ, 3MF); Gaussian splats. Prices the plan in Tripo credits before spending. Use when a request needs a 3D model, mesh, character, prop or asset made, changed or exported, even if the user only says "make me a 3D ..." or "turn this image into 3D".
license: Apache-2.0
compatibility: Requires the trident MCP server (https://github.com/mordor-forge/trident-mcp) with a Tripo API key.
metadata:
  author: mordor-forge
  version: "0.2.0"
  mcp-server: trident
---

# Tripo 3D

Turns a request for a 3D asset into finished files on disk.

## Tools

The tools live on the `trident` MCP server. Harnesses name them differently (`mcp__plugin_trident_trident__text_to_model`, `trident.text_to_model`, or plain `text_to_model`), so match on the part after the server name. This skill writes them as `trident:<tool>`.

- **Tool contract.** Every task-creating tool returns a `task_id` at once. `trident:wait_for_task` then waits for the task and saves its outputs to `<output dir>/<task_id>/<output>.<ext>`, e.g. `model.glb`, `rendered_image.webp`, `generated_image.png`, `front_view.png`.
- **Chaining.** Pass a `task_id` as the next tool's `input`.
- **Inputs.** `input` also takes an absolute local path, which is uploaded automatically, as well as a URL or a file token.
- **Arguments.** They are Tripo's own field names. Each tool's schema lists the allowed values.

## Steps

1. **Brief.** Settle four things:
   - *subject*: one object;
   - *source*: text, one image, several views, or an existing mesh;
   - *use*: render or preview, game engine, animation, 3D print, AR, or editing in a DCC tool;
   - *quality bar*: draft or final.

   Ask at most two questions, and default the rest from the request. Done when each of the four has a value.
2. **Plan.** Pick the route and model version from [Routes](#routes). For any use beyond a plain mesh, follow the matching recipe in [references/pipelines.md](references/pipelines.md). Write the plan as a numbered list of tool calls, one per task.
3. **Price.** Price each step with [references/models-and-costs.md](references/models-and-costs.md), total it, and compare with `trident:get_balance`. Get the user's approval when the total passes 50 credits or the user has not okayed spending. Done when the plan is approved and the balance covers it.
4. **Run.** Call the first tool, then `trident:wait_for_task` with its `task_id`.
   - `done: false`: call `wait_for_task` again.
   - Failure (`failed`, `banned`, `cancelled`, `expired`): report `error_code` and `error_message`, change one thing (prompt, input image or model version), and retry once.
   - Independent steps can run side by side, but only one image task at a time; Tripo allows one per account.

   Done when the step's task is `success` and its files are saved.
5. **Check.** After each mesh or image step, open its `rendered_image` or `generated_image` file. If your harness can view images, read the file. Compare it against the brief: right object, complete, clean silhouette, no fused extras. Fix a bad result at this step, before spending credits on follow-ups that would inherit the defect.
6. **Finish.** Run the remaining plan steps the same way, ending with `trident:convert_model` when the use needs a specific format. Set `export_orientation`, scale and pivot on that final convert. They do not carry over from earlier steps.
7. **Report.** Give every saved file's absolute path and format, the `task_id` of the last step (so later edits can chain from it), and the credits spent (the tasks' `credits_consumed`, or `trident:get_usage`). Other tools such as Blender, game engines and slicers can import these paths directly.

## Routes

| Source | Route |
| --- | --- |
| Text only | `trident:text_to_model`. When the shape matters, use `trident:text_to_image` first, then `trident:image_to_model` with that image's `task_id` as `input`. You see the design before paying for 3D, and the result follows it. |
| One good image | `trident:image_to_model`. For a messy photo (clutter, odd angle, several objects), run `trident:image_to_image` with `template: "3d_enhance"` first. |
| Back or sides matter, or a character | `trident:image_to_multiview` from the front image. Fix any wrong view with `trident:edit_multiview`, then run `trident:multiview_to_model` with the multiview `task_id` as `input`. Own photos go in `inputs: {front, left, back, right}`; front plus any one more view is enough. |
| Existing mesh file | `trident:import_model`, then chain from its `task_id`. |
| Scene-like capture | `trident:image_to_splat`, which saves a `.splat` file, not a mesh. |

Model versions (`trident:list_models` has the full list):

| Need | Version |
| --- | --- |
| Best detail, renders, print | `v3.1-20260211` (H3.1, the default). Use `geometry_quality: "detailed"` for the final take. |
| Game-ready or real-time | `P1-20260311`, with `face_limit` as the polygon budget (50-20,000). P1 takes no `quad`, `smart_low_poly`, `generate_parts` or `geometry_quality`. |
| Low poly with quads or up to 50k faces | `P2-20260801` (preview, about 3x the price of P1). |
| Cheapest look at the shape | Any version with `texture: false, pbr: false`. Texture later with `trident:texture_model`. |

Prompts for `text_to_model` and `text_to_image`:

- Describe one object: its shape, its materials, and its style (stylized, realistic, low-poly).
- Use 20-60 words.
- For image references, ask for a plain background and the whole object in frame.
- For characters to be rigged, ask for a T-pose or A-pose, or use `template: "t_pose"`.
- Put exclusions in `negative_prompt` for 3D, or after `--no` for images.

## Rules

- **Outputs.** Use `wait_for_task` or `trident:download_task` to save outputs; `get_task` URLs are signed and expire. Downloads keep the task's format. `trident:convert_model` makes other formats.
- **Local validation.** An argument error comes back before any credits are spent. Read it, fix that argument, and retry.
- **Parts.** `generate_parts` needs `texture: false` and `pbr: false`. For parts with textures, generate normally, then use `trident:segment_mesh` and `trident:complete_mesh`.
- **Rigging.** Run `trident:rig_check` first. It is free and returns `riggable` and `rig_type`. Then run `trident:rig_model`: `v3.0-20260909` for humanoids, `v2.5-20260210` for animals and creatures. Retargeting takes preset names like `preset:walk`.
- **Unknown versions.** A warning that a model version is "not in the built-in catalog" means it was sent to Tripo anyway. If Tripo then rejects it, use a version from `trident:list_models`.
