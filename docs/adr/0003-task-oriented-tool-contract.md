---
status: accepted
date: 2026-07-08
applies_to: trident-mcp
---

# 0003 - Use A Task-Oriented MCP Tool Contract

## Context

Tripo v3 exposes generation, image preparation, mesh processing, model processing, and animation as asynchronous tasks. MCP clients need a simple and repeatable workflow for starting tasks, polling, and downloading outputs.

## Decision

Model task-creating tools around `ModelOperation`, expose `task_status` / `get_task` and `get_tasks` for polling, and keep download/conversion as explicit follow-up operations.

## Consequences

- Agents can use one lifecycle pattern across most tools.
- Credit-spending operations are visible as task creation steps.
- `download_model` stays simple and does not hide conversion work.
- Synchronous or streaming providers would need an adapter layer to fit this contract.

## Amendment (2026-10-18)

The contract still holds: task-creating tools return a task ID at once. Three things changed:

- Polling is `get_task` and `get_tasks`. The duplicate `task_status` tool is gone.
- `wait_for_task` adds an opt-in blocking step. It polls with progress notifications and saves the outputs while their signed URLs are still valid.
- `download_model` became `download_task`. It saves every output of a task (meshes, images, views, splats), not only the mesh, and it still never converts formats.
