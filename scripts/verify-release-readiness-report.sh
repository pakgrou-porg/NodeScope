#!/usr/bin/env bash
# Validate the stable field set of a locally generated readiness report without
# contacting any deployment, database, host, or external runtime.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <report.json>" >&2
  exit 2
fi

report="$1"
if [[ ! -f "$report" ]]; then
  echo "readiness report does not exist: $report" >&2
  exit 1
fi

for required in \
  '"schema_version":2' \
  '"scope":"local deterministic release readiness"' \
  '"result":"passed"' \
  '"commit_sha":' \
  '"commit_timestamp_unix":' \
  '"working_tree_clean":true' \
  '"checks":[' \
  '"id":"source_and_policy"' \
  '"id":"shared_schema_safety"' \
  '"id":"local_resilience"' \
  '"id":"native_builds"' \
  '"id":"browser_console"' \
  '"command":' \
  '"expected":' \
  '"observed":"passed"' \
  '"evidence_boundary":' \
  '"known_limitation":' \
  '"recovery":' \
  '"live_gates_retained":[' \
  '"No live Framework hardware qualification"' \
  '"No live dual-replica failover, certificate revocation, or isolated restore acceptance"' \
  '"No real Supabase magic-link, approved-backend streaming, or tagged release-attestation verification"'; do
  grep -Fq -- "$required" "$report" || {
    echo "readiness report missing required field or boundary: $required" >&2
    exit 1
  }
done

echo "Release-readiness report verified: $report"
