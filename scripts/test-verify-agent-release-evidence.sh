#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="$(mktemp -d)"
trap 'rm -rf "$fixture"' EXIT

archive="$fixture/nodescope.tar.gz"
checksum="$fixture/nodescope.tar.gz.sha256"
sbom="$fixture/nodescope.spdx.json"
sbom_checksum="$fixture/nodescope.spdx.json.sha256"
printf '%s' 'verified fixture archive' >"$archive"
sha256sum "$archive" >"$checksum"
printf '%s\n' '{"spdxVersion":"SPDX-2.3","packages":[]}' >"$sbom"
sha256sum "$sbom" >"$sbom_checksum"

cat >"$fixture/gh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == "attestation" && "$2" == "verify" ]]; then exit 0; fi
if [[ "$1" == "api" && "$2" == *"releases/tags/v1.2.3"* ]]; then printf '%s\n' '0123456789abcdef0123456789abcdef01234567'; exit 0; fi
if [[ "$1" == "api" && "$2" == *"commits/0123456789abcdef0123456789abcdef01234567"* ]]; then printf '%s\n' '0123456789abcdef0123456789abcdef01234567'; exit 0; fi
exit 1
EOF
chmod +x "$fixture/gh"

NODESCOPE_GH_BIN="$fixture/gh" "$repository_root/scripts/verify-agent-release-evidence.sh" "$archive" "$checksum" "$sbom" "$sbom_checksum" v1.2.3 0123456789abcdef0123456789abcdef01234567 >"$fixture/result.json"
grep -q '"attestation":"verified"' "$fixture/result.json"
grep -q '"sbom_sha256"' "$fixture/result.json"

if NODESCOPE_GH_BIN="$fixture/gh" "$repository_root/scripts/verify-agent-release-evidence.sh" "$archive" "$checksum" "$sbom" "$sbom_checksum" v1.2.3 ffffffffffffffffffffffffffffffffffffffff; then
  printf '%s\n' 'expected source revision mismatch to fail' >&2
  exit 1
fi

printf '%s\n' '{"not":"spdx"}' >"$sbom"
sha256sum "$sbom" >"$sbom_checksum"
if NODESCOPE_GH_BIN="$fixture/gh" "$repository_root/scripts/verify-agent-release-evidence.sh" "$archive" "$checksum" "$sbom" "$sbom_checksum" v1.2.3 0123456789abcdef0123456789abcdef01234567; then
  printf '%s\n' 'expected malformed SPDX SBOM to fail' >&2
  exit 1
fi

printf '%s\n' 'Release evidence verifier regression checks passed.'
