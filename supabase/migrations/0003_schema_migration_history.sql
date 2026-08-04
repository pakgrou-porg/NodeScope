-- NodeScope migration history is schema-local and maintained only by the
-- dedicated nodescope_migrate_login after SET ROLE nodescope_owner.

set role nodescope_owner;

create table nodescope.schema_migrations (
  version text primary key,
  applied_at timestamptz not null default now(),
  applied_by text not null default current_user,
  source_checksum text not null
);

insert into nodescope.schema_migrations (version, source_checksum)
values
  ('0001_nodescope_foundation', 'tracked-in-repository'),
  ('0002_transactional_operations', 'tracked-in-repository'),
  ('0003_schema_migration_history', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
