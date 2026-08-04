-- Inference usage metadata only. This table intentionally has no prompt,
-- response, message, completion, header, request-body, or response-body field.
set role nodescope_owner;

create table if not exists nodescope.inference_usage_events (
  id uuid primary key default gen_random_uuid(),
  occurred_at timestamptz not null,
  route_id text not null check (length(route_id) <= 120),
  model text not null check (length(model) <= 240),
  client_id text not null check (length(client_id) <= 120),
  backend_url text not null check (length(backend_url) <= 2048),
  status_code integer not null check (status_code between 100 and 599),
  streaming boolean not null,
  ttft_milliseconds bigint check (ttft_milliseconds >= 0),
  duration_milliseconds bigint not null check (duration_milliseconds >= 0),
  prompt_tokens bigint check (prompt_tokens >= 0),
  output_tokens bigint check (output_tokens >= 0),
  total_tokens bigint check (total_tokens >= 0),
  outcome text not null check (outcome in ('completed', 'backend_error', 'transport_error'))
);

alter table nodescope.inference_usage_events enable row level security;
revoke all on nodescope.inference_usage_events from public, nodescope_runtime;
grant insert on nodescope.inference_usage_events to nodescope_runtime;

create index if not exists inference_usage_events_route_time_idx on nodescope.inference_usage_events (route_id, occurred_at desc);
create index if not exists inference_usage_events_client_time_idx on nodescope.inference_usage_events (client_id, occurred_at desc);

insert into nodescope.schema_migrations (version, source_checksum)
values ('0016_inference_usage_events', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
