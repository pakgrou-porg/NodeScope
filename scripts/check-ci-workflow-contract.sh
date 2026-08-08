#!/usr/bin/env bash
# Validate CI policy independently of the GitHub Actions runtime.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
workflow="${1:-$repository_root/.github/workflows/ci.yml}"

require() {
  local expected="$1"
  if ! grep -Fq -- "$expected" "$workflow"; then
    printf 'CI workflow is missing required policy: %s\n' "$expected" >&2
    exit 1
  fi
}

require 'windows-agent-runtime:'
require 'runs-on: windows-2022'
require 'go test ./internal/agent'
require 'windows-agent:'
require 'GOOS=windows GOARCH=${{ matrix.goarch }} go build ./cmd/nodescope-agent'
require 'GOOS=windows GOARCH=${{ matrix.goarch }} go test -c'
require 'pnpm/action-setup@v4'

if awk '
  /^[[:space:]]*-[[:space:]]+uses:[[:space:]]+pnpm\/action-setup@v4[[:space:]]*$/ {
    in_pnpm_setup = 1
    in_with_block = 0
    next
  }
  in_pnpm_setup && /^[[:space:]]*-[[:space:]]+uses:/ {
    in_pnpm_setup = 0
    in_with_block = 0
  }
  in_pnpm_setup && /^[[:space:]]+with:[[:space:]]*$/ {
    in_with_block = 1
    next
  }
  in_pnpm_setup && in_with_block && /^[[:space:]]+version:[[:space:]]/ {
    found_override = 1
    exit
  }
  END { exit(found_override ? 0 : 1) }
' "$workflow"; then
  printf 'CI workflow must defer pnpm selection to package.json packageManager rather than specify a pnpm/action-setup version\n' >&2
  exit 1
fi

printf 'NodeScope CI workflow contract is valid.\n'
