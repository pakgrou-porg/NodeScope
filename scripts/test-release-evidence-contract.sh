#!/usr/bin/env bash
# Validate deterministic report semantics and exercise the release-manifest
# assembler with disposable artifacts only.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

for required in \
  'refusing to report a dirty working tree' \
  'commit_timestamp_unix' \
  'No live Framework hardware qualification' \
  'No tagged release or GitHub release-attestation verification'; do
  grep -Fq "$required" scripts/write-release-readiness-report.sh || {
    echo "readiness report must retain $required" >&2
    exit 1
  }
done

for required in \
  'anchore/sbom-action@v0' \
  'provenance: mode=max' \
  'sha256sum "$archive"' \
  'Assemble machine-readable release evidence' \
  'Attest machine-readable release evidence' \
  'softprops/action-gh-release@v2'; do
  grep -Fq "$required" .github/workflows/release.yml || {
    echo "release workflow must retain $required" >&2
    exit 1
  }
done
assemble_line="$(grep -n 'Assemble machine-readable release evidence' .github/workflows/release.yml | cut -d: -f1)"
publish_line="$(grep -n 'Create immutable GitHub release' .github/workflows/release.yml | cut -d: -f1)"
if [[ -z "$assemble_line" || -z "$publish_line" || "$assemble_line" -ge "$publish_line" ]]; then
  echo "release evidence must be assembled before immutable release publication" >&2
  exit 1
fi

fixture="$(mktemp -d)"
trap 'rm -rf "$fixture"' EXIT
mkdir -p "$fixture/release"
printf 'release-candidate\n' >"$fixture/release/nodescope_1.0.0_linux_amd64.tar.gz"
(cd "$fixture/release" && sha256sum nodescope_1.0.0_linux_amd64.tar.gz > nodescope_1.0.0_linux_amd64.tar.gz.sha256)
printf '{"spdxVersion":"SPDX-2.3"}\n' >"$fixture/release/nodescope_v1.0.0_linux_amd64.spdx.json"
GITHUB_REF_NAME=v1.0.0 GITHUB_SHA=fixture-commit ./scripts/assemble-release-evidence.sh --release-directory "$fixture/release" --output "$fixture/release/release-evidence.json"
for required in \
  '"release_tag":"v1.0.0"' \
  '"commit_sha":"fixture-commit"' \
  '"sha256"' \
  '"sboms"' \
  'GitHub Actions artifact attestations'; do
  grep -Fq "$required" "$fixture/release/release-evidence.json" || {
    echo "release evidence manifest must contain $required" >&2
    exit 1
  }
done

echo "Release evidence contract passed."
