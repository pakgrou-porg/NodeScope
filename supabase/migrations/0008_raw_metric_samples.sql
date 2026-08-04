-- Exact raw metric history for the configured two-day retention window.

set role nodescope_owner;

create table nodescope.metric_samples (
  batch_id uuid not null references nodescope.telemetry_batches(id) on delete cascade,
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  device_id text not null check (length(trim(device_id)) > 0),
  metric_name text not null check (length(trim(metric_name)) > 0),
  observed_at timestamptz not null,
  numeric_value double precision,
  quality text not null check (quality in ('fresh', 'stale', 'unavailable', 'unsupported', 'estimated')),
  source text not null check (length(trim(source)) > 0),
  semantics text not null check (length(trim(semantics)) > 0),
  expires_at timestamptz not null,
  primary key (batch_id, device_id, metric_name),
  check (numeric_value is not null or quality in ('unavailable', 'unsupported')),
  check (expires_at > observed_at)
);

create index metric_samples_history_idx on nodescope.metric_samples (host_id, metric_name, observed_at desc);
create index metric_samples_expiry_idx on nodescope.metric_samples (expires_at);

alter table nodescope.metric_samples enable row level security;
create policy metric_samples_runtime_access on nodescope.metric_samples for all to nodescope_runtime using (true) with check (true);
grant select, insert, update, delete on nodescope.metric_samples to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0008_raw_metric_samples', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
