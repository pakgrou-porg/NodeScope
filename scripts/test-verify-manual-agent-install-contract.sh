#!/usr/bin/env bash
# Verify that manual source verification refuses unreviewed tracked or untracked
# content before it can run tests or release-evidence checks.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
verifier="$repository_root/scripts/verify-manual-agent-install.sh"

grep -Fq 'git status --porcelain=v1 --untracked-files=all' "$verifier" || {
  echo 'manual verifier must inspect tracked and untracked source state' >&2
  exit 1
}
grep -Fq 'manual verification requires a clean tracked and untracked source tree' "$verifier" || {
  echo 'manual verifier must explain the clean-tree boundary' >&2
  exit 1
}

fixture="$repository_root/.nodescope-manual-verification-contract.tmp"
trap 'rm -f "$fixture"' EXIT
printf '%s\n' 'fixture only' >"$fixture"
if "$verifier" --offline >/tmp/nodescope-manual-verification-dirty.out 2>&1; then
  echo 'manual verifier accepted an untracked source file' >&2
  exit 1
fi
grep -Fq 'clean tracked and untracked source tree' /tmp/nodescope-manual-verification-dirty.out

echo 'Manual agent installation verification contract passed.'
