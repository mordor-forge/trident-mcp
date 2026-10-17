# trident-mcp

An MCP server that gives agents the whole Tripo 3D API: generating, preparing and processing 3D assets as asynchronous Tripo tasks.

## Language

### Work

**Operation**:
One credit-spending Tripo endpoint that creates a Task, such as text-to-model, texture or rig.
_Avoid_: job, request, tool (a tool is the MCP face of an Operation)

**Task**:
Tripo's record of one Operation run, identified by a task ID, that moves from queued through running to a terminal Status.
_Avoid_: job, generation

**Status**:
Where a Task is in its lifecycle. The terminal Statuses are success, failed, cancelled, banned and expired.

**Output**:
The result fields of a successful Task, mostly short-lived signed URLs plus verdicts such as riggable.
_Avoid_: result, response

**Artifact**:
One Output URL saved as a local file, named by the Task and the Output field it came from.
_Avoid_: download, asset file

### Inputs

**Input**:
A reference to something an Operation consumes: a local file path, a public URL, a file token or a task ID.
_Avoid_: source, image (when the Input could be a model)

**File token**:
Tripo's handle for a file uploaded to it, accepted anywhere an Input is.

**View**:
One of the four fixed camera slots (front, left, back, right) in a multiview image set.
_Avoid_: angle, side

### Models

**Model version**:
A named Tripo AI model that an Operation runs on, such as v3.1-20260211, P1-20260311 or v3.0-20260909 for rigging. The API calls this field `model`.
_Avoid_: model (alone, in prose), engine

**Catalog**:
The server's built-in list of known Model versions, with their aliases, defaults and the parameters each one rejects.
_Avoid_: model list, registry

**Mesh**:
The 3D geometry an Operation produces or consumes, whether textured, rigged or animated.
_Avoid_: model (alone, in prose), object

**Rig**:
A skeleton bound to a Mesh so that it can be animated. Retargeting applies preset animations to a rigged Mesh.

### Account

**Credits**:
Tripo's billing unit (1 credit = US$0.01). Credits are frozen when a Task is created, charged on success and refunded on failure.
