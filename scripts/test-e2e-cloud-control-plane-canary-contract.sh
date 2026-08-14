#!/usr/bin/env bash
# Keep the cloud canary scoped to control-plane proof and prevent it from being
# misrepresented as Framework hardware qualification.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

canary=scripts/e2e-cloud-control-plane-canary.sh
for required in \
  'NODESCOPE_REQUIRE_AGENT_MTLS=true' \
  '--tlsv1.3' \
  '"status":"authenticated"' \
  '"status":"duplicate"' \
  'fresh|cloud-canary|authenticated cloud control-plane evidence|receipt_after_observation' \
  'nodescope.ingest_receipts' \
  'delete from nodescope.hosts' \
  'does not qualify Framework hardware'; do
  if ! grep -Fq -- "$required" "$canary"; then
    echo "cloud control-plane canary must retain $required" >&2
    exit 1
  fi
done

echo "Cloud control-plane canary contract passed."
