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

output_directory="$(dirname "$output")"
mkdir -p "$output_directory"
output_abs="$(cd "$output_directory" && pwd)/$(basename "$output")"
case "$output_abs" in
  "$repository_root"/*)
    echo "--output must refer to a path outside the repository" >&2
    exit 1
    ;;
esac

./scripts/release-readiness-check.sh

if [[ -n "$(git status --porcelain)" ]]; then
  echo "readiness generated an uncommitted change" >&2
  exit 1
fi

commit_sha="$(git rev-parse HEAD)"
commit_timestamp="$(git show -s --format=%ct HEAD)"
cat >"$output_abs" <<EOF
{"schema_version":2,"scope":"local deterministic release readiness","result":"passed","commit_sha":"${commit_sha}","commit_timestamp_unix":${commit_timestamp},"working_tree_clean":true,"checks":[{"id":"source_and_policy","command":"./scripts/release-readiness-check.sh (repository, license, workflow, installation, and release contracts)","expected":"source and policy contracts pass without credentialed deployment access","observed":"passed","evidence_boundary":"local source and CI-contract validation only","known_limitation":"No tagged release or GitHub release-attestation verification","recovery":"Do not promote a release; restore the prior accepted signed tag and remediate the failed contract."},{"id":"shared_schema_safety","command":"./scripts/release-readiness-check.sh (migration and shared-Supabase fixture contracts)","expected":"schema isolation and migration-safety contracts pass without a production database mutation","observed":"passed","evidence_boundary":"disposable-fixture and source-contract validation only","known_limitation":"No real sibling schema or production telemetry path validation","recovery":"Stop before any database apply, rerun isolation gates, and restore through the approved migration recovery procedure."},{"id":"local_resilience","command":"./scripts/release-readiness-check.sh (local resilience and compose-preflight contracts)","expected":"lease, PKI, TLS, archive, and failover/failback rehearsal contracts pass locally","observed":"passed","evidence_boundary":"local rehearsal and preflight only","known_limitation":"No live dual-replica failover, certificate revocation, or isolated restore acceptance","recovery":"Keep deployment paused; fence the prior writer and use the documented replica, PKI, and restore recovery procedures after an approved rehearsal."},{"id":"native_builds","command":"./scripts/release-readiness-check.sh (Go tests and Linux/Windows cross-builds)","expected":"native package tests and cross-builds pass for declared targets","observed":"passed","evidence_boundary":"build and deterministic test validation only","known_limitation":"No live Framework, Asus, or MSI hardware qualification","recovery":"Keep unsupported platforms unenrolled; use installer rollback metadata and revoke a test credential if host qualification fails."},{"id":"browser_console","command":"./scripts/release-readiness-check.sh (TypeScript, Vitest, contracts, and production bundle)","expected":"browser-console source and build checks pass without a real identity provider","observed":"passed","evidence_boundary":"fixture-backed console and source-contract validation only","known_limitation":"No real Supabase magic-link, approved-backend streaming, or tagged release-attestation verification","recovery":"Do not claim operational acceptance; restore the prior accepted UI revision and rerun browser E2E only in an approved environment."}],"evidence_locations":["scripts/release-readiness-check.sh","docs/operations/release-epics.md","docs/operations/local-release-readiness-report.md"],"live_gates_retained":["No live Framework hardware qualification","No live dual-replica failover, certificate revocation, or isolated restore acceptance","No real Supabase magic-link, approved-backend streaming, or tagged release-attestation verification"],"recovery":"Do not promote a release from this report alone. Preserve the failed command output, restore the prior accepted signed tag as appropriate, and rerun this report after remediation."}
EOF

./scripts/verify-release-readiness-report.sh "$output_abs"
printf 'Wrote deterministic local readiness report: %s\n' "$output_abs"
