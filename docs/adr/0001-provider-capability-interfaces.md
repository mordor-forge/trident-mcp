---
status: accepted
date: 2026-07-08
applies_to: trident-mcp
---

# 0001 - Use Provider Capability Interfaces

## Context

The server exposes many MCP tools, but not every backend will support every tool family. Tripo currently supports generation, image prep, model processing, mesh processing, account APIs, and animation. Tests also need to instantiate the server with partial mock providers.

## Decision

Keep provider behavior behind focused capability interfaces in `internal/provider/interfaces.go`. The server conditionally registers MCP tools only when the matching interface is non-nil.

## Consequences

- The MCP server layer can be tested without a concrete Tripo provider.
- New backends can implement capability families incrementally.
- Tool registration logic must stay aligned with interface boundaries.
- Adding a new capability family requires touching provider interfaces, server registration, and tests.
