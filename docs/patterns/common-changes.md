# Common Change Patterns

These examples are written for agents making copy-modify changes. Keep this file in sync with `AGENTS.md` when common workflows change.

## New MCP Tool

Reference implementation:

- Pattern in `internal/server/tools_generation.go`
- Provider mapping in `internal/provider/tripo/generation.go`
- Request types in `internal/provider/types.go`
- Server tests in `internal/server/server_test.go`

Checklist:

1. Add the request/response shape to `internal/provider/types.go`.
2. Add a provider interface method only when a new capability family is needed.
3. Implement the concrete Tripo request in the provider package.
4. Register the MCP tool in the matching `internal/server/tools_*.go` file.
5. Add `httptest` coverage for the Tripo payload.
6. Add an MCP server test for registration and handler behavior.
7. Update [MCP_TOOLS.md](../MCP_TOOLS.md) and any affected skill.

## New Tripo v3 Processing Endpoint

Reference implementation:

- See `internal/provider/tripo/v3_features.go` for model, mesh, account, upload, and animation APIs.
- See `internal/server/tools_v3_processing.go` for matching MCP handlers.

Checklist:

1. Keep upstream task creation asynchronous and return `ModelOperation`.
2. Preserve the shared poll/download lifecycle.
3. Validate local enums when a bad value would otherwise spend credits.
4. Keep Tripo-specific snake_case inside the provider layer.
5. Add tests before implementation for request path, method, and body shape.

## Model Catalog Update

Reference implementation:

- Pattern in `internal/provider/tripo/models.go`
- Public docs in [MCP_TOOLS.md](../MCP_TOOLS.md#models)
- Agent defaults in `skills/3d-gen/SKILL.md`

Checklist:

1. Add the catalog entry with namespace, API version, capabilities, aliases, and default marker if applicable.
2. Add aliases only when they are unambiguous.
3. Let multiview support derive from the capability list.
4. Add tests for model resolution and `list_models`.
5. Update agent-facing docs when defaults or live caveats change.

## Documentation And Skill Updates

Reference implementation:

- README as the short navigation entry point.
- `docs/` for detailed reference.
- `skills/*/SKILL.md` for optional client-side workflow guidance.

Checklist:

1. Keep README concise and link to focused docs.
2. Update `AGENTS.md` for repository workflow, verification, and common patterns.
3. Update ADRs when changing architecture or lasting trade-offs.
4. Validate Mermaid diagrams with `mmdc`.
5. Run `agentready assess . -o .agentready` and keep generated reports local.
