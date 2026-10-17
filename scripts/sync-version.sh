#!/usr/bin/env bash
# Sets the release version in the plugin manifest and every skill's metadata.
# Usage: scripts/sync-version.sh X.Y.Z   (then commit, then tag vX.Y.Z)
set -euo pipefail

version=${1:?usage: scripts/sync-version.sh X.Y.Z}
version=${version#v}
[[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "not a semver version: $version" >&2; exit 2; }

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

tmp=$(mktemp)
jq --arg v "$version" '.version = $v' .claude-plugin/plugin.json >"$tmp" && mv "$tmp" .claude-plugin/plugin.json
for skill in skills/*/SKILL.md; do
  sed -i.bak -E "s/^(  version: )\"[^\"]*\"/\1\"$version\"/" "$skill" && rm -f "$skill.bak"
done
echo "version set to $version"
