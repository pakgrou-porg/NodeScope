#!/usr/bin/env bash
# Verify exact release evidence before an administrator provides artifacts to
# the root-owned native-agent installer. This command never writes host state.
set -euo pipefail

if [[ $# -lt 6 || $# -gt 7 ]]; then
  cat >&2 <<'USAGE'
Usage:
  verify-agent-release-evidence.sh <archive> <archive-checksum-file> <sbom> <sbom-checksum-file> <release-tag> <source-revision> [owner/repository]

The command verifies pinned archive and SPDX SBOM checksums, GitHub artifact
attestation, and that the release target resolves to the immutable source revision.
USAGE
  exit 2
fi

archive="$1"
archive_checksum_file="$2"
sbom="$3"
sbom_checksum_file="$4"
release_tag="$5"
source_revision="$6"
repository="${7:-pakgrou-porg/NodeScope}"

fail() {
  printf 'release evidence verification failed: %s\n' "$1" >&2
  exit 1
}

verify_exact_checksum_sidecar() {
  local target="$1"
  local sidecar="$2"
  local label="$3"
  local expected_name
  local line_count
  local expected_checksum
  local actual_checksum

  expected_name="$(basename "$target")"
  line_count="$(awk -v expected_name="$expected_name" '
    NF == 0 { next }
    NF != 2 || length($1) != 64 || $1 !~ /^[a-fA-F0-9]+$/ || $2 != expected_name { invalid = 1; next }
    { count++; digest = $1 }
    END { if (!invalid && count == 1) print digest; else exit 1 }
  ' "$sidecar")" || fail "$label checksum sidecar must contain exactly one SHA-256 entry for $expected_name"
  expected_checksum="$line_count"
  actual_checksum="$(sha256sum "$target" | awk '{print $1}')"
  [[ "${actual_checksum,,}" == "${expected_checksum,,}" ]] || fail "$label SHA-256 does not match exact checksum sidecar"
}

[[ -f "$archive" && ! -L "$archive" ]] || fail "archive must be a regular non-symlink file"
[[ -f "$archive_checksum_file" && ! -L "$archive_checksum_file" ]] || fail "archive checksum file must be a regular non-symlink file"
[[ -f "$sbom" && ! -L "$sbom" ]] || fail "SBOM must be a regular non-symlink file"
[[ -f "$sbom_checksum_file" && ! -L "$sbom_checksum_file" ]] || fail "SBOM checksum file must be a regular non-symlink file"
[[ "$release_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z]+)*$ ]] || fail "release tag must be pinned vMAJOR.MINOR.PATCH"
[[ "$source_revision" =~ ^[a-fA-F0-9]{40}$ ]] || fail "source revision must be a canonical 40-character hexadecimal GitHub commit"
[[ "$repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail "repository must be owner/name"
command -v gh >/dev/null 2>&1 || fail "GitHub CLI is required for attestation verification"

verify_exact_checksum_sidecar "$archive" "$archive_checksum_file" "archive"
verify_exact_checksum_sidecar "$sbom" "$sbom_checksum_file" "SBOM"
actual_checksum="$(sha256sum "$archive" | awk '{print $1}')"
actual_sbom_checksum="$(sha256sum "$sbom" | awk '{print $1}')"
"$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/validate-spdx-sbom.mjs" "$sbom" || fail "SBOM must be a structurally valid SPDX JSON document"

gh attestation verify "$archive" -R "$repository" >/dev/null
release_target="$(gh api "repos/$repository/releases/tags/$release_tag" --jq '.target_commitish')"
[[ -n "$release_target" ]] || fail "release target is empty"
resolved_revision="$(gh api "repos/$repository/commits/$release_target" --jq '.sha')"
[[ "${resolved_revision,,}" == "${source_revision,,}" ]] || fail "release target does not resolve to supplied source revision"

printf '{"repository":"%s","release_tag":"%s","source_revision":"%s","archive_sha256":"%s","sbom_sha256":"%s","attestation":"verified"}\n' \
  "$repository" "$release_tag" "${source_revision,,}" "${actual_checksum,,}" "${actual_sbom_checksum,,}"
