# Security Policy

## Supported Versions

Security fixes are expected on the default branch and tagged releases derived from it.

## Reporting Issues

Do not open public issues for secrets, credential exposure, or exploitable vulnerabilities. Report privately to the repository maintainer through the preferred private channel for the project.

Include:

- affected commit or release
- reproduction steps
- expected impact
- whether a live Tripo key or generated artifact is involved
- any suggested mitigation

## Local Security Checks

Run:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

CI also runs `govulncheck` in advisory mode.

## Scope

See [THREAT_MODEL.md](THREAT_MODEL.md) for assets, trust boundaries, known threats, and recommended mitigations.
