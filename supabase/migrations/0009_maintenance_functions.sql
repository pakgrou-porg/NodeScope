-- Schema-local maintenance functions invoked by approved pg_cron jobs. The
-- global scheduler is enabled separately; this migration creates only
-- nodescope-owned objects and does not alter any sibling application schema.

set role nodescope_owner;

alter table nodescope.metric_rollups
  add column if not exists p95_value double precision;

create table nodescope.metric_high_water (
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  device_id text not null check (length(trim(device_id)) > 0),
  metric_name text not null check (length(trim(metric_name)) > 0),
  window_seconds integer not null check (window_seconds in (600, 1800, 3600, 7200, 14400)),
  window_ended_at timestamptz not null,
  maximum_value double precision not null,
  source text not null check (length(trim(source)) > 0),
  semantics text not null check (length(trim(semantics)) > 0),
  primary key (host_id, device_id, metric_name, window_seconds)
);

create table nodescope.maintenance_status (
  job_name text primary key check (job_name in ('rollup', 'retention', 'high_water')),
  last_started_at timestamptz,
  last_success_at timestamptz,
  last_row_count bigint not null default 0,
  updated_at timestamptz not null default now()
);

insert into nodescope.maintenance_status (job_name)
values ('rollup'), ('retention'), ('high_water')
on conflict (job_name) do nothing;

alter table nodescope.metric_high_water enable row level security;
alter table nodescope.maintenance_status enable row level security;
create policy metric_high_water_runtime_access on nodescope.metric_high_water for all to nodescope_runtime using (true) with check (true);
create policy maintenance_status_runtime_access on nodescope.maintenance_status for all to nodescope_runtime using (true) with check (true);
grant select, insert, update, delete on nodescope.metric_high_water, nodescope.maintenance_status to nodescope_runtime;

create or replace function nodescope.rollup_recent_samples()
returns bigint
language plpgsql
security definer
set search_path = nodescope, pg_catalog
as $$
declare
  affected_rows bigint := 0;
  resolution_seconds integer;
  retention_interval interval;
  changed_rows bigint;
begin
  update nodescope.maintenance_status
  set last_started_at = now(), updated_at = now()
  where job_name = 'rollup';

  foreach resolution_seconds in array array[60, 300, 600]
  loop
    retention_interval := case resolution_seconds
      when 60 then interval '2 days'
      when 300 then interval '7 days'
      else interval '30 days'
    end;

    insert into nodescope.metric_rollups (
      resolution_seconds, bucket_started_at, host_id, device_id, metric_name,
      sample_count, minimum_value, maximum_value, average_value, last_value,
      p95_value, source, semantics, expires_at
    )
    select
      resolution_seconds,
      date_bin(make_interval(secs => resolution_seconds), sample.observed_at, timestamptz '2000-01-01 00:00:00+00'),
      sample.host_id,
      sample.device_id,
      sample.metric_name,
      count(*)::integer,
      min(sample.numeric_value),
      max(sample.numeric_value),
      avg(sample.numeric_value),
      (array_agg(sample.numeric_value order by sample.observed_at desc))[1],
      percentile_cont(0.95) within group (order by sample.numeric_value),
      (array_agg(sample.source order by sample.observed_at desc))[1],
      (array_agg(sample.semantics order by sample.observed_at desc))[1],
      date_bin(make_interval(secs => resolution_seconds), sample.observed_at, timestamptz '2000-01-01 00:00:00+00') + retention_interval
    from nodescope.metric_samples sample
    where sample.numeric_value is not null
      and sample.quality in ('fresh', 'stale', 'estimated')
      and sample.observed_at >= now() - interval '3 hours'
      and sample.observed_at <= now()
    group by sample.host_id, sample.device_id, sample.metric_name,
      date_bin(make_interval(secs => resolution_seconds), sample.observed_at, timestamptz '2000-01-01 00:00:00+00')
    on conflict (resolution_seconds, bucket_started_at, host_id, device_id, metric_name) do update
    set sample_count = excluded.sample_count,
        minimum_value = excluded.minimum_value,
        maximum_value = excluded.maximum_value,
        average_value = excluded.average_value,
        last_value = excluded.last_value,
        p95_value = excluded.p95_value,
        source = excluded.source,
        semantics = excluded.semantics,
        expires_at = excluded.expires_at;

    get diagnostics changed_rows = row_count;
    affected_rows := affected_rows + changed_rows;
  end loop;

  update nodescope.maintenance_status
  set last_success_at = now(), last_row_count = affected_rows, updated_at = now()
  where job_name = 'rollup';
  return affected_rows;
