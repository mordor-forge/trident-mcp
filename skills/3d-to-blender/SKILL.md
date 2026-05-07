---
name: 3d-to-blender
description: Import 3D models into Blender and set up scenes. Prefer the official Blender connector / Blender Lab MCP server when available, and fall back to the community blender-mcp server if needed. Use when the user wants to import a generated 3D model into Blender, set up a scene, adjust materials, configure rendering, or do any Blender-related work with a trident-mcp generated model. Triggers on "import to Blender", "send to Blender", "open in Blender", "set up Blender scene", "render this model".
---

# 3D to Blender Skill

You are an expert at importing 3D models into Blender and setting up professional scenes. This skill should use the best available Blender connector after a model has been generated with trident-mcp.

## Prerequisites

This skill requires a Blender MCP server to be configured and connected to a running Blender instance.

Before starting, detect which Blender backend is available.

### Preferred backend: official Blender connector / Blender Lab MCP

Treat the official connector as available when tools like these are present:

- `execute_blender_code`
- `get_objects_summary`
- `get_blendfile_summary_datablocks`
- `get_screenshot_of_window_as_image`
- `render_viewport_to_path`

If both the official connector and the community server are present, always prefer the official connector.

### Fallback backend: community `blender-mcp`

Treat the community backend as available when tools like these are present:

- `execute_blender_code`
- `get_scene_info`
- `get_viewport_screenshot`

### If no Blender backend is available

Guide the user toward the official connector first:

> "Blender is not connected. Preferred setup is the official Blender connector:
> 1. Open Blender 5.1+
> 2. Install the Blender Lab MCP add-on
> 3. In the add-on preferences, confirm host `localhost` and port `9876`
> 4. Start or enable auto-start for the MCP bridge server
> 5. Make sure the official Blender MCP server is configured in Claude Code
>
> Note: in the official add-on, the controls live in the add-on preferences, not the N panel.
>
> If you already use the community `blender-mcp`, that can still work as a fallback." 

## Workflow

### Phase 1: Select Model

Identify which model to import:

- **From recent generation**: If we just generated a model in this conversation, use that file path
- **User provides path**: User specifies a path to a .glb, .fbx, .obj, or other supported format
- **From output directory**: List recent files in the model output directory

### Phase 2: Import Model

Use `execute_blender_code` to run the appropriate Blender import operator. This tool is available in both the official connector and the community backend, so it is the safest import path.

**GLTF/GLB:**
```python
bpy.ops.import_scene.gltf(filepath="/path/to/model.glb")
```

**FBX:**
```python
bpy.ops.import_scene.fbx(filepath="/path/to/model.fbx")
```

**OBJ:**
```python
bpy.ops.wm.obj_import(filepath="/path/to/model.obj")
```

**STL:**
```python
bpy.ops.wm.stl_import(filepath="/path/to/model.stl")
```

After import, select the imported object and center it at origin:
```python
import bpy
obj = bpy.context.selected_objects[0]
bpy.context.view_layer.objects.active = obj
bpy.ops.object.origin_set(type='ORIGIN_GEOMETRY', center='BOUNDS')
obj.location = (0, 0, 0)
```

### Phase 3: Scene Setup (Optional)

Ask the user: "Want me to set up a presentation scene?" If yes:

**Three-point lighting:**
```python
import bpy
import math

# Key light
key = bpy.data.lights.new(name="Key", type='AREA')
key.energy = 500
key_obj = bpy.data.objects.new("Key", key)
bpy.context.collection.objects.link(key_obj)
key_obj.location = (3, -3, 4)
key_obj.rotation_euler = (math.radians(60), 0, math.radians(45))

# Fill light
fill = bpy.data.lights.new(name="Fill", type='AREA')
fill.energy = 200
fill_obj = bpy.data.objects.new("Fill", fill)
bpy.context.collection.objects.link(fill_obj)
fill_obj.location = (-3, -2, 3)
fill_obj.rotation_euler = (math.radians(50), 0, math.radians(-45))

# Rim light
rim = bpy.data.lights.new(name="Rim", type='AREA')
rim.energy = 300
rim_obj = bpy.data.objects.new("Rim", rim)
bpy.context.collection.objects.link(rim_obj)
rim_obj.location = (0, 4, 3)
rim_obj.rotation_euler = (math.radians(120), 0, math.radians(180))
```

**Camera at 3/4 view:**
```python
cam = bpy.data.cameras.new("Camera")
cam_obj = bpy.data.objects.new("Camera", cam)
bpy.context.collection.objects.link(cam_obj)
cam_obj.location = (4, -4, 3)
cam_obj.rotation_euler = (math.radians(65), 0, math.radians(45))
bpy.context.scene.camera = cam_obj
```

**Ground plane:**
```python
bpy.ops.mesh.primitive_plane_add(size=10, location=(0, 0, 0))
plane = bpy.context.active_object
mat = bpy.data.materials.new(name="Ground")
mat.use_nodes = True
bsdf = mat.node_tree.nodes["Principled BSDF"]
bsdf.inputs["Base Color"].default_value = (0.8, 0.8, 0.8, 1)
bsdf.inputs["Roughness"].default_value = 0.9
plane.data.materials.append(mat)
```

### Phase 4: Verify

Verify using the best available inspection tool for the active backend:

- Official connector: prefer `get_screenshot_of_window_as_image` for a fast UI snapshot, and `render_viewport_to_path` when the user wants a cleaner viewport render.
- Community backend: use `get_viewport_screenshot`.

If screenshot tools are unavailable but Blender summary tools are available, use `get_objects_summary` or `get_blendfile_summary_datablocks` to sanity-check that the import succeeded.

### Phase 5: Interactive Review

> "Model imported and scene set up! What would you like to do?"
> 1. **Adjust placement** — Move, rotate, or scale the model
> 2. **Change materials** — Modify colors, roughness, metallic properties
> 3. **Adjust lighting** — Change light positions, intensity, or color
> 4. **Configure rendering** — Set up Cycles or EEVEE render settings
> 5. **Render** — Render a final image
> 6. **Done** — Scene is ready

## Tips

- Prefer the official Blender connector when both backends are available.
- When not troubleshooting, refer to the active backend generically as the "Blender connector" rather than naming the MCP implementation.
- Always import at origin and reset transforms for clean scene setup
- Use EEVEE for quick previews, Cycles for final renders
- The three-point lighting setup works well for most product-style renders
- If the model appears too large or small, scale uniformly (S key in Blender, or via script)
