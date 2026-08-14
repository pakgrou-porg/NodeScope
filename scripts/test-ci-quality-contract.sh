#!/usr/bin/env bash
# Keep static Go analysis as a deterministic release and CI control.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

grep -Fq 'go vet ./...' scripts/release-readiness-check.sh || {
  echo "release readiness must run go vet" >&2
  exit 1
}
grep -Fq 'go vet ./...' .github/workflows/ci.yml || {
  echo "CI must run go vet" >&2
  exit 1
}

echo "CI quality contract passed."
