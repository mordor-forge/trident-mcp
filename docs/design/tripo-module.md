# Tripo Module Design

## Intent

`internal/tripo` is the single place that knows Tripo: the HTTP transport, Operations, Tasks, Inputs, the Catalog, files and account. `internal/server` turns it into MCP tools without adding Tripo knowledge of its own. Vocabulary follows [GLOSSARY.md](../../GLOSSARY.md).

## Invariants

- **Field names.** Request structs' `json` tags are both the MCP argument names and Tripo's payload keys. The only translations are explicit `wire` functions on an Operation (today only multiview `inputs`).
- **Local first.** Every local check runs before any upload or credit-spending call: the schema, required fields, cross-field `check`s, and catalog rejections.
- **Retries.** Only HTTP 429 is retried, honouring `Retry-After`. 5xx responses and timeouts on create calls are never retried, because Tripo may already have billed them.
- **Model versions.** A request's `model` is resolved per Operation. An alias becomes its ID, a missing model becomes the Operation's `DefaultFor` version, and fields in the version's `Rejects` are dropped when zero and refused when set. Unknown versions are sent as-is with a warning.
- **Inputs.** Local-path Inputs are uploaded once per call, and only for image and mesh extensions. URLs, file tokens and task IDs pass through untouched.
- **Task IDs.** IDs used in URL paths must match `^[A-Za-z0-9_-]{1,128}$`.
- **Outputs.** `Task.Output` stays Tripo's raw map. Every string field ending in `_url` (nested ones too) is a downloadable Artifact named after the field.
- **Downloads.** `Download` re-fetches the Task first, requires `success`, writes `<dir>/<task_id>/<name>.<ext>` through a temporary file, and never leaves partial files behind.
- **Errors.** Tripo errors surface as `*tripo.Error` with code, message, suggestion and request ID.

## Update rules

- **New endpoint:** add a request type and an `Operation` in `operations.go`, add one `addOperation` line in `internal/server/server.go`, add a row to `docs/MCP_TOOLS.md` (a test enforces this), and add a table case in both test suites.
- **New parameter:** add one field with `json`, `jsonschema` and, if needed, `enum` tags. Add a `check` only for documented cross-field rules.
- **New model version:** add one `catalog` entry with aliases, `DefaultFor` and `Rejects`, and bump `CatalogDate`.
- **Retirements:** remove the Operation or catalog entry, and list it under "Not exposed" in `docs/MCP_TOOLS.md`.
- **Lifecycle, upload, download or error semantics:** update this note and the relevant ADR.
