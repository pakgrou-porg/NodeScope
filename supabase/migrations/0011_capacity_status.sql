-- Shared capacity state lets every replica render the same retention circuit-breaker decision.

set role nodescope_owner;

create table nodescope.capacity_status (
  singleton boolean primary key default true check (singleton),
  measured_at timestamptz not null default now(),
  used_bytes bigint not null check (used_bytes >= 0),
  quota_bytes bigint not null check (quota_bytes > 0),
  mode text not null check (mode in ('normal', 'constrained', 'summary_only', 'protective')),
  accept_raw_batches boolean not null,
  accept_summary_rollups boolean not null,
  detail text not null check (length(trim(detail)) > 0),
  source text not null check (length(trim(source)) > 0)
);

alter table nodescope.capacity_status enable row level security;
create policy capacity_status_runtime_access on nodescope.capacity_status for all to nodescope_runtime using (true) with check (true);
grant select, insert, update, delete on nodescope.capacity_status to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0011_capacity_status', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
