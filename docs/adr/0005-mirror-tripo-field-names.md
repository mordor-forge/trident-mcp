---
status: accepted
date: 2026-10-17
applies_to: trident-mcp
---

# 0005 - MCP arguments mirror Tripo's snake_case field names

## Context

Before this decision, MCP arguments were camelCase and every Tripo field had to be translated by hand into a snake_case payload map. That meant three names for each parameter: the MCP argument, the Go field and the payload key. The hand-built maps had already drifted, for example `targetFaces` versus `faceLimit`, and `originalTaskId` versus `input`.

## Decision

MCP tool arguments use Tripo's own field names (`face_limit`, `texture_quality`, `model`, `input`). A request struct's `json` tags are the MCP schema and the Tripo payload at once. An `enum` tag on the same field feeds the schema enum and the local validation. The only fields that are not Tripo's are a few server-side conveniences, such as multiview `inputs` given as a view map, which are translated into Tripo's wire form inside the Operation.

## Consequences

- When Tripo adds a parameter, we add one struct field. Tripo's docs apply to agents verbatim.
- Tool names and arguments are consistently snake_case.
- This was a breaking change for clients written against the earlier camelCase tool contract.
