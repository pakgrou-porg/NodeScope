#!/usr/bin/env bash
# Exercise both acceptance and rejection paths of the static release policy.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
checker="$repository_root/scripts/check-release-workflow-contract.sh"
workflow="$repository_root/.github/workflows/release.yml"

"$checker" "$workflow"

fixture="$(mktemp)"
trap 'rm -f "$fixture"' EXIT
sed 's/git verify-tag "\$GITHUB_REF_NAME"/true/' "$workflow" > "$fixture"
if "$checker" "$fixture"; then
  printf 'release workflow contract accepted an unsigned-tag fixture\n' >&2
  exit 1
fi

printf 'NodeScope signed-tag release-workflow contract regression test passed.\n'
