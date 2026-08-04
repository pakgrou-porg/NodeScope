-- Migration history is deployment-control metadata, not runtime application data.

set role nodescope_owner;

alter table nodescope.schema_migrations enable row level security;
revoke all on table nodescope.schema_migrations from nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0005_protect_migration_history', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
