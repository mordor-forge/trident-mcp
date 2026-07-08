## Summary

Describe the user-visible change and the affected MCP tools, provider code, docs, or skills.

## Verification

- [ ] `go build ./cmd/trident-mcp`
- [ ] `go test ./... -count=1`
- [ ] `go vet ./...`
- [ ] `golangci-lint run`
- [ ] `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- [ ] Mermaid diagrams validated with `mmdc` when docs diagrams changed
- [ ] Live E2E intentionally skipped or run with command and credit impact noted

## Agent Readiness

- [ ] `AGENTS.md` updated when workflow, boundaries, or verification changed
- [ ] ADR/design docs updated when architecture or invariants changed
- [ ] MCP tool docs updated when tool schemas or behavior changed

## Risk

List API compatibility risks, credit-spending risks, and follow-up work.
