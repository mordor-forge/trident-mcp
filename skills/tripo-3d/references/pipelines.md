# Pipelines

Recipes by use. Each line is one task. Wait for it with `trident:wait_for_task`, check the result, and pass its `task_id` on as the next `input`. Credits are approximate; see [models-and-costs.md](models-and-costs.md).

## Game-ready prop (Unity, Unreal, Godot)

1. `text_to_model` or `image_to_model`: `model: "P1-20260311"`, `face_limit` around 1,500-8,000 for props. 40-50 credits.
2. Optional, only if the P1 topology is not clean enough: `retopologize` with `quad` and `face_limit`. 30 credits.
3. `convert_model`:
   - Godot, or engines with glTF import: `format: "GLTF"`.
   - Unity or Unreal: `format: "FBX"`.
   - Set `pivot_to_center_bottom: true`, plus `export_orientation` if the engine needs it.

   5-10 credits.

## Animated character

1. `text_to_image`: the character in a T-pose, `template: "t_pose"`, plain background. 5 credits. Check the pose: arms clear of the body, legs apart.
2. `image_to_model` with the image's task ID as `input`. Use `model: "P1-20260311"` for real-time or H3.1 for renders. 30-50 credits.
3. `rig_check`. Free. Stop if `riggable` is false and suggest a cleaner pose.
4. `rig_model`:
   - Humanoids: `model: "v3.0-20260909"`.
   - Animals and creatures: `model: "v2.5-20260210"`.
   - Set `rig_type` from `rig_check`.
   - For an engine skeleton, set `spec` to `unity`, `unreal`, `mixamo` or `vrm`.
   - Set `out_format: "fbx"` for engines.

   25 credits.
5. `retarget_animation` with the rig task ID, `animations` (up to 5, e.g. `preset:idle`, `preset:walk`, `preset:run`), and `out_format`. 10 credits per animation.
   - v3.0 rigs take humanoid, serpentine and aquatic presets.
   - v2.5 rigs take creature presets such as `preset:quadruped:walk`.
6. If the engine needs it, `convert_model` with `with_animation: true`, `animate_in_place` for root motion, and the matching `fbx_preset`.

## 3D print

1. `text_to_model` or `image_to_model` with `model: "v3.1-20260211"`, `geometry_quality: "detailed"`, `texture: false, pbr: false` (or 3MF for one color). 30-40 credits.
2. `convert_model` with `format: "STL"` (or `"3MF"`), `flatten_bottom: true`, `pivot_to_center_bottom: true`, and `scale_factor` if a size was given. 10 credits.

## AR (Apple Quick Look) and web

- **Web:** keep the GLB from generation.
- **AR Quick Look:** `convert_model` with `format: "USDZ"`.
- **Real-time budgets:** generate with P1, or `retopologize` first.

## High-fidelity character or object from several views

1. Get a clean front image, either the user's or one from `text_to_image`.
2. `image_to_multiview` with it as `input`. 10 credits.
3. Check the four `*_view` files. Fix bad ones with `edit_multiview`: `prompts: [{view: "back", prompt: "..."}]`, 5 credits per view.
4. `multiview_to_model` with the multiview or edit task ID as `input`. 30-50 credits.

## Retexture an existing mesh

1. `import_model` with the mesh path. Skip this if it is already a Tripo task.
2. `texture_model`:
   - `texture_prompt`: one of `{text}`, `{image}` or `{images: [front, left, back, right]}`, optionally `style_image` with text.
   - `model: "v3.5-20260815"` for the newest texturer and `texture_quality: "fast"`, or `"detailed"` for HD.
   - `part_names` to texture only some parts.

   10-30 credits.

## Editable parts

- **Geometry parts at generation time:** `generate_parts: true` with `texture: false, pbr: false`. +20 credits.
- **After the fact:** `segment_mesh` (40 credits; `model: "v2.0-20260430"` for semantic parts), then `complete_mesh` with `completion_mode: "ai_completion"` to rebuild hidden geometry, or `"quick_cap"` to close holes cheaply.
- **Guided by a description:** `smart_segment` with `seg_type: "image"` and a `hint` such as "wheels, body, windows". 85 credits.

## UVs and clean topology for hand editing

- `retopologize` with `model: "v3.0"`, `quad: true` and `smart_uv: true` in one task (35-40 credits).
- Or run `smart_uv` alone on a mesh under about 80k triangles (20 credits). Its `uv_render` output shows the layout.
