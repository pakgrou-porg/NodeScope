#!/usr/bin/env bash
# Validate one complete NodeScope server replica before an approved cloud-canary deployment.
# This script never starts, stops, pulls, or builds containers.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
compose_dir="$repository_root/deploy/compose"
compose_file="$compose_dir/compose.yaml"
replica_env="$compose_dir/replica.env"

usage() {
  cat >&2 <<'USAGE'
usage: scripts/preflight-cloud-replica-compose.sh [--replica-env PATH]

Validate the currently staged replica environment, protected runtime secrets,
certificate mount, endpoint ordering, and Compose configuration. The script
does not deploy containers or print secret values.
USAGE
  exit 2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --replica-env)
      [[ $# -ge 2 ]] || usage
      replica_env="$2"
      shift 2
      ;;
    *) usage ;;
  esac
done

command -v docker >/dev/null 2>&1 || { echo "docker is required for replica preflight" >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Docker Compose v2 is required for replica preflight" >&2; exit 1; }
[[ -f "$compose_file" && ! -L "$compose_file" ]] || { echo "compose file must be a regular file" >&2; exit 1; }
[[ -f "$replica_env" && ! -L "$replica_env" ]] || { echo "replica environment file must be a regular file" >&2; exit 1; }

value_for() {
  local key="$1"
  local line
  line="$(grep -E "^${key}=" "$replica_env" | tail -n 1 || true)"
  [[ -n "$line" ]] || { echo "missing required replica setting: $key" >&2; exit 1; }
  printf '%s' "${line#*=}"
}

replica_id="$(value_for NODESCOPE_REPLICA_ID)"
replica_role="$(value_for NODESCOPE_REPLICA_ROLE)"
primary_endpoint="$(value_for NODESCOPE_PRIMARY_ENDPOINT)"
secondary_endpoint="$(value_for NODESCOPE_SECONDARY_ENDPOINT)"
runtime_secrets="$(value_for NODESCOPE_RUNTIME_SECRETS_FILE)"
certificate_directory="$(value_for NODESCOPE_CERTIFICATE_DIRECTORY)"
runtime_directory="$(value_for NODESCOPE_RUNTIME_DIRECTORY)"

[[ "$replica_role" == "preferred" || "$replica_role" == "secondary" ]] || { echo "replica role must be preferred or secondary" >&2; exit 1; }
[[ -n "$replica_id" ]] || { echo "replica ID must not be blank" >&2; exit 1; }
[[ "$primary_endpoint" != "$secondary_endpoint" ]] || { echo "replica endpoints must be distinct" >&2; exit 1; }

for endpoint in "$primary_endpoint" "$secondary_endpoint"; do
  [[ "$endpoint" =~ ^https://[^/@?#:]+(:[0-9]+)?$ ]] || { echo "replica endpoint must be credential-free HTTPS host:port" >&2; exit 1; }
done

for forbidden_key in NODESCOPE_RUNTIME_DB_PASSWORD NODESCOPE_SERVICE_ROLE_KEY NODESCOPE_AGENT_TOKEN NODESCOPE_CA_PRIVATE_KEY; do
  ! grep -qE "^${forbidden_key}=" "$replica_env" || { echo "replica environment must not contain $forbidden_key" >&2; exit 1; }
done

[[ -f "$runtime_secrets" && ! -L "$runtime_secrets" ]] || { echo "protected runtime secrets file must be a regular file" >&2; exit 1; }
[[ -d "$certificate_directory" && ! -L "$certificate_directory" ]] || { echo "certificate directory must be a real directory" >&2; exit 1; }
[[ -f "$certificate_directory/server.crt" && -f "$certificate_directory/server.key" ]] || { echo "server certificate and private key are required" >&2; exit 1; }
[[ -d "$runtime_directory" && ! -L "$runtime_directory" ]] || { echo "runtime directory must be a real directory" >&2; exit 1; }

for protected_path in "$runtime_secrets" "$certificate_directory/server.key"; do
  mode="$(stat -c '%a' "$protected_path")"
  [[ "${mode: -1}" == "0" ]] || { echo "protected file must not be world-readable: $protected_path" >&2; exit 1; }
done

# The canonical deployment places replica.env beside compose.yaml. A separate
# path is allowed only for inspection; full Compose expansion requires staging
# that exact protected file on the target host before this command is repeated.
[[ "$replica_env" == "$compose_dir/replica.env" ]] || { echo "preflight passed basic checks; stage replica.env beside compose.yaml for final Compose expansion" >&2; exit 1; }

(
  cd "$compose_dir"
  docker compose -f "$compose_file" config --quiet
)

echo "Cloud replica compose preflight passed: replica=$replica_id role=$replica_role"
