#!/usr/bin/env bash
# Validates the Claude Code plugin packaging: manifests parse, the plugin wires
# the trident MCP server, skill metadata agrees with the plugin version, and -
# when the Claude Code CLI is available (or npx can fetch it) - the official
# validator passes.
#
# Usage: scripts/validate-packaging.sh [--expect-version X.Y.Z] [--offline]
#   --expect-version  fail unless every manifest carries this version (release CI passes the tag)
#   --offline         skip checks that download tools
set -uo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root" || exit 1

expect=""
offline=false
while [ $# -gt 0 ]; do
  case "$1" in
    --expect-version) expect=${2#v}; shift 2 ;;
    --offline) offline=true; shift ;;
    -h | --help) sed -n '2,9p' "$0"; exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

errors=0
ok() { printf 'ok   %s\n' "$*"; }
fail() { printf 'FAIL %s\n' "$*"; errors=$((errors + 1)); }
skip() { printf 'skip %s\n' "$*"; }

command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }

plugin=.claude-plugin/plugin.json
for f in "$plugin" glama.json; do
  if jq empty "$f" 2>/dev/null; then ok "$f parses"; else fail "$f is not valid JSON"; fi
done

check() {
  if jq -e "$2" "$plugin" >/dev/null 2>&1; then ok "$1"; else fail "$1"; fi
}
check "plugin name is trident" '.name == "trident"'
check "plugin version is semver" '.version | test("^[0-9]+\\.[0-9]+\\.[0-9]+$")'
check "plugin runs the trident-mcp binary" '.mcpServers.trident.command == "trident-mcp"'
check "plugin passes the API key via TRIDENT_TRIPO_API_KEY" '.mcpServers.trident.env.TRIDENT_TRIPO_API_KEY == "${user_config.tripo_api_key}"'
check "plugin API key setting is sensitive" '.userConfig.tripo_api_key.sensitive == true'

version=$(jq -r .version "$plugin")
if [ -n "$expect" ]; then
  if [ "$version" = "$expect" ]; then ok "plugin version $version matches $expect"; else fail "plugin version $version, expected $expect (run scripts/sync-version.sh $expect)"; fi
fi
for skill in skills/*/SKILL.md; do
  skill_version=$(awk -F'"' '/^  version:/ { print $2; exit }' "$skill")
  if [ "$skill_version" = "$version" ]; then ok "$skill version $skill_version"; else fail "$skill version '$skill_version' differs from plugin version $version"; fi
done

if command -v claude >/dev/null 2>&1; then
  if claude plugin validate . >/dev/null 2>&1; then ok "claude plugin validate"; else fail "claude plugin validate"; claude plugin validate . 2>&1 | sed 's/^/     /'; fi
elif ! $offline && command -v npx >/dev/null 2>&1; then
  if npx -y @anthropic-ai/claude-code plugin validate . >/dev/null 2>&1; then ok "claude plugin validate (npx)"; else fail "claude plugin validate (npx)"; npx -y @anthropic-ai/claude-code plugin validate . 2>&1 | sed 's/^/     /'; fi
else
  skip "claude plugin validate (no claude CLI)"
fi

if [ "$errors" -gt 0 ]; then
  echo "$errors check(s) failed"
  exit 1
fi
echo "packaging ok"
