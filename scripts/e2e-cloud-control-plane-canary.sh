#!/usr/bin/env bash
# Disposable cloud control-plane canary. It proves authenticated ingestion,
# TLS 1.3 mTLS, idempotency, and persisted evidence quality using ephemeral
# NodeScope rows and certificates. It does not qualify Framework hardware.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required}"
: "${NODESCOPE_RUNTIME_DB_PASSWORD:?NODESCOPE_RUNTIME_DB_PASSWORD is required}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"
for dependency in curl git go psql sha256sum; do
  if ! command -v "$dependency" >/dev/null 2>&1; then
    echo "required dependency is unavailable: $dependency" >&2
    exit 2
  fi
done

host="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -E 's#^[a-z]+://[^@]*@([^:/?]+).*#\1#')"
port="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -nE 's#^[a-z]+://[^@]*@[^:/?]+:([0-9]+).*#\1#p')"
port="${port:-5432}"
canary_port="18443"
host_id="$(cat /proc/sys/kernel/random/uuid)"
agent_id="$(cat /proc/sys/kernel/random/uuid)"
token="nodescope-cloud-canary-$(cat /proc/sys/kernel/random/uuid)"
digest="$(printf '%s' "$token" | sha256sum | awk '{print $1}')"
temporary_directory="$(mktemp -d)"
server_log="$temporary_directory/server.log"
server_pid=""
phase="initialization"

migrator_psql() {
  PGCONNECT_TIMEOUT=10 PGHOST="$host" PGPORT="$port" PGDATABASE=postgres \
    PGUSER=nodescope_migrate_login PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" \
    PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1 "$@"
}

