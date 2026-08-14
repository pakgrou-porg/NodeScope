#!/usr/bin/env bash
# Assemble a release manifest from artifacts that have already passed checksum
# verification. GitHub Actions attests this manifest in the tagged release job.
set -euo pipefail

release_directory=""
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --release-directory)
      release_directory="${2:-}"
      shift 2
      ;;
    --output)
      output="${2:-}"
      shift 2
      ;;
    *)
      echo "usage: $0 --release-directory <directory> --output <path>" >&2
      exit 2
      ;;
  esac
done
if [[ -z "$release_directory" || -z "$output" ]]; then
  echo "--release-directory and --output are required" >&2
  exit 2
fi
if [[ ! -d "$release_directory" ]]; then
  echo "release directory does not exist" >&2
  exit 1
fi

mapfile -t archives < <(find "$release_directory" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) -printf '%f\n' | LC_ALL=C sort)
mapfile -t sboms < <(find "$release_directory" -maxdepth 1 -type f -name '*.spdx.json' -printf '%f\n' | LC_ALL=C sort)
if [[ "${#archives[@]}" -eq 0 || "${#sboms[@]}" -eq 0 ]]; then
  echo "release evidence requires at least one archive and SPDX SBOM" >&2
  exit 1
fi

artifacts_json=""
for archive in "${archives[@]}"; do
  checksum_file="$release_directory/$archive.sha256"
  if [[ ! -f "$checksum_file" ]]; then
    echo "missing checksum for $archive" >&2
    exit 1
  fi
  (cd "$release_directory" && sha256sum --check "$(basename "$checksum_file")") >/dev/null
  digest="$(awk '{print $1}' "$checksum_file")"
  artifacts_json+="${artifacts_json:+,}{\"name\":\"${archive}\",\"sha256\":\"${digest}\",\"checksum\":\"${archive}.sha256\"}"
done

sboms_json=""
for sbom in "${sboms[@]}"; do
  sboms_json+="${sboms_json:+,}\"${sbom}\""
done

mkdir -p "$(dirname "$output")"
cat >"$output" <<EOF
{"schema_version":1,"release_tag":"${GITHUB_REF_NAME:-local-unreleased}","commit_sha":"${GITHUB_SHA:-local-uncommitted}","artifacts":[${artifacts_json}],"sboms":[${sboms_json}],"provenance":"GitHub Actions artifact attestations are attached by the release workflow","signing":"The release tag is verified before assembly; this manifest is itself attested by GitHub Actions.","verification":"Verify each .sha256 file before consuming an artifact, then verify the associated GitHub attestation."}
EOF

printf 'Wrote release evidence manifest: %s\n' "$output"
