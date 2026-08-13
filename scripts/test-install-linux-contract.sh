#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
installer="$repository_root/deploy/agent/install-linux.sh"
fixture="$(mktemp -d)"
trap 'rm -rf "$fixture"' EXIT

cp "$installer" "$fixture/install-linux.sh"
chmod +x "$fixture/install-linux.sh"

"$repository_root/scripts/check-install-linux-contract.sh"

sed -i 's/<release-tag> <source-revision>/<release-tag>/' "$fixture/install-linux.sh"
if "$repository_root/scripts/check-install-linux-contract.sh" "$fixture/install-linux.sh"; then
  printf '%s\n' 'expected installer contract mutation to be rejected' >&2
  exit 1
fi

printf '%s\n' 'Linux installer contract regression checks passed.'
