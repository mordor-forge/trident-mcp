---
status: accepted
date: 2026-10-17
applies_to: trident-mcp
supersedes: 0001
---

# 0004 - Target Tripo only, behind one deep client module

## Context

ADR-0001 put provider capability interfaces in front of the MCP server so that other 3D backends (Meshy, Rodin, ...) could be added incrementally. Only one adapter was ever written. The result was 9 interfaces with 27 methods, each mapping 1:1 to a Tripo endpoint. Tool registration gates never varied, and the server mocks never checked arguments. The real logic, which is task lifecycle, input upload, catalog rules and error handling, was spread across seven files underneath that seam.

## Decision

trident-mcp targets Tripo only. All Tripo knowledge lives in one module, `internal/tripo`: the transport and error type, Operations, Tasks, Inputs, the Catalog, files and account. The MCP server depends on `*tripo.Client` directly and registers every Operation through one generic adapter.

The only seam is HTTP: `tripo.Options.BaseURL` and `HTTPClient`. Two adapters sit behind it: the live Tripo API, and the `tripotest` fake used by every test.

## Consequences

- Tests cross MCP → tripo → fake HTTP in one go, so argument decoding, payload encoding and output parsing are tested together.
- Adding a Tripo endpoint means one request type, one `Operation` value, and one line in the server registry.
- A second backend would need its own MCP server or a new seam. Composing several generators belongs in a pipeline layer above the MCP servers, not inside this one.
