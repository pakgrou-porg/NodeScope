#!/usr/bin/env bash
# Keep the live disposable fixture gate fail-closed while credential-free CI
# checks its composition and ordering.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

fixture=scripts/verify-shared-supabase-fixture.sh
for required in \
  'verify-sibling-denials.sh' \
  'create_nodescope_rls_fixture.sql' \
  'drop_nodescope_rls_fixture.sql' \
  'for dependency in git go psql' \
  'set role nodescope_runtime' \
  'verify_shared_isolation.sql' \
  'TestLoadConfigRejectsDirectDatabaseConfiguration' \
  'rollback;' \
  'migration must be unrecorded'; do
  if ! grep -Fq "$required" "$fixture"; then
    echo "shared-Supabase fixture gate must retain $required" >&2
    exit 1
  fi
done

if ! grep -Fq 'force row level security' supabase/isolation/create_nodescope_rls_fixture.sql; then
  echo "RLS fixture must force row-level security" >&2
  exit 1
fi
if ! grep -Fq 'actor = current_user' supabase/isolation/create_nodescope_rls_fixture.sql; then
  echo "RLS fixture must bind runtime rows to the effective role" >&2
  exit 1
fi

echo "Shared-Supabase disposable fixture contract passed."
