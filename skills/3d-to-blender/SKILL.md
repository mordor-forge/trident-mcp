---
name: 3d-to-blender
description: Import 3D models into Blender via blender-mcp and set up scenes. Use when the user wants to import a generated 3D model into Blender, set up a scene, adjust materials, configure rendering, or do any Blender-related work with a trident-mcp generated model. Triggers on "import to Blender", "send to Blender", "open in Blender", "set up Blender scene", "render this model".
---

# 3D to Blender Skill

You are an expert at importing 3D models into Blender and setting up professional scenes. This skill uses **blender-mcp** (ahujasid/blender-mcp) to control Blender programmatically after a model has been generated with trident-mcp.

## Prerequisites

This skill requires **blender-mcp** to be configured and connected to a running Blender instance with the blender-mcp addon enabled.

Before starting, verify the connection by checking for blender-mcp tools (like `get_scene_info` or `execute_blender_code`). If not available, guide the user:

> "Blender MCP is not connected. To set it up:
> 1. Open Blender
> 2. Install the blender-mcp addon (Edit > Preferences > Add-ons)
> 3. Enable it and click 'Connect' in the addon panel
> 4. Make sure blender-mcp is configured in your Claude Code MCP settings"

## Workflow

### Phase 1: Select Model

Identify which model to import:

- **From recent generation**: If we just generated a model in this conversation, use that file path
- **User provides path**: User specifies a path to a .glb, .fbx, .obj, or other supported format
- **From output directory**: List recent files in the model output directory

### Phase 2: Import Model

Use `execute_blender_code` to run the appropriate Blender import operator:

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

Use `get_viewport_screenshot` to capture and show the user the current state of the scene.

### Phase 5: Interactive Review

> "Model imported and scene set up! What would you like to do?"
> 1. **Adjust placement** — Move, rotate, or scale the model
> 2. **Change materials** — Modify colors, roughness, metallic properties
> 3. **Adjust lighting** — Change light positions, intensity, or color
> 4. **Configure rendering** — Set up Cycles or EEVEE render settings
> 5. **Render** — Render a final image
> 6. **Done** — Scene is ready

## Tips

- Always import at origin and reset transforms for clean scene setup
- Use EEVEE for quick previews, Cycles for final renders
- The three-point lighting setup works well for most product-style renders
- If the model appears too large or small, scale uniformly (S key in Blender, or via script)
