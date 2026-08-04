-- Least-privilege agent enrollment, credential rotation, verification, and storage auditing.
-- No function reads from or writes to any non-NodeScope schema.

set role nodescope_owner;

alter table nodescope.agents
  add column if not exists credential_rotated_at timestamptz not null default now(),
  add column if not exists credential_expires_at timestamptz,
  add column if not exists rotation_version integer not null default 1 check (rotation_version >= 1),
  add column if not exists last_credential_used_at timestamptz;

create index if not exists agents_credential_expiry_idx
  on nodescope.agents (credential_expires_at)
  where credential_expires_at is not null and revoked_at is null;

-- Login and group-role creation is deliberately performed by the privileged
-- bootstrap script, not by this migration. The dedicated NodeScope migrator has
-- no CREATEROLE privilege and must not be broadened merely to apply schema DDL.
grant usage on schema nodescope to nodescope_enroller, nodescope_verifier, nodescope_storage_auditor;

create or replace function nodescope.enroll_or_rotate_agent(
  p_slug text,
  p_display_name text,
  p_platform text,
  p_address inet,
  p_credential_digest bytea,
  p_credential_hint text,
  p_credential_expires_at timestamptz default now() + interval '90 days'
) returns table(host_id uuid, agent_id uuid, rotation_version integer)
language plpgsql
security definer
set search_path = nodescope, pg_temp
as $$
declare
  v_host_id uuid;
  v_agent_id uuid;
  v_rotation_version integer;
begin
  if p_slug !~ '^[a-z0-9][a-z0-9-]{0,62}$' then
    raise exception 'invalid host slug';
  end if;
  if length(trim(p_display_name)) = 0 or length(trim(p_platform)) = 0 then
    raise exception 'display name and platform are required';
  end if;
  if octet_length(p_credential_digest) <> 32 then
    raise exception 'agent credential digest must be SHA-256 length';
  end if;
  if length(trim(p_credential_hint)) < 4 or length(trim(p_credential_hint)) > 32 then
    raise exception 'credential hint must be 4 to 32 characters';
  end if;
  if p_credential_expires_at <= now() then
    raise exception 'credential expiry must be in the future';
  end if;

  insert into nodescope.hosts (slug, display_name, platform, address)
  values (p_slug, p_display_name, p_platform, p_address)
  on conflict (slug) do update
    set display_name = excluded.display_name,
        platform = excluded.platform,
        address = excluded.address,
        updated_at = now()
  returning id into v_host_id;

  insert into nodescope.agents (
    host_id, display_name, credential_digest, credential_hint, credential_rotated_at,
    credential_expires_at, rotation_version, revoked_at
  ) values (
    v_host_id, p_display_name || ' native agent', p_credential_digest, p_credential_hint, now(),
    p_credential_expires_at, 1, null
  )
  on conflict (host_id) do update
    set credential_digest = excluded.credential_digest,
        credential_hint = excluded.credential_hint,
        credential_rotated_at = now(),
        credential_expires_at = excluded.credential_expires_at,
        rotation_version = nodescope.agents.rotation_version + 1,
        revoked_at = null,
        updated_at = now()
  returning id, rotation_version into v_agent_id, v_rotation_version;

  insert into nodescope.audit_events (
    actor_type, actor_id, action, target_type, target_id, outcome, metadata
  ) values (
    'system', session_user, 'rotate_agent_credential', 'agent', v_agent_id::text, 'completed',
    jsonb_build_object('host_id', v_host_id, 'rotation_version', v_rotation_version, 'credential_hint', p_credential_hint)
  );

  return query select v_host_id, v_agent_id, v_rotation_version;
end;
$$;

revoke all on function nodescope.enroll_or_rotate_agent(text, text, text, inet, bytea, text, timestamptz) from public;
grant execute on function nodescope.enroll_or_rotate_agent(text, text, text, inet, bytea, text, timestamptz) to nodescope_enroller;

