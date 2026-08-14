# NodeScope pg_cron Activation

This runbook enables the three **NodeScope-owned** database maintenance schedules only after the protected shared-Supabase project has passed the required isolation controls and an administrator has explicitly authorized the change. The package intentionally does not schedule a replica-local timer: maintenance remains database-side so two independent server replicas cannot duplicate it.

> The read-only preflight **does not enable `pg_cron`** and **does not create, modify, or remove any job**. It uses the dedicated migrator login only to inspect NodeScope maintenance routines, NodeScope-namespaced cron jobs, and recent status metadata. It never reads a sibling schema.

## Scheduled work

| Job | Cadence | Database command | Purpose |
| --- | --- | --- | --- |
| `nodescope-rollup-minute` | Every minute | `nodescope.rollup_recent_samples()` | Materializes 1-, 5-, and 10-minute summaries. |
| `nodescope-high-water-minute` | Every minute, separately named | `nodescope.refresh_high_water_marks()` | Refreshes 10/30/60/120/240-minute high-water evidence. |
| `nodescope-retention-daily` | 03:10 UTC daily | `nodescope.prune_expired_telemetry()` | Removes expired raw telemetry, rollups, and probe summaries. |

The schedule SQL removes and recreates **only these exact three names**, validates `pg_cron`, validates each NodeScope routine, and requires exactly three configured NodeScope jobs after scheduling. Supabase documents that pg_cron jobs and execution history reside in the `cron` schema and recommends keeping jobs short with limited concurrency.[1] NodeScope therefore uses three small, staggered schedules rather than a broad replica-local recurring worker.

## Read-only preflight

The shared-project administrator sources only the existing migrator secret through protected storage, then runs:

```bash
set -a
. /etc/nodescope/credentials/migrator.env
set +a

./scripts/preflight-nodescope-pg-cron.sh
```

The expected pre-activation result is `scheduled_jobs=0` and `failed_runs_last_5d=0`. It validates the extension, all maintenance functions, all three maintenance status rows, job-set consistency, and the five-day failure count. If a partial job set exists, it fails closed. After activation, rerun the same command with `--require-scheduled`; it must report `scheduled_jobs=3`.

## Confirmation-required activation

The following is a **state-changing shared-database operation**. Do not perform it until the owner has approved it, the disposable shared-Supabase fixture gate has passed for the current tracked source, and the read-only preflight has passed. The administrator must use a protected `NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL`; do not use a server runtime, native agent, verifier, storage auditor, browser environment variable, Portainer variable, or copied connection string.

```bash
psql "$NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL" \
  --no-psqlrc -v ON_ERROR_STOP=1 \
  -f supabase/operations/schedule_maintenance.sql

./scripts/preflight-nodescope-pg-cron.sh --require-scheduled
```

Capture only the redacted schedule names, cron expressions, extension version, status count, and failed-run count. Do not capture database URLs, passwords, SQL-editor session data, raw logs, or unrelated `cron` jobs. Job-run failures are visible through `cron.job_run_details`; Supabase’s troubleshooting guide recommends examining non-succeeded recent runs and verifying that the pg_cron scheduler is active.[2]

## Failure and recovery

If the preflight or post-activation check fails, stop. Do not broaden `nodescope_migrate_login`, grant runtime replicas cron access, create a replica-local timer, or manipulate non-NodeScope cron records. An authorized shared-project administrator may remove only the exact three NodeScope job names using the tracked schedule file’s explicit `cron.unschedule` set, then rerun the read-only preflight. Preserve redacted evidence and resolve the extension, privilege, scheduler, or maintenance-function failure before another activation attempt.

## References

[1] [Supabase Cron documentation](https://supabase.com/docs/guides/cron)

[2] [Supabase pg_cron debugging guide](https://supabase.com/docs/guides/troubleshooting/pgcron-debugging-guide-n1KTaz)
