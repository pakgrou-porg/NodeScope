#!/usr/bin/env bash
# Verifies summary/protective capacity behavior with real runtime-role ingestion.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required}"
: "${NODESCOPE_RUNTIME_DB_PASSWORD:?NODESCOPE_RUNTIME_DB_PASSWORD is required}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"

host_id="$(cat /proc/sys/kernel/random/uuid)"
agent_id="$(cat /proc/sys/kernel/random/uuid)"
token="nodescope-capacity-smoke-$(cat /proc/sys/kernel/random/uuid)"
digest="$(printf '%s' "$token" | sha256sum | awk '{print $1}')"
port="18082"
server_pid=""

migrator_psql() {
  PGHOST=db.vafiuhbqldcogrmnqbjw.supabase.co PGPORT=5432 PGDATABASE=postgres \
    PGUSER=nodescope_migrate_login PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" \
    PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1 "$@"
}
cleanup() {
  if [[ -n "$server_pid" ]]; then kill "$server_pid" 2>/dev/null || true; wait "$server_pid" 2>/dev/null || true; fi
  migrator_psql -c "begin; set role nodescope_owner; delete from nodescope.hosts where id = '${host_id}'::uuid; delete from nodescope.capacity_status where singleton = true; commit;" >/dev/null 2>&1 || true
}
trap cleanup EXIT

migrator_psql -c "begin; set role nodescope_owner; insert into nodescope.hosts(id, slug, display_name, platform) values ('${host_id}'::uuid, 'capacity-${host_id:0:8}', 'Capacity Smoke Host', 'test'); insert into nodescope.agents(id, host_id, display_name, credential_digest, credential_hint) values ('${agent_id}'::uuid, '${host_id}'::uuid, 'Capacity Smoke Agent', decode('${digest}', 'hex'), 'capacity-smoke'); insert into nodescope.capacity_status(singleton, used_bytes, quota_bytes, mode, accept_raw_batches, accept_summary_rollups, detail, source) values (true, 960, 1000, 'protective', false, true, 'test capacity gate', 'smoke') on conflict (singleton) do update set used_bytes = excluded.used_bytes, quota_bytes = excluded.quota_bytes, mode = excluded.mode, accept_raw_batches = excluded.accept_raw_batches, accept_summary_rollups = excluded.accept_summary_rollups, detail = excluded.detail, source = excluded.source; commit;"

go build -o /tmp/nodescope-server ./cmd/nodescope-server
NODESCOPE_ALLOW_JSON_INGEST=true NODESCOPE_REPLICA_ID=framework NODESCOPE_REPLICA_ROLE=preferred NODESCOPE_LISTEN_ADDRESS="127.0.0.1:${port}" NODESCOPE_PRIMARY_ENDPOINT=https://10.116.2.145:8443 NODESCOPE_SECONDARY_ENDPOINT=https://10.116.2.56:8443 NODESCOPE_SUPABASE_URL=https://vafiuhbqldcogrmnqbjw.supabase.co NODESCOPE_RUNTIME_DB_HOST=db.vafiuhbqldcogrmnqbjw.supabase.co NODESCOPE_RUNTIME_DB_PASSWORD="$NODESCOPE_RUNTIME_DB_PASSWORD" /tmp/nodescope-server >/tmp/nodescope-capacity-smoke.log 2>&1 &
server_pid="$!"
sleep 2

curl --fail --silent --show-error -H 'Content-Type: application/json' -H "Authorization: Bearer ${token}" --data "{\"schemaVersion\":1,\"codec\":\"protobuf+zstd\",\"agentId\":\"${agent_id}\",\"hostId\":\"${host_id}\",\"bootId\":\"capacity-boot\",\"sequence\":1,\"observedAt\":\"2026-07-22T12:02:00Z\",\"sampleCount\":1,\"metricValueCount\":1,\"uncompressedBytes\":128,\"compressedBytes\":64,\"checksumSha256\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"samples\":[{\"deviceId\":\"cpu-0\",\"metric\":{\"name\":\"cpu.utilization\",\"unit\":\"percent\",\"value\":43.5,\"quality\":\"fresh\",\"source\":\"smoke\",\"semantics\":\"capacity-protected CPU utilization\",\"observedAt\":\"2026-07-22T12:02:00Z\"}}]}" "http://127.0.0.1:${port}/api/v1/ingest" | grep -q '"status":"accepted"'

raw_count="$(migrator_psql --tuples-only --no-align -c "set role nodescope_owner; select count(*) from nodescope.telemetry_batches where host_id = '${host_id}'::uuid;")"
receipt_count="$(migrator_psql --tuples-only --no-align -c "set role nodescope_owner; select count(*) from nodescope.ingest_receipts where host_id = '${host_id}'::uuid and raw_retained = false;")"
latest_value="$(migrator_psql --tuples-only --no-align -c "set role nodescope_owner; select numeric_value from nodescope.metric_latest where host_id = '${host_id}'::uuid and metric_name = 'cpu.utilization';")"
test "$raw_count" = "0"
test "$receipt_count" = "1"
test "$latest_value" = "43.5"
printf 'Capacity protection smoke test passed.\n'
