#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
checker="$repository_root/scripts/check-installation-guidance.sh"
fixture="$(mktemp -d)"
trap 'rm -rf "$fixture"' EXIT

mkdir -p "$fixture/docs"
cp -R "$repository_root/docs/agents" "$fixture/docs/agents"
cp -R "$repository_root/docs/architecture" "$fixture/docs/architecture"

"$checker" "$fixture"

printf '\nsudo dnf install amdrocm-amdsmi\n' >> "$fixture/docs/agents/preflight-dependencies.md"
if "$checker" "$fixture"; then
  printf '%s\n' 'expected unsupported Fedora package guidance to be rejected' >&2
  exit 1
fi

printf '%s\n' 'Installation guidance regression checks passed.'
