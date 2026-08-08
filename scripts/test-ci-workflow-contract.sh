#!/usr/bin/env bash
# Exercise acceptance and failure paths of the static CI policy checker.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
checker="$repository_root/scripts/check-ci-workflow-contract.sh"
workflow="$repository_root/.github/workflows/ci.yml"

"$checker" "$workflow"

fixture="$(mktemp)"
unrelated_fixture="$(mktemp)"
trap 'rm -f "$fixture" "$unrelated_fixture"' EXIT
awk '
  /- uses: pnpm\/action-setup@v4/ && !added {
    print
    print "        with:"
    print "          version: 10"
    added = 1
    next
  }
  { print }
' "$workflow" > "$fixture"

if "$checker" "$fixture"; then
  printf 'CI workflow contract accepted a conflicting pnpm version fixture\n' >&2
  exit 1
fi

cat "$workflow" > "$unrelated_fixture"
printf '\nmetadata:\n  version: "fixture-only"\n' >> "$unrelated_fixture"
if ! "$checker" "$unrelated_fixture"; then
  printf 'CI workflow contract rejected an unrelated version field fixture\n' >&2
  exit 1
fi

printf 'NodeScope CI workflow contract regression test passed.\n'
