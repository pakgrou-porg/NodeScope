-- Read-only, schema-local fleet summary for the standalone terminal CLI and TUI.
-- This function deliberately exposes operational status only; no prompt, response,
-- raw request metadata, credential material, or sibling-schema data is returned.

set role nodescope_owner;

create or replace function nodescope.fleet_ingestion_status()
returns table(
  host_slug text,
  display_name text,
  platform text,
  effective_interval_seconds integer,
  freshness_state text,
  latest_receipt timestamptz,
  current_metric_count bigint,
  unavailable_metric_count bigint,
  stale_metric_count bigint,
  clock_offset_seconds double precision,
  clock_offset_quality text
)
language sql
security definer
set search_path = nodescope, pg_temp
as $$
  select
    h.slug,
    h.display_name,
    h.platform,
    coalesce(o.interval_seconds, s.global_interval_seconds) as effective_interval_seconds,
    case
      when max(m.received_at) is null then 'unavailable'
      when now() - max(m.received_at) > make_interval(secs => coalesce(o.interval_seconds, s.global_interval_seconds) * 2) then 'stale'
      else 'fresh'
    end as freshness_state,
    max(m.received_at) as latest_receipt,
    count(m.metric_name) as current_metric_count,
    count(m.metric_name) filter (where m.quality = 'unavailable') as unavailable_metric_count,
    count(m.metric_name) filter (where m.quality = 'stale') as stale_metric_count,
    max(m.numeric_value) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_offset_seconds,
    max(m.quality) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_offset_quality
  from nodescope.hosts h
  cross join nodescope.collection_settings s
  left join nodescope.host_collection_overrides o on o.host_id = h.id
  left join nodescope.metric_latest m on m.host_id = h.id
  where s.singleton = true
  group by h.slug, h.display_name, h.platform, o.interval_seconds, s.global_interval_seconds
  order by h.display_name;
$$;

revoke all on function nodescope.fleet_ingestion_status() from public;
grant execute on function nodescope.fleet_ingestion_status() to nodescope_verifier;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0015_terminal_fleet_status', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
