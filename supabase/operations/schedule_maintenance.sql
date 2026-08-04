-- Project-level pg_cron definitions, approved for the shared Supabase project.
-- This file touches only cron jobs whose names are explicitly NodeScope-owned.

select cron.unschedule(jobid)
from cron.job
where jobname in ('nodescope-rollup-minute', 'nodescope-high-water-minute', 'nodescope-retention-daily');

select cron.schedule(
  'nodescope-rollup-minute',
  '* * * * *',
  'select nodescope.rollup_recent_samples();'
);

select cron.schedule(
  'nodescope-high-water-minute',
  '* * * * *',
  'select nodescope.refresh_high_water_marks();'
);

select cron.schedule(
  'nodescope-retention-daily',
  '10 3 * * *',
  'select nodescope.prune_expired_telemetry();'
);
