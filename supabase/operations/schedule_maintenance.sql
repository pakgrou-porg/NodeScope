-- Project-level pg_cron definitions. Apply only after explicit shared-project
-- administrator authorization and a successful read-only preflight. This file
-- touches only cron jobs whose names are explicitly NodeScope-owned.
\set ON_ERROR_STOP on

do $$
begin
  if not exists (select 1 from pg_extension where extname = 'pg_cron') then
    raise exception 'pg_cron extension is not enabled';
  end if;
  if to_regclass('cron.job') is null then
    raise exception 'cron.job is unavailable';
  end if;
  if to_regprocedure('nodescope.rollup_recent_samples()') is null
    or to_regprocedure('nodescope.refresh_high_water_marks()') is null
    or to_regprocedure('nodescope.prune_expired_telemetry()') is null then
    raise exception 'required NodeScope maintenance function is unavailable';
  end if;
end
$$;

select cron.unschedule(jobid)
from cron.job
where jobname in ('nodescope-rollup-minute', 'nodescope-high-water-minute', 'nodescope-retention-daily');

select cron.schedule(
  'nodescope-rollup-minute',
  '* * * * *',
  'set role nodescope_owner; select nodescope.rollup_recent_samples();'
);

select cron.schedule(
  'nodescope-high-water-minute',
  '* * * * *',
  'set role nodescope_owner; select nodescope.refresh_high_water_marks();'
);

select cron.schedule(
  'nodescope-retention-daily',
  '10 3 * * *',
  'set role nodescope_owner; select nodescope.prune_expired_telemetry();'
);

do $$
declare
  configured_jobs integer;
begin
  select count(*) into configured_jobs
  from cron.job
  where jobname in ('nodescope-rollup-minute', 'nodescope-high-water-minute', 'nodescope-retention-daily');
  if configured_jobs <> 3 then
    raise exception 'expected exactly three NodeScope pg_cron jobs, found %', configured_jobs;
  end if;
end
$$;
