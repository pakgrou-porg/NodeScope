-- Align storage quality values with the typed NodeScope metric domain.

set role nodescope_owner;

alter table nodescope.metric_latest
  drop constraint if exists metric_latest_quality_check;
alter table nodescope.metric_latest
  add constraint metric_latest_quality_check
  check (quality in ('fresh', 'stale', 'unavailable', 'unsupported', 'estimated'));

insert into nodescope.schema_migrations (version, source_checksum)
values ('0007_metric_quality_alignment', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
