-- NodeScope Release 1 foundation migration.
-- Apply to a shared Supabase project that also hosts TTRPG-OCR. NodeScope owns
-- only the `nodescope` schema and its dedicated roles; it must not create,
-- alter, grant on, or migrate any TTRPG-OCR object.
-- This file does not configure SMTP, redirect URLs, or invite settings; bootstrap
-- applies those project-level settings through the Supabase administration workflow.

create extension if not exists pgcrypto;

-- Group roles have no login capability. A later bootstrap step creates a
-- credentialed NodeScope runtime user and grants it nodescope_runtime only.
do $$
begin
  if not exists (select 1 from pg_roles where rolname = 'nodescope_owner') then
    create role nodescope_owner noinherit nologin;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'nodescope_runtime') then
    create role nodescope_runtime noinherit nologin;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'nodescope_migrator') then
    create role nodescope_migrator noinherit nologin;
  end if;
end
$$;

grant nodescope_owner to postgres;
grant nodescope_migrator to postgres;

create schema if not exists nodescope authorization nodescope_owner;

revoke all on schema nodescope from public, anon, authenticated, service_role;
grant usage on schema nodescope to nodescope_runtime, nodescope_migrator;

-- NodeScope stores Supabase Auth UUIDs as opaque identifiers. It deliberately
-- has no foreign key, grant, or read privilege on the shared auth schema.
-- The NodeScope server validates JWTs and enforces that association itself.
set role nodescope_owner;

create type nodescope.role_name as enum ('viewer', 'operator', 'administrator');
create type nodescope.actor_type as enum ('user', 'client', 'agent', 'system');
create type nodescope.operation_state as enum ('pending', 'acknowledged', 'completed', 'failed', 'denied');
create type nodescope.discovery_state as enum ('discovered', 'approved', 'rejected', 'retired');

