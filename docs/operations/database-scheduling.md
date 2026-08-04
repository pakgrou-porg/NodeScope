# NodeScope database-side maintenance scheduling

NodeScope retention and rollup jobs run inside the shared Supabase database through `pg_cron`, not through a replica-local timer. This preserves idempotent maintenance when Framework or Asus is unavailable and avoids duplicate work from two independently running replicas.

> Supabase Cron runs SQL snippets or database functions inside Postgres; jobs and job-run records live in the `cron` schema.[1]

NodeScope uses a small set of schema-local functions, each guarded by the `nodescope` schema boundary and transaction semantics. Jobs are intentionally staggered and short: the minute rollup is lightweight, raw retention cleanup is separate, and summary retention cleanup runs daily. Supabase recommends keeping jobs below ten minutes and avoiding excessive concurrent schedules; NodeScope therefore limits its own recurring maintenance jobs to well below that ceiling.[1]

The `nodescope_migrate_login` is the only role allowed to create or modify NodeScope schedule definitions. Runtime replicas neither create schedules nor use `setInterval`/`node-cron`. The browser console reads maintenance freshness and failures from NodeScope state and from the controlled diagnostic queries; it does not expose the shared `cron` schema directly.

## Operational checks

Administrators verify the scheduler worker and inspect failed runs through `cron.job_run_details`, as documented by Supabase.[2] A failed job must create or refresh a NodeScope in-console alert through the server’s normal control path. Before adding a new database-side job, run the dedicated migrator rollback preflight and sibling-schema noninterference gate.

## References

[1] [Supabase Cron documentation](https://supabase.com/docs/guides/cron)

[2] [Supabase pg_cron debugging guide](https://supabase.com/docs/guides/troubleshooting/pgcron-debugging-guide-n1KTaz)
