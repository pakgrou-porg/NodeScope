#!/usr/bin/env bash
# Keep workflow JavaScript actions on supported Node 24-capable major versions.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

for workflow in .github/workflows/ci.yml .github/workflows/release.yml; do
  grep -Fq 'actions/checkout@v7' "$workflow" || {
    echo "$workflow must use actions/checkout@v7" >&2
    exit 1
  }
  grep -Fq 'actions/setup-go@v7' "$workflow" || {
    echo "$workflow must use actions/setup-go@v7" >&2
    exit 1
  }
done

grep -Fq 'pnpm/action-setup@v6' .github/workflows/ci.yml || {
  echo "CI must use pnpm/action-setup@v6" >&2
  exit 1
}
grep -Fq 'actions/setup-node@v7' .github/workflows/ci.yml || {
  echo "CI must use actions/setup-node@v7" >&2
  exit 1
}
grep -Fq 'gitleaks/gitleaks-action@v3' .github/workflows/ci.yml || {
  echo "CI must use gitleaks/gitleaks-action@v3" >&2
  exit 1
}

if grep -Eq '(actions/checkout@v[1-6]|actions/setup-node@v[1-6]|actions/setup-go@v[1-6]|pnpm/action-setup@v[1-5]|gitleaks/gitleaks-action@v[1-2])' .github/workflows/ci.yml .github/workflows/release.yml; then
  echo "workflow contains a deprecated Node 20-era action major" >&2
  exit 1
fi

echo "Workflow runtime contract passed."
