-- Fenced maintenance leases support backup, retention, and rollout leadership
-- without allowing concurrent replicas to perform the same privileged task.

set role nodescope_owner;

create table nodescope.maintenance_leases (
  lease_name text primary key check (length(trim(lease_name)) > 0),
  holder_replica_id text not null check (length(trim(holder_replica_id)) > 0),
  fencing_token bigint not null default 0,
  acquired_at timestamptz not null default now(),
  renewed_at timestamptz not null default now(),
  expires_at timestamptz not null,
  check (expires_at > acquired_at)
);

alter table nodescope.maintenance_leases enable row level security;
create policy maintenance_leases_runtime_access
on nodescope.maintenance_leases
for all to nodescope_runtime
using (true)
with check (true);

grant select, insert, update, delete on nodescope.maintenance_leases to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0004_maintenance_leases', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
