#!/usr/bin/env bash
# Verify local console RBAC and route-loading contracts without creating users,
# sending magic links, or contacting Supabase Auth.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

if ! command -v pnpm >/dev/null 2>&1; then
  echo "required dependency is unavailable: pnpm" >&2
  exit 2
fi

pnpm vitest run server/nodescope/router.test.ts client/src/App.routes.test.ts

cat <<'JSON'
{"schema_version":1,"scope":"local console RBAC readiness","result":"passed","controls":{"viewer_fleet_read":"locally validated","viewer_host_alert_read":"locally validated","viewer_configuration_mutation_denial":"locally validated","administrator_configuration_mutation_audit":"locally validated","runtime_approval_authorization":"locally validated","route_loading":"locally validated","supabase_magic_link":"live environment gate","invite_only_user_lifecycle":"live environment gate","browser_session_e2e":"live environment gate","degraded_replica_browser_e2e":"live environment gate"},"recovery":"No user, session, or invitation is created. If a later live E2E fails, disable the test invite, invalidate the test session, and restore the prior callback configuration before retrying."}
JSON
