#!/usr/bin/env bash
# Validate the release workflow's local policy contract without requiring a
# GitHub environment, signing key, or tag push.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
workflow="${1:-$repository_root/.github/workflows/release.yml}"

require() {
  local expected="$1"
  if ! grep -Fq -- "$expected" "$workflow"; then
    printf 'release workflow is missing required policy: %s\n' "$expected" >&2
    exit 1
  fi
}

require 'verify-release-tag:'
require 'environment: approved-release'
require 'NODESCOPE_RELEASE_SIGNING_PUBLIC_KEY: ${{ secrets.NODESCOPE_RELEASE_SIGNING_PUBLIC_KEY }}'
require 'git verify-tag "$GITHUB_REF_NAME"'
require 'needs: [verify-release-tag, binaries, windows-agent, container]'
require 'name: native-windows-${{ matrix.goarch }}'

release_dependencies="$(grep -Fxc '    needs: verify-release-tag' "$workflow" || true)"
if [[ "$release_dependencies" != "3" ]]; then
  printf 'release workflow must gate Linux, Windows, and container jobs on signed-tag verification; found %s dependencies\n' "$release_dependencies" >&2
  exit 1
fi

printf 'NodeScope signed-tag release-workflow contract is valid.\n'
