-- Resolve PL/pgSQL variable and table-column ambiguity in maintenance functions.

set role nodescope_owner;

create or replace function nodescope.rollup_recent_samples()
returns bigint
language plpgsql
security definer
set search_path = nodescope, pg_catalog
as $$
declare
  affected_rows bigint := 0;
  bucket_resolution integer;
  bucket_retention interval;
  changed_rows bigint;
begin
  update nodescope.maintenance_status
  set last_started_at = now(), updated_at = now()
  where job_name = 'rollup';

  foreach bucket_resolution in array array[60, 300, 600]
  loop
    bucket_retention := case bucket_resolution
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
      bucket_resolution,
      date_bin(make_interval(secs => bucket_resolution), sample.observed_at, timestamptz '2000-01-01 00:00:00+00'),
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
      date_bin(make_interval(secs => bucket_resolution), sample.observed_at, timestamptz '2000-01-01 00:00:00+00') + bucket_retention
    from nodescope.metric_samples sample
    where sample.numeric_value is not null
      and sample.quality in ('fresh', 'stale', 'estimated')
      and sample.observed_at >= now() - interval '3 hours'
      and sample.observed_at <= now()
    group by sample.host_id, sample.device_id, sample.metric_name,
      date_bin(make_interval(secs => bucket_resolution), sample.observed_at, timestamptz '2000-01-01 00:00:00+00')
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
  high_water_window integer;
  changed_rows bigint;
begin
  update nodescope.maintenance_status
  set last_started_at = now(), updated_at = now()
  where job_name = 'high_water';

  foreach high_water_window in array array[600, 1800, 3600, 7200, 14400]
  loop
    insert into nodescope.metric_high_water (
      host_id, device_id, metric_name, window_seconds, window_ended_at,
      maximum_value, source, semantics
    )
    select
      sample.host_id,
      sample.device_id,
      sample.metric_name,
      high_water_window,
      now(),
      max(sample.numeric_value),
      (array_agg(sample.source order by sample.observed_at desc))[1],
      (array_agg(sample.semantics order by sample.observed_at desc))[1]
    from nodescope.metric_samples sample
    where sample.numeric_value is not null
      and sample.quality in ('fresh', 'stale', 'estimated')
      and sample.observed_at >= now() - make_interval(secs => high_water_window)
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

insert into nodescope.schema_migrations (version, source_checksum)
values ('0010_fix_maintenance_function_variables', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
