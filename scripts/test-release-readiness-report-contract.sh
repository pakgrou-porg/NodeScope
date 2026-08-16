#!/usr/bin/env bash
# Keep the aggregate report deterministic, locally scoped, and explicit about
# the environment checks it cannot truthfully claim.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

for required in \
  'refusing to report a dirty working tree' \
  '--output must refer to a path outside the repository' \
  '"schema_version":2' \
  '"working_tree_clean":true' \
  '"source_and_policy"' \
  '"shared_schema_safety"' \
  '"local_resilience"' \
  '"native_builds"' \
  '"browser_console"' \
  'No real Supabase magic-link, approved-backend streaming, or tagged release-attestation verification' \
  './scripts/verify-release-readiness-report.sh "$output_abs"'; do
  grep -Fq -- "$required" scripts/write-release-readiness-report.sh || {
    echo "readiness-report generator must retain: $required" >&2
    exit 1
  }
done

for required in \
  '"schema_version":2' \
  '"id":"browser_console"' \
  '"live_gates_retained":[' \
  'Release-readiness report verified:'; do
  grep -Fq -- "$required" scripts/verify-release-readiness-report.sh || {
    echo "readiness-report verifier must retain: $required" >&2
    exit 1
  }
done

echo "Release-readiness report contract passed."
