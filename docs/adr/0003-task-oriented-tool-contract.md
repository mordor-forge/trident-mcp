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