end;
$$;

create or replace function nodescope.refresh_high_water_marks()
returns bigint
language plpgsql
security definer
set search_path = nodescope, pg_catalog
as $$
declare
  affected_rows bigint := 0;
  window_seconds integer;
  changed_rows bigint;
begin
  update nodescope.maintenance_status
  set last_started_at = now(), updated_at = now()
  where job_name = 'high_water';

  foreach window_seconds in array array[600, 1800, 3600, 7200, 14400]
  loop
    insert into nodescope.metric_high_water (
      host_id, device_id, metric_name, window_seconds, window_ended_at,
      maximum_value, source, semantics
    )
    select
      sample.host_id,
      sample.device_id,
      sample.metric_name,
      window_seconds,
      now(),
      max(sample.numeric_value),
      (array_agg(sample.source order by sample.observed_at desc))[1],
      (array_agg(sample.semantics order by sample.observed_at desc))[1]
    from nodescope.metric_samples sample
    where sample.numeric_value is not null
      and sample.quality in ('fresh', 'stale', 'estimated')
      and sample.observed_at >= now() - make_interval(secs => window_seconds)
      and sample.observed_at <= now()
    group by sample.host_id, sample.device_id, sample.metric_name
    on conflict (host_id, device_id, metric_name, window_seconds) do update
    set window_ended_at = excluded.window_ended_at,
        maximum_value = excluded.maximum_value,
        source = excluded.source,
        semantics = excluded.semantics;

    get diagnostics changed_rows = row_count;
    affected_rows := affected_rows + changed_rows;
  end loop;

  update nodescope.maintenance_status
  set last_success_at = now(), last_row_count = affected_rows, updated_at = now()
  where job_name = 'high_water';
  return affected_rows;
end;
$$;

create or replace function nodescope.prune_expired_telemetry()
returns bigint
language plpgsql
security definer
set search_path = nodescope, pg_catalog
as $$
declare
  affected_rows bigint := 0;
  changed_rows bigint;
begin
  update nodescope.maintenance_status
  set last_started_at = now(), updated_at = now()
  where job_name = 'retention';

  delete from nodescope.telemetry_batches where expires_at <= now();
  get diagnostics changed_rows = row_count;
  affected_rows := affected_rows + changed_rows;

  delete from nodescope.metric_rollups where expires_at <= now();
  get diagnostics changed_rows = row_count;
  affected_rows := affected_rows + changed_rows;

  delete from nodescope.storage_probe_summaries where created_at < now() - interval '30 days';
  get diagnostics changed_rows = row_count;
  affected_rows := affected_rows + changed_rows;

  update nodescope.maintenance_status
  set last_success_at = now(), last_row_count = affected_rows, updated_at = now()
  where job_name = 'retention';
  return affected_rows;
end;
$$;

revoke all on function nodescope.rollup_recent_samples() from public, anon, authenticated, service_role;
revoke all on function nodescope.refresh_high_water_marks() from public, anon, authenticated, service_role;
revoke all on function nodescope.prune_expired_telemetry() from public, anon, authenticated, service_role;
grant execute on function nodescope.rollup_recent_samples(), nodescope.refresh_high_water_marks(), nodescope.prune_expired_telemetry() to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0009_maintenance_functions', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
