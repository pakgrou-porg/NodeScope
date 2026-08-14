#!/usr/bin/env bash
# Run only after explicit authorization to create and remove the disposable
# sibling fixture. Proves routine database identities cannot replace owner use.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

: "${NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL:?NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL is required}"
: "${NODESCOPE_VERIFIER_DATABASE_URL:?NODESCOPE_VERIFIER_DATABASE_URL is required}"
: "${NODESCOPE_STORAGE_AUDITOR_DATABASE_URL:?NODESCOPE_STORAGE_AUDITOR_DATABASE_URL is required}"

for dependency in psql; do
  command -v "$dependency" >/dev/null 2>&1 || { echo "required dependency is unavailable: $dependency" >&2; exit 2; }
done

admin_psql=(psql "$NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL" --no-psqlrc -q -v ON_ERROR_STOP=1)
verifier_psql=(psql "$NODESCOPE_VERIFIER_DATABASE_URL" --no-psqlrc -q -v ON_ERROR_STOP=1)
storage_auditor_psql=(psql "$NODESCOPE_STORAGE_AUDITOR_DATABASE_URL" --no-psqlrc -q -v ON_ERROR_STOP=1)

"${admin_psql[@]}" -f supabase/isolation/create_sibling_fixture.sql
cleanup() {
  "${admin_psql[@]}" -f supabase/isolation/drop_sibling_fixture.sql >/dev/null 2>&1 || true
}
trap cleanup EXIT

expect_denied() {
  local role="$1"
  local label="$2"
  local statement="$3"
  shift 3
  local -a command=("$@")

  if "${command[@]}" -c "begin; set role ${role}; ${statement}; rollback;" >/tmp/nodescope-operational-role-denial.out 2>&1; then
    printf 'UNSAFE: %s permitted %s\n' "$role" "$label" >&2
    exit 1
  fi
  printf 'denied: %s %s\n' "$role" "$label"
}

expect_allowed() {
  local role="$1"
  local label="$2"
  local statement="$3"
  shift 3
  local -a command=("$@")

  "${command[@]}" -c "begin read only; set role ${role}; ${statement}; rollback;" >/dev/null
  printf 'allowed: %s %s\n' "$role" "$label"
}

for statement in \
  "select * from nodescope.hosts" \
  "select * from nodescope.metric_samples" \
  "insert into nodescope.hosts(slug, display_name, platform, address) values ('denied', 'denied', 'denied', '127.0.0.1')" \
  "alter table nodescope.hosts add column denied boolean" \
  "select * from nodescope_isolation_fixture.documents" \
  "select nodescope_isolation_fixture.count_documents()"; do
  expect_denied nodescope_verifier "direct-or-sibling access" "$statement" "${verifier_psql[@]}"
  expect_denied nodescope_storage_auditor "direct-or-sibling access" "$statement" "${storage_auditor_psql[@]}"
done

expect_allowed nodescope_verifier "host status function" \
  "select count(*) from nodescope.host_ingestion_status('__nodescope_denied_probe__')" "${verifier_psql[@]}"
expect_allowed nodescope_verifier "fleet status function" \
  "select count(*) from nodescope.fleet_ingestion_status()" "${verifier_psql[@]}"
expect_denied nodescope_verifier "storage evidence function" \
  "select count(*) from nodescope.storage_probe_evidence('__nodescope_denied_probe__', now() - interval '1 hour')" "${verifier_psql[@]}"

expect_allowed nodescope_storage_auditor "storage evidence function" \
  "select count(*) from nodescope.storage_probe_evidence('__nodescope_denied_probe__', now() - interval '1 hour')" "${storage_auditor_psql[@]}"
expect_denied nodescope_storage_auditor "host status function" \
  "select count(*) from nodescope.host_ingestion_status('__nodescope_denied_probe__')" "${storage_auditor_psql[@]}"
expect_denied nodescope_storage_auditor "fleet status function" \
  "select count(*) from nodescope.fleet_ingestion_status()" "${storage_auditor_psql[@]}"

cleanup
trap - EXIT
printf 'Operational verifier and storage-auditor denial gate passed.\n'
