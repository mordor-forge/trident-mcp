# Provider Boundary Design

## Intent

`trident-mcp` keeps MCP tool registration and Tripo HTTP details separated. The server layer depends on provider capability interfaces, while the Tripo adapter owns request translation, enum normalization, upload handling, task polling, and download behavior.

## Preconditions

- Runtime config must provide `TRIPO_API_KEY` before constructing the provider.
- `MODEL_OUTPUT_DIR` is created automatically if missing before downloads are attempted.
- Server tool registration must receive nil for unsupported capability families.
- Tripo task IDs used in URL paths must pass `validateTaskID`.
- Credit-spending live tests must check balance first and require explicit opt-in flags.

## Invariants

- MCP request fields stay camelCase; Tripo payload fields are translated to snake_case only in the provider layer.
- All task-creating provider methods return `provider.ModelOperation` with a task ID and initial status.
- The server only registers tools when their backing capability interface exists.
- `download_model` never converts formats. Format conversion is a separate async task.
- `list_models` is deterministic and static; it does not call Tripo live discovery.
- P1 requests omit unsupported `smart_low_poly`; P1 topology is controlled with `faceLimit`.
- `multiview_to_3d` accepts either ordered image paths/URLs or a reusable multiview `taskId`, never a mix.

## Rationale

Capability interfaces make the MCP surface testable without live network calls and keep future providers from needing to mimic every Tripo feature at once. Static model listing makes client behavior predictable and avoids flaky tests caused by remote catalog drift. Async task handling is explicit because Tripo's API is task-oriented across generation, processing, mesh, and animation workflows.

## Update Rules

Update this design note when changing:

- provider capability interfaces
- task lifecycle semantics
- model catalog behavior
- file upload or download behavior
- tool registration rules
- cross-tool input contracts such as multiview task reuse
