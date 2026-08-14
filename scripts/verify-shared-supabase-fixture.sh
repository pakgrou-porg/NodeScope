#!/usr/bin/env bash
# Run the disposable shared-Supabase isolation gate. It creates only controlled
# fixtures, verifies runtime/migrator/agent boundaries, rolls back the selected
# migration, and removes all fixtures before returning success.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 supabase/migrations/NNNN_name.sql" >&2
  exit 2
fi

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"
migration_file="$1"
migration_basename="$(basename -- "$migration_file")"
expected_migration="supabase/migrations/$migration_basename"
if [[ "$migration_file" != "$expected_migration" ]] || ! [[ "$migration_basename" =~ ^[0-9]{4}_[A-Za-z0-9][A-Za-z0-9._-]*\.sql$ ]] || [[ ! -f "$migration_file" || -L "$migration_file" ]]; then
  echo "migration must be a direct regular source-controlled SQL file in supabase/migrations" >&2
  exit 2
fi
if ! git ls-files --error-unmatch -- "$migration_file" >/dev/null 2>&1 || ! git diff --quiet -- "$migration_file"; then
  echo "migration must be source-controlled and clean" >&2
  exit 2
fi

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required}"
: "${NODESCOPE_RUNTIME_DB_PASSWORD:?NODESCOPE_RUNTIME_DB_PASSWORD is required}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"
for dependency in git go psql; do
  if ! command -v "$dependency" >/dev/null 2>&1; then
    echo "required dependency is unavailable: $dependency" >&2
    exit 2
  fi
done

host="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -E 's#^[a-z]+://[^@]*@([^:/?]+).*#\1#')"
port="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -nE 's#^[a-z]+://[^@]*@[^:/?]+:([0-9]+).*#\1#p')"
port="${port:-5432}"
migration_version="${migration_basename%.sql}"

primary_psql=(psql "$NODESCOPE_SUPABASE_DB_URL" --no-psqlrc -q -v ON_ERROR_STOP=1)
runtime_psql=(env PGCONNECT_TIMEOUT=10 PGHOST="$host" PGPORT="$port" PGDATABASE=postgres PGUSER=nodescope_runtime_login PGPASSWORD="$NODESCOPE_RUNTIME_DB_PASSWORD" PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1)
migrator_psql=(env PGCONNECT_TIMEOUT=10 PGHOST="$host" PGPORT="$port" PGDATABASE=postgres PGUSER=nodescope_migrate_login PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1)

cleanup() {
  "${primary_psql[@]}" -f supabase/isolation/drop_nodescope_rls_fixture.sql >/dev/null 2>&1 || true
}
trap cleanup EXIT

./scripts/verify-sibling-denials.sh
"${primary_psql[@]}" -f supabase/isolation/create_nodescope_rls_fixture.sql

# Runtime must operate through the approved RLS policy, but must not write an
# actor row for any other effective role. The allowed probe is rolled back.
"${runtime_psql[@]}" -c "begin; set role nodescope_runtime; insert into nodescope.rls_isolation_fixture(actor, payload) values ('nodescope_runtime', 'fixture') ; select count(*) from nodescope.rls_isolation_fixture; rollback;" >/dev/null
if "${runtime_psql[@]}" -c "begin; set role nodescope_runtime; insert into nodescope.rls_isolation_fixture(actor, payload) values ('other', 'blocked'); rollback;" >/tmp/nodescope-rls-fixture-denial.out 2>&1; then
  echo "UNSAFE: runtime role bypassed the disposable NodeScope RLS policy" >&2
  exit 1
fi

# Catalog-only verifier proves ownership, RLS, and sibling privilege boundaries.
"${migrator_psql[@]}" -f supabase/isolation/verify_shared_isolation.sql >/dev/null

# Agents do not have a database login. Their boundary is authenticated HTTPS
# ingestion, enforced by rejecting every direct database configuration key.
go test ./internal/agent -run '^TestLoadConfigRejectsDirectDatabaseConfiguration$' -count=1 >/dev/null

precondition="$(${migrator_psql[@]} -At -c "begin read only; set role nodescope_owner; select coalesce((select 'recorded' from nodescope.schema_migrations where version = '$migration_version'), 'unrecorded'); rollback;")"
if [[ "$precondition" != "unrecorded" ]]; then
  echo "migration must be unrecorded for a meaningful rollback preflight" >&2
  exit 1
fi

"${migrator_psql[@]}" -c "begin;" -f "$migration_file" -c "rollback;"
postcondition="$(${migrator_psql[@]} -At -c "begin read only; set role nodescope_owner; select coalesce((select 'recorded' from nodescope.schema_migrations where version = '$migration_version'), 'unrecorded'); rollback;")"
if [[ "$postcondition" != "unrecorded" ]]; then
  echo "UNSAFE: rollback preflight persisted a migration ledger entry" >&2
  exit 1
fi

cleanup
trap - EXIT
printf 'Shared-Supabase disposable fixture gate passed: runtime RLS, sibling denial, agent database boundary, catalog isolation, and rollback-only migration preflight.\n'
