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
gh_bin="${NODESCOPE_GH_BIN:-gh}"

fail() {
  printf 'release evidence verification failed: %s\n' "$1" >&2
  exit 1
}

[[ -f "$archive" && ! -L "$archive" ]] || fail "archive must be a regular non-symlink file"
[[ -f "$archive_checksum_file" && ! -L "$archive_checksum_file" ]] || fail "archive checksum file must be a regular non-symlink file"
[[ -f "$sbom" && ! -L "$sbom" ]] || fail "SBOM must be a regular non-symlink file"
[[ -f "$sbom_checksum_file" && ! -L "$sbom_checksum_file" ]] || fail "SBOM checksum file must be a regular non-symlink file"
[[ "$release_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z]+)*$ ]] || fail "release tag must be pinned vMAJOR.MINOR.PATCH"
[[ "$source_revision" =~ ^[a-fA-F0-9]{40,64}$ ]] || fail "source revision must be immutable hexadecimal"
[[ "$repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail "repository must be owner/name"
command -v "$gh_bin" >/dev/null 2>&1 || fail "GitHub CLI is required for attestation verification"

expected_checksum="$(awk 'NF >= 1 {print $1; exit}' "$archive_checksum_file")"
[[ "$expected_checksum" =~ ^[a-fA-F0-9]{64}$ ]] || fail "archive checksum file does not begin with a SHA-256 digest"
actual_checksum="$(sha256sum "$archive" | awk '{print $1}')"
[[ "${actual_checksum,,}" == "${expected_checksum,,}" ]] || fail "archive SHA-256 does not match pinned checksum"

expected_sbom_checksum="$(awk 'NF >= 1 {print $1; exit}' "$sbom_checksum_file")"
[[ "$expected_sbom_checksum" =~ ^[a-fA-F0-9]{64}$ ]] || fail "SBOM checksum file does not begin with a SHA-256 digest"
actual_sbom_checksum="$(sha256sum "$sbom" | awk '{print $1}')"
[[ "${actual_sbom_checksum,,}" == "${expected_sbom_checksum,,}" ]] || fail "SBOM SHA-256 does not match pinned checksum"
node -e 'const fs = require("fs"); const sbom = JSON.parse(fs.readFileSync(process.argv[1], "utf8")); if (typeof sbom.spdxVersion !== "string" || !Array.isArray(sbom.packages)) process.exit(1)' "$sbom" || fail "SBOM must be a valid SPDX JSON document with packages"

"$gh_bin" attestation verify "$archive" -R "$repository" >/dev/null
release_target="$("$gh_bin" api "repos/$repository/releases/tags/$release_tag" --jq '.target_commitish')"
[[ -n "$release_target" ]] || fail "release target is empty"
resolved_revision="$("$gh_bin" api "repos/$repository/commits/$release_target" --jq '.sha')"
[[ "${resolved_revision,,}" == "${source_revision,,}" ]] || fail "release target does not resolve to supplied source revision"

printf '{"repository":"%s","release_tag":"%s","source_revision":"%s","archive_sha256":"%s","sbom_sha256":"%s","attestation":"verified"}\n' \
  "$repository" "$release_tag" "${source_revision,,}" "${actual_checksum,,}" "${actual_sbom_checksum,,}"
