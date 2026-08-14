#!/usr/bin/env bash
# Enforce that cloud replica preflight is validation-only and remains fail closed.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

script="scripts/preflight-cloud-replica-compose.sh"
compose="deploy/compose/compose.yaml"

[[ -x "$script" || -f "$script" ]] || { echo "missing cloud replica preflight" >&2; exit 1; }
for required in \
  'docker compose version' \
  'NODESCOPE_PRIMARY_ENDPOINT' \
  'NODESCOPE_SECONDARY_ENDPOINT' \
  'NODESCOPE_RUNTIME_SECRETS_FILE' \
  'NODESCOPE_CERTIFICATE_DIRECTORY' \
  'NODESCOPE_RUNTIME_DB_PASSWORD' \
  'server.key' \
  'config --quiet'; do
  grep -Fq "$required" "$script" || { echo "preflight missing required boundary: $required" >&2; exit 1; }
done
! grep -Eq 'docker compose .*\b(up|down|pull|build)\b' "$script" || { echo "preflight must not deploy or mutate containers" >&2; exit 1; }
grep -Fq 'read_only: true' "$compose" || { echo "compose must keep server root filesystem read-only" >&2; exit 1; }
grep -Fq 'no-new-privileges:true' "$compose" || { echo "compose must retain no-new-privileges" >&2; exit 1; }
grep -Fq 'nodescope-probe' "$compose" || { echo "compose must retain health probe" >&2; exit 1; }

echo "Cloud replica compose preflight contract passed."