create or replace function nodescope.host_ingestion_status(p_slug text)
returns table(
  host_slug text,
  latest_receipt timestamptz,
  current_metric_count bigint,
  unavailable_metric_count bigint,
  stale_metric_count bigint,
  clock_offset_seconds double precision,
  clock_offset_quality text,
  clock_offset_source text,
  clock_observed_at timestamptz,
  clock_received_at timestamptz
)
language sql
security definer
set search_path = nodescope, pg_temp
as $$
  select h.slug,
         max(m.received_at) as latest_receipt,
         count(m.metric_name) as current_metric_count,
         count(m.metric_name) filter (where m.quality = 'unavailable') as unavailable_metric_count,
         count(m.metric_name) filter (where m.quality = 'stale') as stale_metric_count,
         max(m.numeric_value) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_offset_seconds,
         max(m.quality) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_offset_quality,
         max(m.source) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_offset_source,
         max(m.observed_at) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_observed_at,
         max(m.received_at) filter (where m.device_id = 'agent-clock' and m.metric_name = 'agent.clock_offset_seconds') as clock_received_at
  from nodescope.hosts h
  left join nodescope.metric_latest m on m.host_id = h.id
  where h.slug = p_slug
  group by h.slug;
$$;

revoke all on function nodescope.host_ingestion_status(text) from public;
grant execute on function nodescope.host_ingestion_status(text) to nodescope_verifier;

create or replace function nodescope.storage_probe_evidence(p_slug text, p_since timestamptz)
returns table(
  host_slug text,
  first_received_at timestamptz,
  last_received_at timestamptz,
  received_batch_count bigint,
  expected_batch_count bigint,
  completeness_percent numeric,
  max_gap_seconds numeric,
  median_compressed_bytes numeric,
  p95_compressed_bytes numeric,
  p99_compressed_bytes numeric,
  total_compressed_bytes bigint,
  metric_cardinality bigint,
  telemetry_relation_bytes bigint,
  telemetry_index_bytes bigint,
  raw_sample_relation_bytes bigint,
  raw_sample_index_bytes bigint
)
language sql
security definer
set search_path = nodescope, pg_temp
as $$
  with host_row as (
    select id, slug from nodescope.hosts where slug = p_slug
  ), batches as (
    select b.*, lag(b.received_at) over (order by b.received_at) as previous_received_at
    from nodescope.telemetry_batches b
    join host_row h on h.id = b.host_id
    where b.received_at >= p_since
  ), aggregate as (
    select min(received_at) as first_received_at,
           max(received_at) as last_received_at,
           count(*) as received_batch_count,
           max(extract(epoch from received_at - previous_received_at)) as max_gap_seconds,
           percentile_cont(0.5) within group (order by compressed_bytes) as median_compressed_bytes,
           percentile_cont(0.95) within group (order by compressed_bytes) as p95_compressed_bytes,
           percentile_cont(0.99) within group (order by compressed_bytes) as p99_compressed_bytes,
           coalesce(sum(compressed_bytes), 0) as total_compressed_bytes,
           coalesce(sum(metric_value_count), 0) as metric_value_count
    from batches
  ), cardinality as (
    select count(distinct ms.device_id || E'\x1f' || ms.metric_name) as metric_cardinality
    from nodescope.metric_samples ms
    join host_row h on h.id = ms.host_id
    where ms.observed_at >= p_since
  ), configured as (
    select coalesce(o.interval_seconds, s.global_interval_seconds) as interval_seconds
    from nodescope.collection_settings s
    left join host_row h on true
    left join nodescope.host_collection_overrides o on o.host_id = h.id
    where s.singleton = true
  )
  select h.slug,
         a.first_received_at,
         a.last_received_at,
         a.received_batch_count,
         case when a.first_received_at is null then 0 else greatest(1, floor(extract(epoch from a.last_received_at - a.first_received_at) / c.interval_seconds)::bigint + 1) end,
         case when a.first_received_at is null then 0 else round((a.received_batch_count::numeric / greatest(1, floor(extract(epoch from a.last_received_at - a.first_received_at) / c.interval_seconds)::bigint + 1)) * 100, 2) end,
         coalesce(a.max_gap_seconds, 0),
         a.median_compressed_bytes,
         a.p95_compressed_bytes,
         a.p99_compressed_bytes,
         a.total_compressed_bytes,
         coalesce(cardinality.metric_cardinality, 0),
         pg_table_size('nodescope.telemetry_batches'::regclass),
         pg_indexes_size('nodescope.telemetry_batches'::regclass),
         pg_table_size('nodescope.metric_samples'::regclass),
         pg_indexes_size('nodescope.metric_samples'::regclass)
  from host_row h
  cross join aggregate a
  cross join cardinality
  cross join configured c;
$$;

revoke all on function nodescope.storage_probe_evidence(text, timestamptz) from public;
grant execute on function nodescope.storage_probe_evidence(text, timestamptz) to nodescope_storage_auditor;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0014_least_privilege_agent_operations', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