cleanup() {
  local status="$1"
  if [[ "$status" -ne 0 ]]; then
    printf 'Cloud control-plane canary failed at phase=%s.\n' "$phase" >&2
    if [[ -s "$server_log" ]]; then
      tail -n 20 "$server_log" >&2 || true
    fi
  fi
  if [[ -n "$server_pid" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  migrator_psql -c "begin; set role nodescope_owner; delete from nodescope.hosts where id = '${host_id}'::uuid; commit;" >/dev/null 2>&1 || true
  rm -rf "$temporary_directory"
}
on_exit() {
  local status=$?
  trap - EXIT
  cleanup "$status"
  exit "$status"
}
trap on_exit EXIT

phase="create ephemeral database identities"
migrator_psql -c "begin; set role nodescope_owner; insert into nodescope.hosts(id, slug, display_name, platform) values ('${host_id}'::uuid, 'cloud-canary-${host_id:0:8}', 'Cloud Control-Plane Canary', 'cloud/linux'); insert into nodescope.agents(id, host_id, display_name, credential_digest, credential_hint) values ('${agent_id}'::uuid, '${host_id}'::uuid, 'Cloud Canary Agent', decode('${digest}', 'hex'), 'cloud'); commit;"

phase="build disposable server and internal PKI tools"
go build -o "$temporary_directory/nodescope-pki" ./cmd/nodescope-pki
go build -o "$temporary_directory/nodescope-server" ./cmd/nodescope-server
"$temporary_directory/nodescope-pki" init-root --common-name 'NodeScope Cloud Canary Root' --output-directory "$temporary_directory/pki" --years 1 >/dev/null
"$temporary_directory/nodescope-pki" issue --kind replica --common-name cloud-canary --ca-certificate "$temporary_directory/pki/root-ca.pem" --ca-key "$temporary_directory/pki/root-ca-key.pem" --certificate-output "$temporary_directory/server.crt" --key-output "$temporary_directory/server.key" --ip-san 127.0.0.1 --days 7 >/dev/null
"$temporary_directory/nodescope-pki" issue --kind agent --common-name cloud-canary-agent --ca-certificate "$temporary_directory/pki/root-ca.pem" --ca-key "$temporary_directory/pki/root-ca-key.pem" --certificate-output "$temporary_directory/agent.crt" --key-output "$temporary_directory/agent.key" --days 7 >/dev/null

phase="start TLS 1.3 mTLS server"
NODESCOPE_ALLOW_JSON_INGEST=true \
NODESCOPE_REPLICA_ID=cloud-canary \
NODESCOPE_REPLICA_ROLE=preferred \
NODESCOPE_LISTEN_ADDRESS="127.0.0.1:${canary_port}" \
NODESCOPE_PRIMARY_ENDPOINT="https://127.0.0.1:${canary_port}" \
NODESCOPE_SECONDARY_ENDPOINT="https://127.0.0.1:$((canary_port + 1))" \
NODESCOPE_SUPABASE_URL="https://vafiuhbqldcogrmnqbjw.supabase.co" \
NODESCOPE_RUNTIME_DB_HOST="$host" \
NODESCOPE_RUNTIME_DB_PORT="$port" \
NODESCOPE_RUNTIME_DB_PASSWORD="$NODESCOPE_RUNTIME_DB_PASSWORD" \
NODESCOPE_TLS_CERT_PATH="$temporary_directory/server.crt" \
NODESCOPE_TLS_KEY_PATH="$temporary_directory/server.key" \
NODESCOPE_AGENT_CLIENT_CA_CERT_PATH="$temporary_directory/pki/root-ca.pem" \
NODESCOPE_REQUIRE_AGENT_MTLS=true \
"$temporary_directory/nodescope-server" >"$server_log" 2>&1 &
server_pid="$!"

phase="wait for authenticated TLS readiness"
for attempt in $(seq 1 20); do
  if curl --silent --show-error --fail --tlsv1.3 --cacert "$temporary_directory/pki/root-ca.pem" --cert "$temporary_directory/agent.crt" --key "$temporary_directory/agent.key" "https://127.0.0.1:${canary_port}/readyz" | grep -q '"status":"ready"'; then
    break
  fi
  if [[ "$attempt" == 20 ]]; then
    echo "cloud canary server did not become ready" >&2
    exit 1
  fi
  sleep 1
done

phase="reject missing client certificate"
if curl --silent --show-error --fail --tlsv1.3 --cacert "$temporary_directory/pki/root-ca.pem" "https://127.0.0.1:${canary_port}/healthz" >/dev/null 2>&1; then
  echo "UNSAFE: cloud canary accepted a request without an agent mTLS certificate" >&2
  exit 1
fi

phase="verify bearer preflight authentication"
preflight="$(curl --silent --show-error --fail --tlsv1.3 --cacert "$temporary_directory/pki/root-ca.pem" --cert "$temporary_directory/agent.crt" --key "$temporary_directory/agent.key" -H "Authorization: Bearer ${token}" "https://127.0.0.1:${canary_port}/api/v1/ingest/preflight")"
grep -q '"status":"authenticated"' <<<"$preflight"
invalid_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' --tlsv1.3 --cacert "$temporary_directory/pki/root-ca.pem" --cert "$temporary_directory/agent.crt" --key "$temporary_directory/agent.key" -H 'Authorization: Bearer invalid' "https://127.0.0.1:${canary_port}/api/v1/ingest/preflight")"
test "$invalid_status" = 401

phase="submit and retry fresh telemetry"
observed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
payload="${temporary_directory}/payload.json"
cat >"$payload" <<EOF
{"schemaVersion":1,"codec":"protobuf+zstd","agentId":"${agent_id}","hostId":"${host_id}","bootId":"cloud-canary-boot","sequence":1,"observedAt":"${observed_at}","sampleCount":1,"metricValueCount":1,"uncompressedBytes":128,"compressedBytes":64,"checksumSha256":"0000000000000000000000000000000000000000000000000000000000000000","samples":[{"deviceId":"cpu-0","metric":{"name":"cpu.utilization","unit":"percent","value":42.5,"quality":"fresh","source":"cloud-canary","semantics":"authenticated cloud control-plane evidence","observedAt":"${observed_at}"}}]}
EOF

first_response="$(curl --silent --show-error --fail --tlsv1.3 --cacert "$temporary_directory/pki/root-ca.pem" --cert "$temporary_directory/agent.crt" --key "$temporary_directory/agent.key" -H 'Content-Type: application/json' -H "Authorization: Bearer ${token}" --data @"$payload" "https://127.0.0.1:${canary_port}/api/v1/ingest")"
grep -q '"status":"accepted"' <<<"$first_response"
second_response="$(curl --silent --show-error --fail --tlsv1.3 --cacert "$temporary_directory/pki/root-ca.pem" --cert "$temporary_directory/agent.crt" --key "$temporary_directory/agent.key" -H 'Content-Type: application/json' -H "Authorization: Bearer ${token}" --data @"$payload" "https://127.0.0.1:${canary_port}/api/v1/ingest")"
grep -q '"status":"duplicate"' <<<"$second_response"

phase="verify persisted receipt-time evidence"
evidence="$(migrator_psql -At -c "begin read only; set role nodescope_owner; select quality || '|' || source || '|' || semantics || '|' || case when received_at >= observed_at then 'receipt_after_observation' else 'invalid_receipt_order' end from nodescope.metric_latest where host_id = '${host_id}'::uuid and device_id = 'cpu-0' and metric_name = 'cpu.utilization'; select (select count(*) from nodescope.telemetry_batches where host_id = '${host_id}'::uuid) + (select count(*) from nodescope.ingest_receipts where host_id = '${host_id}'::uuid); rollback;")"
if ! grep -qx 'fresh|cloud-canary|authenticated cloud control-plane evidence|receipt_after_observation' <<<"$evidence"; then
  printf 'Unexpected redacted persisted metric evidence: %s\n' "$(head -n 1 <<<"$evidence")" >&2
  exit 1
fi
if ! grep -qx '1' <<<"$evidence"; then
  printf 'Unexpected persisted idempotency receipt count: %s\n' "$(tail -n 1 <<<"$evidence")" >&2
  exit 1
fi

printf 'Cloud control-plane canary passed: TLS 1.3 mTLS, bearer authentication, authenticated preflight, rejected invalid credential, idempotent duplicate delivery, and fresh persisted receipt-time evidence.\n'
