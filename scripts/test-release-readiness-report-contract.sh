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
  'verify-release-readiness-report.mjs' \
  'Release-readiness report verified:' \
  'JSON.parse' \
  'expectedCheckIds' \
  'duplicate id' \
  'live_gates_retained' \
  'parsed and semantically verified'; do
  grep -Fq -- "$required" scripts/verify-release-readiness-report.sh scripts/verify-release-readiness-report.mjs || {
    echo "readiness-report verifier must retain: $required" >&2
    exit 1
  }
done

fixture="$(mktemp)"
malformed_fixture="$(mktemp)"
duplicate_fixture="$(mktemp)"
trap 'rm -f "$fixture" "$malformed_fixture" "$duplicate_fixture"' EXIT
cat >"$fixture" <<'JSON'
{"schema_version":2,"scope":"local deterministic release readiness","result":"passed","commit_sha":"0123456789abcdef0123456789abcdef01234567","commit_timestamp_unix":1,"working_tree_clean":true,"checks":[{"id":"source_and_policy","command":"test","expected":"test","observed":"passed","evidence_boundary":"local","known_limitation":"no live","recovery":"stop"},{"id":"shared_schema_safety","command":"test","expected":"test","observed":"passed","evidence_boundary":"local","known_limitation":"no live","recovery":"stop"},{"id":"local_resilience","command":"test","expected":"test","observed":"passed","evidence_boundary":"local","known_limitation":"no live","recovery":"stop"},{"id":"native_builds","command":"test","expected":"test","observed":"passed","evidence_boundary":"local","known_limitation":"no live","recovery":"stop"},{"id":"browser_console","command":"test","expected":"test","observed":"passed","evidence_boundary":"local","known_limitation":"no live","recovery":"stop"}],"evidence_locations":["local"],"live_gates_retained":["No live Framework hardware qualification","No live dual-replica failover, certificate revocation, or isolated restore acceptance","No real Supabase magic-link, approved-backend streaming, or tagged release-attestation verification"],"recovery":"stop"}
JSON
./scripts/verify-release-readiness-report.sh "$fixture"

printf '{' >"$malformed_fixture"
if ./scripts/verify-release-readiness-report.sh "$malformed_fixture" >/tmp/nodescope-readiness-malformed.out 2>&1; then
  echo "readiness-report verifier accepted malformed JSON" >&2
  exit 1
fi
grep -Fq 'cannot parse JSON' /tmp/nodescope-readiness-malformed.out

sed '0,/"id":"browser_console"/s//"id":"native_builds"/' "$fixture" >"$duplicate_fixture"
if ./scripts/verify-release-readiness-report.sh "$duplicate_fixture" >/tmp/nodescope-readiness-duplicate.out 2>&1; then
  echo "readiness-report verifier accepted duplicate check identifiers" >&2
  exit 1
fi
grep -Fq 'duplicate id' /tmp/nodescope-readiness-duplicate.out

echo "Release-readiness report contract passed."
