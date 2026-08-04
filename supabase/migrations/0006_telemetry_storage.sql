-- Compact telemetry persistence. Raw samples are retained as bounded, idempotent
-- batches; latest state and aggregated rollups drive normal console queries.

set role nodescope_owner;

create table nodescope.telemetry_batches (
  id uuid primary key default gen_random_uuid(),
  idempotency_key text not null unique check (length(trim(idempotency_key)) between 8 and 256),
  agent_id text not null check (length(trim(agent_id)) > 0),
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  boot_id text not null check (length(trim(boot_id)) > 0),
  sequence bigint not null check (sequence >= 0),
  observed_at timestamptz not null,
  received_at timestamptz not null default now(),
  expires_at timestamptz not null,
  compressed_bytes integer not null check (compressed_bytes > 0),
  metric_value_count integer not null check (metric_value_count between 1 and 4096),
  payload jsonb not null,
  unique (agent_id, boot_id, sequence),
  check (expires_at > received_at)
);

create index telemetry_batches_expiry_idx on nodescope.telemetry_batches (expires_at);
create index telemetry_batches_host_observed_idx on nodescope.telemetry_batches (host_id, observed_at desc);

create table nodescope.metric_latest (
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  device_id text not null check (length(trim(device_id)) > 0),
  metric_name text not null check (length(trim(metric_name)) > 0),
  observed_at timestamptz not null,
  received_at timestamptz not null default now(),
  numeric_value double precision,
  text_value text,
  quality text not null check (quality in ('fresh', 'stale', 'unavailable', 'unsupported', 'invalid')),
  source text not null check (length(trim(source)) > 0),
  semantics text not null check (length(trim(semantics)) > 0),
  attributes jsonb not null default '{}'::jsonb,
  primary key (host_id, device_id, metric_name),
  check (numeric_value is not null or text_value is not null or quality in ('unavailable', 'unsupported', 'invalid'))
);

create index metric_latest_host_quality_idx on nodescope.metric_latest (host_id, quality);
create index metric_latest_metric_idx on nodescope.metric_latest (metric_name, observed_at desc);

create table nodescope.metric_rollups (
  resolution_seconds integer not null check (resolution_seconds in (60, 300, 600)),
  bucket_started_at timestamptz not null,
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  device_id text not null check (length(trim(device_id)) > 0),
  metric_name text not null check (length(trim(metric_name)) > 0),
  sample_count integer not null check (sample_count >= 0),
  minimum_value double precision,
  maximum_value double precision,
  average_value double precision,
  last_value double precision,
  p95_sketch bytea,
  source text not null check (length(trim(source)) > 0),
  semantics text not null check (length(trim(semantics)) > 0),
  expires_at timestamptz not null,
  primary key (resolution_seconds, bucket_started_at, host_id, device_id, metric_name),
  check (expires_at > bucket_started_at)
);

create index metric_rollups_history_idx on nodescope.metric_rollups (host_id, metric_name, bucket_started_at desc);
create index metric_rollups_expiry_idx on nodescope.metric_rollups (expires_at);

create table nodescope.storage_probe_summaries (
  id uuid primary key default gen_random_uuid(),
  host_id uuid not null references nodescope.hosts(id) on delete cascade,
  observed_started_at timestamptz not null,
  observed_finished_at timestamptz not null,
  batch_count bigint not null check (batch_count > 0),
  compressed_byte_count bigint not null check (compressed_byte_count > 0),
  metric_value_count bigint not null check (metric_value_count > 0),
  database_bytes bigint,
  reported_by text not null check (length(trim(reported_by)) > 0),
  created_at timestamptz not null default now(),
  check (observed_finished_at > observed_started_at)
);

alter table nodescope.telemetry_batches enable row level security;
alter table nodescope.metric_latest enable row level security;
alter table nodescope.metric_rollups enable row level security;
alter table nodescope.storage_probe_summaries enable row level security;

create policy telemetry_batches_runtime_access on nodescope.telemetry_batches for all to nodescope_runtime using (true) with check (true);
create policy metric_latest_runtime_access on nodescope.metric_latest for all to nodescope_runtime using (true) with check (true);
create policy metric_rollups_runtime_access on nodescope.metric_rollups for all to nodescope_runtime using (true) with check (true);
create policy storage_probe_summaries_runtime_access on nodescope.storage_probe_summaries for all to nodescope_runtime using (true) with check (true);

grant select, insert, update, delete on nodescope.telemetry_batches, nodescope.metric_latest, nodescope.metric_rollups, nodescope.storage_probe_summaries to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0006_telemetry_storage', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
