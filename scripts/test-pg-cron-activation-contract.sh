#!/usr/bin/env bash
# Ensure the pg_cron package stays NodeScope-namespaced, non-secret, and
# confirmation-gated before any protected Supabase scheduling change.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
schedule_sql="$repository_root/supabase/operations/schedule_maintenance.sql"
preflight="$repository_root/scripts/preflight-nodescope-pg-cron.sh"
guide="$repository_root/docs/operations/pg-cron-activation.md"

for required in \
  "extname = 'pg_cron'" \
  "to_regclass('cron.job')" \
  "nodescope-rollup-minute" \
  "nodescope-high-water-minute" \
  "nodescope-retention-daily" \
  "set role nodescope_owner; select nodescope.rollup_recent_samples();" \
  "set role nodescope_owner; select nodescope.refresh_high_water_marks();" \
  "set role nodescope_owner; select nodescope.prune_expired_telemetry();" \
  'expected exactly three NodeScope pg_cron jobs'; do
  grep -Fq -- "$required" "$schedule_sql" || { echo "pg_cron schedule package missing: $required" >&2; exit 1; }
done

if grep -Eq "cron\.(schedule|unschedule).*[^a-z]nodescope" "$schedule_sql"; then
  echo 'pg_cron schedule package must not name a non-NodeScope job' >&2
  exit 1
fi

for required in \
  '--require-scheduled' \
  'begin read only; set role nodescope_owner;' \
  'NODESCOPE_MIGRATOR_DB_PASSWORD' \
  'cron.job_run_details' \
  'NodeScope pg_cron job set is incomplete' \
  'pg_cron preflight passed:'; do
  grep -Fq -- "$required" "$preflight" || { echo "pg_cron preflight missing: $required" >&2; exit 1; }
done

for required in \
  'Confirmation-required activation' \
  'does not enable `pg_cron`' \
  'does not create, modify, or remove any job' \
  'NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL' \
  'preflight-nodescope-pg-cron.sh' \
  'schedule_maintenance.sql' \
  'cron.job_run_details'; do
  grep -Fq -- "$required" "$guide" || { echo "pg_cron activation guide missing: $required" >&2; exit 1; }
done

echo 'NodeScope pg_cron activation contract passed.'
