#!/usr/bin/env bash
# Apply one future NodeScope migration only after isolation gates pass.
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
if [[ "$migration_file" != "$expected_migration" ]] || ! [[ "$migration_basename" =~ ^[0-9]{4}_[A-Za-z0-9][A-Za-z0-9._-]*\.sql$ ]]; then
  echo "migration must be a direct source-controlled SQL file in supabase/migrations" >&2
  exit 2
fi
if [[ ! -f "$migration_file" || -L "$migration_file" ]]; then
  echo "migration must be a regular file, not a missing path or symlink" >&2
  exit 2
fi
if ! git ls-files --error-unmatch -- "$migration_file" >/dev/null 2>&1; then
  echo "migration must be source-controlled" >&2
  exit 2
fi
if ! git diff --quiet -- "$migration_file"; then
  echo "migration must be clean before application" >&2
  exit 2
fi

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required for post-apply verification}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"

host="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -E 's#^[a-z]+://[^@]*@([^:/?]+).*#\1#')"
port="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -nE 's#^[a-z]+://[^@]*@[^:/?]+:([0-9]+).*#\1#p')"
port="${port:-5432}"

# Gate 1: both NodeScope logins must be denied all access to a disposable
# sibling-application fixture before any NodeScope migration is even parsed.
./scripts/verify-sibling-denials.sh

# Gate 2: run the exact migration in a single rolled-back transaction through
# the dedicated migration login. No production DDL may occur in this step.
preflight="$(mktemp)"
trap 'rm -f "$preflight"' EXIT
absolute_migration="$repository_root/$migration_file"
printf '%s\n%s\n%s\n%s\n' '\set ON_ERROR_STOP on' 'begin;' "\\i $absolute_migration" 'rollback;' > "$preflight"
PGHOST="$host" PGPORT="$port" PGDATABASE=postgres PGUSER=nodescope_migrate_login \
  PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" PGSSLMODE=require \
  psql --no-psqlrc -q -v ON_ERROR_STOP=1 -f "$preflight"

# Apply only after both gates pass.
PGHOST="$host" PGPORT="$port" PGDATABASE=postgres PGUSER=nodescope_migrate_login \
  PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" PGSSLMODE=require \
  psql --no-psqlrc -v ON_ERROR_STOP=1 -f "$migration_file"

# Confirm the post-apply shared-project privilege boundary through the same
# dedicated migrator connection used for preflight and application.
PGHOST="$host" PGPORT="$port" PGDATABASE=postgres PGUSER=nodescope_migrate_login \
  PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" PGSSLMODE=require \
  psql --no-psqlrc -q -v ON_ERROR_STOP=1 \
  -f supabase/isolation/verify_shared_isolation.sql

echo "Applied NodeScope migration after all isolation gates: $migration_file"
