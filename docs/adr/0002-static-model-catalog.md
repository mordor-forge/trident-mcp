---
status: accepted
date: 2026-07-08
applies_to: trident-mcp
---

# 0002 - Keep The Model Catalog Static

## Context

Tripo model availability can change independently of this server. MCP clients benefit from predictable tool schemas and predictable model aliases. Unit tests should not depend on live network discovery.

## Decision

Expose `list_models` from a static compatibility catalog in `internal/provider/tripo/models.go`. Normalize friendly aliases such as `v3.1`, `h3.1`, and `p1` locally.

## Consequences

- Tests remain deterministic.
- Agents can reason over a stable model matrix.
- The catalog must be manually refreshed when Tripo changes supported models.
- The catalog describes server-supported compatibility, not guaranteed live account entitlement.
