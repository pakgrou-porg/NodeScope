#!/usr/bin/env bash
# Ensure the Framework Portainer package retains replica hardening and safe secret boundaries.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
stack="$repository_root/deploy/portainer/framework-stack.yaml"
stack_env="$repository_root/deploy/portainer/framework-stack.env.example"
runtime_env="$repository_root/deploy/portainer/runtime.env.example"
guide="$repository_root/docs/operations/framework-portainer-deployment.md"

for file in "$stack" "$stack_env" "$runtime_env" "$guide"; do
  [[ -f "$file" ]] || { echo "missing Framework Portainer deployment artifact: $file" >&2; exit 1; }
done

for required in 'read_only: true' 'no-new-privileges:true' 'cap_drop:' '- ALL' 'NODESCOPE_TLS_CERT_PATH' 'NODESCOPE_TLS_KEY_PATH' 'nodescope-probe' 'NODESCOPE_PRIMARY_ENDPOINT' 'NODESCOPE_SECONDARY_ENDPOINT'; do
  grep -Fq -- "$required" "$stack" || { echo "Framework stack is missing required control: $required" >&2; exit 1; }
done

for forbidden in 'NODESCOPE_RUNTIME_DB_PASSWORD=' 'NODESCOPE_SERVICE_ROLE_KEY=' 'NODESCOPE_AGENT_TOKEN=' 'PRIVATE KEY'; do
  ! grep -Fq -- "$forbidden" "$stack_env" || { echo "Framework Portainer environment template must not contain $forbidden" >&2; exit 1; }
done

grep -Fq 'NODESCOPE_RUNTIME_DB_PASSWORD=REPLACE_WITH_NODESCOPE_RUNTIME_LOGIN_PASSWORD' "$runtime_env" || { echo "Framework runtime secret template must require the runtime database password" >&2; exit 1; }
grep -Fq 'Git repository stack' "$guide" || { echo "Framework guide must require the reviewed Git stack path" >&2; exit 1; }
grep -Fq 'preflight-cloud-replica-compose.sh' "$guide" || { echo "Framework guide must require the preflight" >&2; exit 1; }
grep -Fq '## Rollback' "$guide" || { echo "Framework guide must include rollback" >&2; exit 1; }

echo "Framework Portainer stack contract passed."
