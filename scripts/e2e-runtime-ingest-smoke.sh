#!/usr/bin/env bash
# End-to-end verification for the dedicated NodeScope runtime database role.
# It creates an ephemeral NodeScope host/agent, submits one telemetry batch,
# verifies latest-state persistence, and removes all fixture rows on exit.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required}"
: "${NODESCOPE_RUNTIME_DB_PASSWORD:?NODESCOPE_RUNTIME_DB_PASSWORD is required}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"

host_id="$(cat /proc/sys/kernel/random/uuid)"
agent_id="$(cat /proc/sys/kernel/random/uuid)"
token="nodescope-smoke-$(cat /proc/sys/kernel/random/uuid)"
digest="$(printf '%s' "$token" | sha256sum | awk '{print $1}')"
port="18081"
server_log="$(mktemp)"
server_pid=""

migrator_psql() {
  PGHOST=db.vafiuhbqldcogrmnqbjw.supabase.co PGPORT=5432 PGDATABASE=postgres \
    PGUSER=nodescope_migrate_login PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" \
    PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1 "$@"
}

cleanup() {
  if [[ -n "$server_pid" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  migrator_psql -c "begin; set role nodescope_owner; delete from nodescope.hosts where id = '${host_id}'::uuid; commit;" >/dev/null 2>&1 || true
  rm -f "$server_log"
}
trap cleanup EXIT

migrator_psql -c "begin; set role nodescope_owner; insert into nodescope.hosts(id, slug, display_name, platform) values ('${host_id}'::uuid, 'smoke-${host_id:0:8}', 'Smoke Host', 'test'); insert into nodescope.agents(id, host_id, display_name, credential_digest, credential_hint) values ('${agent_id}'::uuid, '${host_id}'::uuid, 'Smoke Agent', decode('${digest}', 'hex'), 'smoke'); commit;"

# Build outside the test host so the only long-lived process is short-lived.
go build -o /tmp/nodescope-server ./cmd/nodescope-server
NODESCOPE_ALLOW_JSON_INGEST=true NODESCOPE_REPLICA_ID=framework NODESCOPE_REPLICA_ROLE=preferred \
NODESCOPE_LISTEN_ADDRESS="127.0.0.1:${port}" \
NODESCOPE_PRIMARY_ENDPOINT=https://10.116.2.145:8443 \
NODESCOPE_SECONDARY_ENDPOINT=https://10.116.2.56:8443 \
NODESCOPE_SUPABASE_URL=https://vafiuhbqldcogrmnqbjw.supabase.co \
NODESCOPE_RUNTIME_DB_HOST=db.vafiuhbqldcogrmnqbjw.supabase.co \
NODESCOPE_RUNTIME_DB_PASSWORD="$NODESCOPE_RUNTIME_DB_PASSWORD" \
/tmp/nodescope-server >"$server_log" 2>&1 &
server_pid="$!"
sleep 2

curl --fail --silent --show-error \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${token}" \
  --data "{\"schemaVersion\":1,\"codec\":\"protobuf+zstd\",\"agentId\":\"${agent_id}\",\"hostId\":\"${host_id}\",\"bootId\":\"smoke-boot\",\"sequence\":1,\"observedAt\":\"2026-07-22T12:00:00Z\",\"sampleCount\":1,\"metricValueCount\":1,\"uncompressedBytes\":128,\"compressedBytes\":64,\"checksumSha256\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"samples\":[{\"deviceId\":\"cpu-0\",\"metric\":{\"name\":\"cpu.utilization\",\"unit\":\"percent\",\"value\":42.5,\"quality\":\"fresh\",\"source\":\"smoke\",\"semantics\":\"smoke-only host CPU utilization\",\"observedAt\":\"2026-07-22T12:00:00Z\"}}]}" \
  "http://127.0.0.1:${port}/api/v1/ingest" | grep -q '"status":"accepted"'

count="$(migrator_psql --tuples-only --no-align -c "set role nodescope_owner; select count(*) from nodescope.metric_latest where host_id = '${host_id}'::uuid and metric_name = 'cpu.utilization';")"
test "$count" = "1"

printf 'Runtime end-to-end ingestion smoke test passed.\n'
