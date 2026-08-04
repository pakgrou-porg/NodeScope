-- Minimal idempotency receipts used when the capacity governor stops raw retention.

set role nodescope_owner;

create table nodescope.ingest_receipts (
  idempotency_key text primary key check (length(trim(idempotency_key)) between 8 and 256),
  agent_id text not null check (length(trim(agent_id)) > 0),
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  received_at timestamptz not null default now(),
  expires_at timestamptz not null,
  raw_retained boolean not null default false,
  check (expires_at > received_at)
);

create index ingest_receipts_expiry_idx on nodescope.ingest_receipts (expires_at);

alter table nodescope.ingest_receipts enable row level security;
create policy ingest_receipts_runtime_access on nodescope.ingest_receipts for all to nodescope_runtime using (true) with check (true);
grant select, insert, update, delete on nodescope.ingest_receipts to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0012_ingest_receipts', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
