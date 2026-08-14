#!/usr/bin/env bash
# Keep the local RBAC rehearsal distinct from live Supabase Auth/browser proof.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

harness=scripts/rehearse-console-rbac-local.sh
for required in \
  'server/nodescope/router.test.ts' \
  'client/src/App.routes.test.ts' \
  'viewer_configuration_mutation_denial":"locally validated' \
  'supabase_magic_link":"live environment gate' \
  'browser_session_e2e":"live environment gate' \
  'degraded_replica_browser_e2e":"live environment gate' \
  'No user, session, or invitation is created'; do
  if ! grep -Fq "$required" "$harness"; then
    echo "console RBAC readiness rehearsal must retain $required" >&2
    exit 1
  fi
done

echo "Console RBAC readiness contract passed."
