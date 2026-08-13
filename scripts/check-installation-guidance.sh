#!/usr/bin/env bash
# Verify that operator-facing installation guidance remains aligned with the
# supported native-agent security and experimental-telemetry boundaries.
set -euo pipefail

repository_root="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
docs_root="$repository_root/docs"
framework_guide="$docs_root/agents/framework-asus-install.md"
manual_guide="$docs_root/agents/manual-install-framework-asus-v2.md"
windows_guide="$docs_root/agents/windows-msi-install.md"

fail() {
  printf 'installation guidance policy failed: %s\n' "$1" >&2
  exit 1
}

for path in "$framework_guide" "$manual_guide" "$windows_guide"; do
  [[ -f "$path" ]] || fail "missing required guide $path"
done

if grep -RInE --include='*.md' '(dnf install .*amdrocm-amdsmi|dnf install .*xrt|apt install .*xrt)' "$docs_root"; then
  fail "unqualified Fedora accelerator package-install guidance is forbidden"
fi
if grep -RInE --include='*.md' '(adduser .*docker|usermod .*docker|Grant the service account read-only Docker)' "$docs_root"; then
  fail "direct Docker-group or socket-access guidance is forbidden"
fi

grep -qi 'experimental' "$framework_guide" || fail "Framework guide must classify unqualified Fedora accelerators as experimental"
grep -q 'NODESCOPE_AGENT_CREDENTIAL_FILE' "$framework_guide" || fail "Framework guide must use a protected credential-file reference"
grep -q 'NODESCOPE_REQUIRE_CLIENT_MTLS' "$framework_guide" || fail "Framework guide must document explicit client mTLS policy"
grep -q -- '--ingestion-preflight' "$framework_guide" || fail "Framework guide must document authenticated non-mutating preflight"
grep -q 'fixed-schema HTTPS proxy' "$framework_guide" || fail "Framework guide must preserve proxy-only container inventory guidance"
grep -q 'NODESCOPE_REQUIRE_CLIENT_MTLS' "$manual_guide" || fail "manual guide must document explicit client mTLS policy"
grep -q -- '--ingestion-preflight' "$manual_guide" || fail "manual guide must document authenticated non-mutating preflight"
grep -q 'NODESCOPE_REQUIRE_CLIENT_MTLS' "$windows_guide" || fail "Windows guide must document explicit client mTLS policy"
grep -q -- '--ingestion-preflight' "$windows_guide" || fail "Windows guide must document authenticated non-mutating preflight"

printf '%s\n' 'Installation guidance policy is valid.'
