# Models and costs

Prices are in Tripo credits (1 credit = US$0.01), from Tripo's pricing page as of 2026-10-17. A task's `credits_consumed` is the authoritative cost. Credits are frozen when a task starts and refunded if it fails. `trident:list_models` lists the model versions the server knows.

## 3D generation

| Task | H3.1 / H3.0 / v2.5 (no texture / texture) | P1 (no texture / texture) | P2 (no texture / texture) |
| --- | --- | --- | --- |
| `text_to_model` | 10 / 20 | 30 / 40 | 100 / 110 |
| `image_to_model` | 20 / 30 | 40 / 50 | 100 / 110 |
| `multiview_to_model` | 20 / 30 | 40 / 50 | 100 / 110 |

Add-ons:

- `texture_quality: "detailed"` (HD): +10.
- `"extreme"` (8K): +20.
- `geometry_quality: "detailed"`: +20 (H3.x only).
- `quad`: +5.
- `smart_low_poly`: +10.
- `generate_parts`: +20.

On P2, detailed texture costs 120 in total and extreme 130.

Typical times are about 1-3 minutes for H3.1, and a little faster for P1.

## Images

| Model | ≤1K / 2K / 4K |
| --- | --- |
| `seedream_v4`, `seedream_v5` | 5 / 5 / 5 |
| `banana` | 5 (≤1K only) |
| `banana2` | 10 / 10 / 15 |
| `banana_pro` | 15 / 15 / 20 |
| `chat_image_2`, `chat_image_2.5_*` | 10 / 10 / 15 (`quality: "high"` 20 / 25 / 40) |

- `image_to_multiview` costs 10.
- `edit_multiview` costs 5 per edited view.
- `image_to_splat` costs 30.

## Processing

| Task | Credits |
| --- | --- |
| `texture_model` fast / standard / detailed / extreme | 10 / 10 / 20 / 30 |
| `convert_model` basic / advanced (any of `quad`, `face_limit`, `flatten_bottom*`, `texture_size`, `texture_format`, `pivot_to_center_bottom`, `scale_factor`) | 5 / 10 |
| `retopologize` v3.0 triangles / v3.0 quads / v2.0 / v1.0 | 35 / 40 / 30 / 10 |
| `smart_uv` | 20 |
| `segment_mesh` | 40 |
| `complete_mesh` ai_completion / quick_cap | 50 / 30 |
| `smart_segment` image / model | 85 / 55 |
| `rig_check` | free |
| `rig_model` | 25 |
| `retarget_animation` | 10 per animation |

`import_model` has no published price.

## Choosing a version

- **3D generation:**
  - H3.1 is the default for quality.
  - P1 is for game-ready topology on a face budget.
  - P2 (preview) is for quads or budgets up to 50k faces.
  - v2.5 and H3.0 are only for reproducing older results.
- **Texture:**
  - v3.0 is the default.
  - v3.5 is the newest. It adds `fast` quality and `delight`, which removes baked-in lighting.
- **Rig:**
  - v1.0 is the default, bipeds only.
  - v2.5 handles all creature types but has no humanoid presets.
  - v3.0 is best for humanoids and adds `rig_type: "others"` and the unity/unreal/vrm/actorcore specs.
- **Segmentation:**
  - v1.0 is geometric.
  - v2.0 (beta) is semantic, with granularity and reference-image control.
- **Images:**
  - `seedream_v4` and `seedream_v5` are cheap and good for 3D references.
  - Use `banana_pro` or `chat_image_2.5_*` for harder prompts.
  - Use `transparent` backgrounds only on the 2.5 chat models.
