#!/usr/bin/env bash
# Keep the detailed Framework auxiliary-agent procedure safe to delegate.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
runbook="$repository_root/docs/operations/framework-auxiliary-agent-runbook.md"

for required in \
  'without starting a container' \
  'Do NOT start, stop, build, pull, remove, or deploy containers' \
  'Do NOT run agent enrollment, credential rotation, SQL mutations, migrations, certificate issuance, revocation, or backup commands' \
  'Do NOT read, print, upload, or copy secret values' \
  'preflight-cloud-replica-compose.sh' \
  'FRAMEWORK_NODESCOPE_PREFLIGHT' \
  'confirmation_required_next' \
  'Deploy the stack' \
  'It must not weaken `read_only`, `cap_drop`, `no-new-privileges`'; do
  grep -Fq -- "$required" "$runbook" || { echo "Framework auxiliary-agent runbook missing required boundary: $required" >&2; exit 1; }
done

echo "Framework auxiliary-agent runbook contract passed."