create table nodescope.user_roles (
  user_id uuid primary key,
  role nodescope.role_name not null default 'viewer',
  invited_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table nodescope.hosts (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique check (slug ~ '^[a-z0-9][a-z0-9-]{0,62}$'),
  display_name text not null check (length(trim(display_name)) > 0),
  platform text not null check (length(trim(platform)) > 0),
  address inet,
  replica_priority smallint,
  enabled boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (replica_priority)
);

create table nodescope.host_capabilities (
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  capability text not null check (length(trim(capability)) > 0),
  state nodescope.discovery_state not null default 'discovered',
  source text not null check (length(trim(source)) > 0),
  details jsonb not null default '{}'::jsonb,
  observed_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (host_id, capability)
);

create table nodescope.agents (
  id uuid primary key default gen_random_uuid(),
  host_id uuid not null unique references nodescope.hosts(id) on delete cascade,
  display_name text not null,
  credential_digest bytea not null unique,
  credential_hint text not null check (length(credential_hint) between 4 and 32),
  enrolled_at timestamptz not null default now(),
  revoked_at timestamptz,
  last_seen_at timestamptz,
  software_version text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table nodescope.collection_settings (
  singleton boolean primary key default true check (singleton),
  global_interval_seconds integer not null default 5 check (global_interval_seconds between 1 and 60),
  updated_by uuid,
  updated_at timestamptz not null default now()
);

insert into nodescope.collection_settings (singleton, global_interval_seconds)
values (true, 5)
on conflict (singleton) do nothing;

create table nodescope.host_collection_overrides (
  host_id uuid primary key references nodescope.hosts(id) on delete cascade,
  interval_seconds integer not null check (interval_seconds between 1 and 60),
  updated_by uuid,
  updated_at timestamptz not null default now()
);

create table nodescope.storage_baselines (
  id uuid primary key default gen_random_uuid(),
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  resource_kind text not null check (resource_kind in ('filesystem', 'network_mount', 'docker_volume')),
  resource_key text not null check (length(trim(resource_key)) > 0),
  expected_state jsonb not null,
  learned_at timestamptz not null default now(),
  refreshed_by uuid,
  retired_at timestamptz,
  unique (host_id, resource_kind, resource_key)
);

create table nodescope.runtime_candidates (
  id uuid primary key default gen_random_uuid(),
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  runtime_kind text not null check (runtime_kind in ('vllm', 'llama_cpp', 'lm_studio', 'other')),
  endpoint text not null,
  discovered_via text not null,
  state nodescope.discovery_state not null default 'discovered',
  details jsonb not null default '{}'::jsonb,
  approved_by uuid,
  approved_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (host_id, endpoint)
);

create table nodescope.audit_events (
  id uuid primary key default gen_random_uuid(),
  request_id uuid not null default gen_random_uuid(),
  actor_type nodescope.actor_type not null,
  actor_id text not null check (length(trim(actor_id)) > 0),
  actor_user_id uuid,
  action text not null check (length(trim(action)) > 0),
  target_type text not null check (length(trim(target_type)) > 0),
  target_id text,
  outcome nodescope.operation_state not null,
  metadata jsonb not null default '{}'::jsonb,
  occurred_at timestamptz not null default now()
);

create index audit_events_occurred_at_idx on nodescope.audit_events (occurred_at desc);
create index audit_events_target_idx on nodescope.audit_events (target_type, target_id, occurred_at desc);

create table nodescope.operations (
  id uuid primary key default gen_random_uuid(),
  audit_event_id uuid not null unique references nodescope.audit_events(id) on delete restrict,
  host_id uuid references nodescope.hosts(id) on delete set null,
  action text not null check (action in ('set_collection_interval', 'refresh_storage_baseline')),
  desired_state jsonb not null,
  state nodescope.operation_state not null default 'pending',
  agent_acknowledged_at timestamptz,
  completed_at timestamptz,
  failure_reason text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index operations_pending_host_idx on nodescope.operations (host_id, created_at asc)
where state in ('pending', 'acknowledged');

create or replace function nodescope.touch_updated_at()
returns trigger
language plpgsql
security invoker
set search_path = nodescope, public
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

create trigger user_roles_touch_updated_at before update on nodescope.user_roles
for each row execute function nodescope.touch_updated_at();
create trigger hosts_touch_updated_at before update on nodescope.hosts
for each row execute function nodescope.touch_updated_at();
create trigger host_capabilities_touch_updated_at before update on nodescope.host_capabilities
for each row execute function nodescope.touch_updated_at();
create trigger agents_touch_updated_at before update on nodescope.agents
for each row execute function nodescope.touch_updated_at();
create trigger runtime_candidates_touch_updated_at before update on nodescope.runtime_candidates
for each row execute function nodescope.touch_updated_at();
create trigger operations_touch_updated_at before update on nodescope.operations
for each row execute function nodescope.touch_updated_at();

-- The shared Supabase project's generic roles receive no NodeScope schema
-- privilege. Browser, TUI, CLI, MCP, and AgentZero requests go through the
-- NodeScope server, which validates shared Supabase JWTs before connecting as
-- the dedicated nodescope_runtime role. The database intentionally does not
-- call auth.uid() or read auth.users, preserving cross-application isolation.

alter table nodescope.user_roles enable row level security;
alter table nodescope.hosts enable row level security;
alter table nodescope.host_capabilities enable row level security;
alter table nodescope.agents enable row level security;
alter table nodescope.collection_settings enable row level security;
alter table nodescope.host_collection_overrides enable row level security;
alter table nodescope.storage_baselines enable row level security;
alter table nodescope.runtime_candidates enable row level security;
alter table nodescope.audit_events enable row level security;
alter table nodescope.operations enable row level security;

-- The server-only database role is constrained to the NodeScope schema. It
-- receives RLS permission only after server-side JWT/role authorization has
-- succeeded; the role has no privilege on TTRPG-OCR schemas or objects.
create policy user_roles_runtime_access on nodescope.user_roles for all to nodescope_runtime using (true) with check (true);
create policy hosts_runtime_access on nodescope.hosts for all to nodescope_runtime using (true) with check (true);
create policy capabilities_runtime_access on nodescope.host_capabilities for all to nodescope_runtime using (true) with check (true);
create policy agents_runtime_access on nodescope.agents for all to nodescope_runtime using (true) with check (true);
create policy settings_runtime_access on nodescope.collection_settings for all to nodescope_runtime using (true) with check (true);
create policy overrides_runtime_access on nodescope.host_collection_overrides for all to nodescope_runtime using (true) with check (true);
create policy baselines_runtime_access on nodescope.storage_baselines for all to nodescope_runtime using (true) with check (true);
create policy runtimes_runtime_access on nodescope.runtime_candidates for all to nodescope_runtime using (true) with check (true);
create policy audit_runtime_access on nodescope.audit_events for all to nodescope_runtime using (true) with check (true);
create policy operations_runtime_access on nodescope.operations for all to nodescope_runtime using (true) with check (true);

revoke all on all tables in schema nodescope from public, anon, authenticated, service_role;
revoke all on all sequences in schema nodescope from public, anon, authenticated, service_role;
grant select, insert, update, delete on all tables in schema nodescope to nodescope_runtime;
grant usage, select on all sequences in schema nodescope to nodescope_runtime;
grant usage on type nodescope.role_name, nodescope.actor_type, nodescope.operation_state, nodescope.discovery_state to nodescope_runtime;

alter default privileges for role nodescope_owner in schema nodescope revoke all on tables from public, anon, authenticated, service_role;
alter default privileges for role nodescope_owner in schema nodescope revoke all on sequences from public, anon, authenticated, service_role;
alter default privileges for role nodescope_owner in schema nodescope grant select, insert, update, delete on tables to nodescope_runtime;
alter default privileges for role nodescope_owner in schema nodescope grant usage, select on sequences to nodescope_runtime;
alter default privileges for role nodescope_owner in schema nodescope grant execute on functions to nodescope_runtime;

reset role;
