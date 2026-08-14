#!/usr/bin/env bash
# Ensure the catalog-only isolation verifier remains usable by the dedicated
# least-privilege migrator login, which intentionally lacks auth-schema usage.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

verifier=supabase/isolation/verify_shared_isolation.sql
if grep -Fq "'auth.users'" "$verifier"; then
  echo "shared-isolation verifier must not resolve auth.users by text under the migrator login" >&2
  exit 1
fi
for required in 'namespace.nspname = '\''auth'\''' 'relation.relname = '\''users'\''' 'has_table_privilege('\''nodescope_runtime'\'', relation.oid, '\''SELECT'\'')'; do
  if ! grep -Fq "$required" "$verifier"; then
    echo "shared-isolation verifier must retain catalog-only auth-user denial checks" >&2
    exit 1
  fi
done

echo "Shared-isolation verifier least-privilege contract passed."
