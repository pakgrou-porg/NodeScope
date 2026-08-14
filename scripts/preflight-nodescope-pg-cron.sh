#!/usr/bin/env bash
# Read-only shared-Supabase pg_cron readiness check. It never enables an
# extension, creates a schedule, mutates NodeScope state, or examines sibling data.
set -euo pipefail

require_scheduled=false
if [[ "${1:-}" == "--require-scheduled" ]]; then
  require_scheduled=true
elif [[ $# -ne 0 ]]; then
  echo "usage: $0 [--require-scheduled]" >&2
  exit 2
fi

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

: "${NODESCOPE_SUPABASE_DB_URL:?NODESCOPE_SUPABASE_DB_URL is required}"
: "${NODESCOPE_MIGRATOR_DB_PASSWORD:?NODESCOPE_MIGRATOR_DB_PASSWORD is required}"
command -v psql >/dev/null 2>&1 || { echo 'required dependency is unavailable: psql' >&2; exit 2; }

host="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -E 's#^[a-z]+://[^@]*@([^:/?]+).*#\1#')"
port="$(printf '%s' "$NODESCOPE_SUPABASE_DB_URL" | sed -nE 's#^[a-z]+://[^@]*@[^:/?]+:([0-9]+).*#\1#p')"
port="${port:-5432}"
migrator_psql=(env PGCONNECT_TIMEOUT=10 PGHOST="$host" PGPORT="$port" PGDATABASE=postgres PGUSER=nodescope_migrate_login PGPASSWORD="$NODESCOPE_MIGRATOR_DB_PASSWORD" PGSSLMODE=require psql --no-psqlrc -q -v ON_ERROR_STOP=1)

read_only() {
  "${migrator_psql[@]}" -At -c "begin read only; set role nodescope_owner; $1; rollback;"
}

extension_version="$(read_only "select extversion from pg_extension where extname = 'pg_cron'")"
[[ -n "$extension_version" ]] || { echo 'pg_cron extension is not enabled' >&2; exit 1; }

for routine in nodescope.rollup_recent_samples nodescope.refresh_high_water_marks nodescope.prune_expired_telemetry; do
  read_only "select to_regprocedure('${routine}()') is not null" | grep -qx 't' || { echo "required NodeScope routine is unavailable: $routine" >&2; exit 1; }
done

status_rows="$(read_only "select job_name || '|' || case when last_success_at is null then 'never' else 'recorded' end from nodescope.maintenance_status order by job_name")"
[[ "$(printf '%s\n' "$status_rows" | grep -c '^[a-z_]*|')" -eq 3 ]] || { echo 'NodeScope maintenance status rows are incomplete' >&2; exit 1; }

job_rows="$(read_only "select jobname || '|' || schedule from cron.job where jobname in ('nodescope-rollup-minute', 'nodescope-high-water-minute', 'nodescope-retention-daily') order by jobname")"
job_count="$(printf '%s\n' "$job_rows" | sed '/^$/d' | wc -l | tr -d ' ')"
if [[ "$job_count" -ne 0 && "$job_count" -ne 3 ]]; then
  echo "NodeScope pg_cron job set is incomplete: expected zero or three jobs, found $job_count" >&2
  exit 1
fi
if [[ "$require_scheduled" == true && "$job_count" -ne 3 ]]; then
  echo 'NodeScope pg_cron jobs are not all scheduled' >&2
  exit 1
fi

failed_runs="$(read_only "select count(*) from cron.job_run_details details join cron.job job on job.jobid = details.jobid where job.jobname in ('nodescope-rollup-minute', 'nodescope-high-water-minute', 'nodescope-retention-daily') and details.status not in ('succeeded', 'running') and details.start_time > now() - interval '5 days'")"
[[ "$failed_runs" =~ ^[0-9]+$ ]] || { echo 'could not determine NodeScope pg_cron run failures' >&2; exit 1; }

printf 'pg_cron preflight passed: version=%s scheduled_jobs=%s failed_runs_last_5d=%s\n' "$extension_version" "$job_count" "$failed_runs"
