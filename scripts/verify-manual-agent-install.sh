#!/usr/bin/env bash
# Run non-mutating source and release-evidence verification before a manual
# native-agent install. It never invokes an installer or writes host state.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

command -v git >/dev/null 2>&1 || { printf '%s\n' 'manual verification requires git on PATH' >&2; exit 1; }
command -v pnpm >/dev/null 2>&1 || { printf '%s\n' 'manual verification requires pnpm on PATH' >&2; exit 1; }
command -v go >/dev/null 2>&1 || { printf '%s\n' 'manual verification requires the pinned Go toolchain on PATH' >&2; exit 1; }

if [[ $# -eq 1 && "$1" == "--offline" ]]; then
  git diff --check
  pnpm check
  go test ./...
  printf '%s\n' 'Manual agent source verification passed in offline mode; release evidence was not checked.'
  exit 0
fi
if [[ $# -lt 4 || $# -gt 5 ]]; then
  cat >&2 <<'USAGE'
Usage:
  verify-manual-agent-install.sh --offline
  verify-manual-agent-install.sh <archive> <archive-checksum-file> <sbom> <sbom-checksum-file> <release-tag> <source-revision> [owner/repository]
USAGE
  exit 2
fi

git diff --check
pnpm check
go test ./...
"$repository_root/scripts/verify-agent-release-evidence.sh" "$@"
printf '%s\n' 'Manual agent source and signed-release evidence verification passed.'
