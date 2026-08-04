-- Full discovered Docker container inventory; alerts remain separately selected.

set role nodescope_owner;

create table nodescope.container_inventory (
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  container_id text not null check (length(trim(container_id)) > 0),
  name text not null check (length(trim(name)) > 0),
  image text not null check (length(trim(image)) > 0),
  state text not null check (length(trim(state)) > 0),
  health text not null default 'unreported' check (length(trim(health)) > 0),
  selected_for_alerting boolean not null default false,
  observed_at timestamptz not null,
  updated_at timestamptz not null default now(),
  primary key (host_id, container_id)
);

create index container_inventory_host_state_idx on nodescope.container_inventory (host_id, state, observed_at desc);

alter table nodescope.container_inventory enable row level security;
create policy container_inventory_runtime_access on nodescope.container_inventory for all to nodescope_runtime using (true) with check (true);
grant select, insert, update, delete on nodescope.container_inventory to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0013_container_inventory', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
