#!/usr/bin/env bash
# Produce a deterministic, machine-readable local readiness report. The report
# is intentionally scoped to local validation and does not claim environment or
# operational acceptance.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      output="${2:-}"
      shift 2
      ;;
    *)
      echo "usage: $0 --output <path>" >&2
      exit 2
      ;;
  esac
done
if [[ -z "$output" ]]; then
  echo "--output is required" >&2
  exit 2
fi
if [[ -n "$(git status --porcelain)" ]]; then
  echo "refusing to report a dirty working tree" >&2
  exit 1
fi

./scripts/release-readiness-check.sh

if [[ -n "$(git status --porcelain)" ]]; then
  echo "readiness generated an uncommitted change" >&2
  exit 1
fi

commit_sha="$(git rev-parse HEAD)"
commit_timestamp="$(git show -s --format=%ct HEAD)"
mkdir -p "$(dirname "$output")"
cat >"$output" <<EOF
{"schema_version":1,"scope":"local deterministic release readiness","result":"passed","commit_sha":"${commit_sha}","commit_timestamp_unix":${commit_timestamp},"commands":[{"id":"aggregate_readiness","command":"./scripts/release-readiness-check.sh","expected":"all deterministic local checks pass and leave no generated-contract drift","observed":"passed"}],"evidence_locations":["scripts/release-readiness-check.sh","docs/operations/release-epics.md"],"known_limitations":["No live Framework hardware qualification","No live dual-replica failover, certificate revocation, or isolated restore acceptance","No tagged release or GitHub release-attestation verification"],"recovery":"Do not promote a release; restore the prior accepted signed tag and rerun this report after remediation."}
EOF

printf 'Wrote deterministic local readiness report: %s\n' "$output"
