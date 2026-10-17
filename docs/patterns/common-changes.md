# Common Change Patterns

Worked examples for the recipes in [AGENTS.md](../../AGENTS.md#change-recipes).

## Add a parameter Tripo just shipped

Say Tripo adds `symmetry_axis` (`x`, `y` or `z`) to `/models/convert`. Add one field to `ConvertModelRequest` in `internal/tripo/operations.go`:

```go
SymmetryAxis string `json:"symmetry_axis,omitempty" enum:"x,y,z" jsonschema:"Axis to mirror across when force_symmetry is set"`
```

From that one field:

- The MCP schema shows the field with its enum and description.
- The SDK and `Submit` both validate it.
- It is sent to Tripo as `symmetry_axis`.

To finish, add the key to the `convert_model` case in `TestOperations_SendTripoPayloads`.

## Add an endpoint

Say Tripo ships `POST /mesh/repair` that returns a mesh:

```go
// internal/tripo/operations.go
RepairMesh = Operation[RepairMeshRequest]{
	Name: "repair_mesh", Path: "/mesh/repair", Produces: ProducesMesh,
	Description: "Close holes and fix non-manifold geometry in a mesh.",
}

type RepairMeshRequest struct {
	Input Input `json:"input" jsonschema:"Task ID of the mesh (or a file token / URL)"`
}
```

Then:

1. Add `addOperation(s, tripo.RepairMesh)` in `internal/server/server.go`.
2. Add `repair_mesh` to `operationTools` in `internal/server/server_test.go`, with a `TestOperationTools_ReachTripo` case.
3. Add a case to `TestOperations_SendTripoPayloads`.
4. Add a row to `docs/MCP_TOOLS.md`.

If the endpoint takes a `model`, add catalog entries with `DefaultFor: []string{"repair_mesh"}`. `TestCatalog_DefaultsAndFilters` lists every Operation that needs a default.

## Add a model version

Say Tripo ships P3. Add one entry to `catalog` in `internal/tripo/catalog.go`:

```go
{ID: "P3-20270115", Name: "Tripo P3", Family: "3d", Aliases: []string{"p3", "tripo-p3"}, Stage: "preview",
	Rejects:     []string{"smart_low_poly", "generate_parts", "geometry_quality"},
	Description: "…"},
```

Then:

1. Bump `CatalogDate`.
2. If it becomes the default, move `DefaultFor` to it.
3. Update the model guidance in `skills/tripo-3d/references/models-and-costs.md`.
4. Run the spending e2e if the request shape changed.

## Retire something

Delete the Operation or catalog entry. Then list it under "Not exposed" in `docs/MCP_TOOLS.md` with the reason, so agents stop looking for it.
