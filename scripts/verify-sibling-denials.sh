#!/usr/bin/env bash
# Proves both NodeScope database logins are denied read, write, DDL, and routine
# modification against a disposable sibling-application schema. This script is
# a required preflight before every future production NodeScope migration.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required}"
: "${NODESCOPE_RUNTIME_DB_PASSWORD:?NODESCOPE_RUNTIME_DB_PASSWORD is required}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"

primary_psql=(psql "$NODESCOPE_SUPABASE_DB_URL" --no-psqlrc -q -v ON_ERROR_STOP=1)
"${primary_psql[@]}" -f supabase/isolation/create_sibling_fixture.sql
cleanup() {
  "${primary_psql[@]}" -f supabase/isolation/drop_sibling_fixture.sql >/dev/null 2>&1 || true
}
trap cleanup EXIT

expect_denied() {
  local login="$1"
  local password="$2"
  local assumed_role="$3"
  local label="$4"
  local statement="$5"
  local output

  if output=$(PGHOST=db.vafiuhbqldcogrmnqbjw.supabase.co \
    PGPORT=5432 PGDATABASE=postgres PGUSER="$login" PGPASSWORD="$password" \
    PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1 \
    -c "begin; set role ${assumed_role}; ${statement}; rollback;" 2>&1); then
    printf 'UNSAFE: %s permitted %s against sibling fixture\n' "$login" "$label" >&2
    exit 1
  fi
  printf 'denied: %s %s\n' "$login" "$label"
}

verify_login() {
  local login="$1"
  local password="$2"
  local assumed_role="$3"

  expect_denied "$login" "$password" "$assumed_role" "SELECT" "select * from nodescope_isolation_fixture.documents"
  expect_denied "$login" "$password" "$assumed_role" "INSERT" "insert into nodescope_isolation_fixture.documents(title) values ('blocked')"
  expect_denied "$login" "$password" "$assumed_role" "UPDATE" "update nodescope_isolation_fixture.documents set title = 'blocked'"
  expect_denied "$login" "$password" "$assumed_role" "DELETE" "delete from nodescope_isolation_fixture.documents"
  expect_denied "$login" "$password" "$assumed_role" "ALTER TABLE" "alter table nodescope_isolation_fixture.documents add column blocked boolean"
  expect_denied "$login" "$password" "$assumed_role" "DROP TABLE" "drop table nodescope_isolation_fixture.documents"
  expect_denied "$login" "$password" "$assumed_role" "FUNCTION EXECUTE" "select nodescope_isolation_fixture.count_documents()"
  expect_denied "$login" "$password" "$assumed_role" "FUNCTION REPLACE" "create or replace function nodescope_isolation_fixture.count_documents() returns bigint language sql as \$\$ select 0 \$\$"
}

verify_login nodescope_runtime_login "$NODESCOPE_RUNTIME_DB_PASSWORD" nodescope_runtime
verify_login nodescope_migrate_login "$NODESCOPE_MIGRATOR_DB_PASSWORD" nodescope_owner

printf 'Shared-project sibling-schema denial gate passed.\n'
