#!/usr/bin/env bash
# Verify migration application rejects paths that escape the reviewed migration set.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

if grep -Fq 'psql "$NODESCOPE_SUPABASE_DB_URL"' scripts/apply-nodescope-migration.sh; then
  echo "migration application must not reuse an opaque database URL credential for post-apply verification" >&2
  exit 1
fi
if ! grep -Fq 'PGUSER=nodescope_migrate_login' scripts/apply-nodescope-migration.sh; then
  echo "migration application must use the dedicated migrator login" >&2
  exit 1
fi

assert_rejected() {
  local description="$1"
  local path="$2"
  if ./scripts/apply-nodescope-migration.sh "$path" >/tmp/nodescope-migration-contract.out 2>&1; then
    echo "expected $description to be rejected" >&2
    exit 1
  fi
}

temporary_dir="$(mktemp -d)"
temporary_migration="supabase/migrations/9999_contract-escape.sql"
trap 'rm -rf "$temporary_dir" "$temporary_migration" /tmp/nodescope-migration-contract.out' EXIT
printf '%s\n' 'select 1;' > "$temporary_dir/outside.sql"
ln -s "$repository_root/$temporary_dir/outside.sql" "$temporary_migration"

assert_rejected "traversal path" "supabase/migrations/0001_nodescope_foundation.sql/../0001_nodescope_foundation.sql"
assert_rejected "symlink migration" "$temporary_migration"
rm -f "$temporary_migration"
printf '%s\n' 'select 1;' > "$temporary_migration"
assert_rejected "untracked migration" "$temporary_migration"

if env -u NODESCOPE_SUPABASE_DB_URL -u NODESCOPE_MIGRATOR_DB_PASSWORD \
  ./scripts/apply-nodescope-migration.sh supabase/migrations/0001_nodescope_foundation.sql >/tmp/nodescope-migration-contract.out 2>&1; then
  echo "expected clean tracked migration to stop before live application without database credentials" >&2
  exit 1
fi
if ! grep -q 'NODESCOPE_SUPABASE_DB_URL is required' /tmp/nodescope-migration-contract.out; then
  echo "clean tracked migration did not pass local path-containment validation" >&2
  exit 1
fi

echo "NodeScope migration application path-containment contract passed."
