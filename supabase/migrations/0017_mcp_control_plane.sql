-- NodeScope MCP control-plane persistence. All changes remain schema-local and
-- record a server-side audit event. No tool argument can carry inference content.
set role nodescope_owner;

create table if not exists nodescope.alert_rules (
  id text primary key check (length(id) <= 120),
  metric text not null check (length(metric) <= 160),
  comparison text not null check (comparison in ('gt', 'lt')),
  threshold numeric not null,
  duration_seconds integer not null check (duration_seconds between 1 and 86400),
  severity text not null check (severity in ('info', 'warning', 'critical')),
  enabled boolean not null default true,
  scope text not null check (scope in ('fleet', 'host')),
  host_id uuid references nodescope.hosts(id) on delete cascade,
  updated_by text,
  updated_at timestamptz not null default now(),
  check ((scope = 'fleet' and host_id is null) or (scope = 'host' and host_id is not null))
);

create table if not exists nodescope.alerts (
  id text primary key check (length(id) <= 120),
  host_id uuid references nodescope.hosts(id) on delete set null,
  rule_id text references nodescope.alert_rules(id) on delete set null,
  title text not null check (length(title) <= 240),
  severity text not null check (severity in ('info', 'warning', 'critical')),
  state text not null check (state in ('active', 'acknowledged', 'resolved')) default 'active',
  observed_at timestamptz not null default now(),
  acknowledged_at timestamptz,
  acknowledged_by text,
  acknowledgement_note text check (length(acknowledgement_note) <= 500)
);

alter table nodescope.alert_rules enable row level security;
alter table nodescope.alerts enable row level security;
create policy alert_rules_runtime_access on nodescope.alert_rules for all to nodescope_runtime using (true) with check (true);
create policy alerts_runtime_access on nodescope.alerts for all to nodescope_runtime using (true) with check (true);
grant select, insert, update, delete on nodescope.alert_rules, nodescope.alerts to nodescope_runtime;

create or replace function nodescope.mcp_set_collection_interval(
  p_actor_id text, p_host_slug text, p_interval_seconds integer
)
returns uuid
language plpgsql security definer set search_path = nodescope, pg_catalog
as $$
declare v_host_id uuid; v_audit_id uuid;
begin
  if p_interval_seconds not between 1 and 60 then raise exception 'interval must be 1..60'; end if;
  if coalesce(p_host_slug, '') = '' then
    update nodescope.collection_settings set global_interval_seconds = p_interval_seconds, updated_at = now();
  else
    select id into v_host_id from nodescope.hosts where slug = p_host_slug and enabled;
    if v_host_id is null then raise exception 'host is unavailable'; end if;
    insert into nodescope.host_collection_overrides(host_id, interval_seconds, updated_at) values (v_host_id, p_interval_seconds, now())
    on conflict (host_id) do update set interval_seconds = excluded.interval_seconds, updated_at = now();
  end if;
  insert into nodescope.audit_events(actor_type, actor_id, action, target_type, target_id, outcome, metadata)
  values ('client', p_actor_id, 'set_collection_interval', case when v_host_id is null then 'fleet' else 'host' end, coalesce(p_host_slug, 'fleet'), 'completed', jsonb_build_object('interval_seconds', p_interval_seconds))
  returning id into v_audit_id;
  return v_audit_id;
end; $$;

create or replace function nodescope.mcp_acknowledge_alert(
  p_actor_id text, p_alert_id text, p_note text
)
returns uuid
language plpgsql security definer set search_path = nodescope, pg_catalog
as $$
declare v_audit_id uuid;
begin
  update nodescope.alerts set state = 'acknowledged', acknowledged_at = now(), acknowledged_by = p_actor_id, acknowledgement_note = nullif(p_note, '')
  where id = p_alert_id and state = 'active';
  if not found then raise exception 'active alert not found'; end if;
  insert into nodescope.audit_events(actor_type, actor_id, action, target_type, target_id, outcome)
  values ('client', p_actor_id, 'acknowledge_alert', 'alert', p_alert_id, 'completed') returning id into v_audit_id;
  return v_audit_id;
end; $$;

create or replace function nodescope.mcp_refresh_storage_baseline(
  p_actor_id text, p_host_slug text
)
returns uuid
language plpgsql security definer set search_path = nodescope, pg_catalog
as $$
declare v_host_id uuid; v_operation_id uuid;
begin
  select id into v_host_id from nodescope.hosts where slug = p_host_slug and enabled;
  if v_host_id is null then raise exception 'host is unavailable'; end if;
  select operation_id into v_operation_id from nodescope.create_operation_with_audit('client', p_actor_id, null, 'refresh_storage_baseline', 'host', p_host_slug, v_host_id, jsonb_build_object('acknowledged_diff', true));
  return v_operation_id;
end; $$;

revoke all on function nodescope.mcp_set_collection_interval(text, text, integer), nodescope.mcp_acknowledge_alert(text, text, text), nodescope.mcp_refresh_storage_baseline(text, text) from public;
grant execute on function nodescope.mcp_set_collection_interval(text, text, integer), nodescope.mcp_acknowledge_alert(text, text, text), nodescope.mcp_refresh_storage_baseline(text, text) to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum) values ('0017_mcp_control_plane', 'tracked-in-repository') on conflict (version) do nothing;
reset role;
