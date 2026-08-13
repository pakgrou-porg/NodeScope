#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
installer="${1:-$repository_root/deploy/agent/install-linux.sh}"

fail() {
  printf 'Linux installer contract failed: %s\n' "$1" >&2
  exit 1
}

bash -n "$installer"
grep -q 'agent-binary> <agent-sha256> <unit-file> <unit-sha256> <release-tag> <source-revision' "$installer" || fail "installer must require pinned release and source arguments"
grep -q 'metadata_path=' "$installer" || fail "installer must persist provenance metadata"
grep -q 'previous_binary_backup=' "$installer" || fail "installer metadata must include prior binary rollback reference"
grep -q 'previous_unit_backup=' "$installer" || fail "installer metadata must include prior unit rollback reference"
grep -q 'previous_metadata_backup=' "$installer" || fail "installer metadata must include prior metadata rollback reference"
grep -q 'NODESCOPE_AGENT_CREDENTIAL_FILE=' "$installer" || fail "installer template must use a credential-file reference"
grep -q 'NODESCOPE_REQUIRE_CLIENT_MTLS=false' "$installer" || fail "installer template must expose explicit client mTLS policy"

printf '%s\n' 'Linux installer contract is valid.'
