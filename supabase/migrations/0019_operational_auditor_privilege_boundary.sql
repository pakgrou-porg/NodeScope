-- Constrain routine verification and storage evidence to explicit schema-local
-- security-definer functions. Login creation remains outside migrations because
-- the dedicated NodeScope migrator intentionally has no CREATEROLE privilege.

set role nodescope_owner;

revoke all on all tables in schema nodescope from nodescope_verifier, nodescope_storage_auditor;
revoke all on all sequences in schema nodescope from nodescope_verifier, nodescope_storage_auditor;
revoke all on all functions in schema nodescope from nodescope_verifier, nodescope_storage_auditor;

grant usage on schema nodescope to nodescope_verifier, nodescope_storage_auditor;
grant execute on function nodescope.host_ingestion_status(text) to nodescope_verifier;
grant execute on function nodescope.fleet_ingestion_status() to nodescope_verifier;
grant execute on function nodescope.storage_probe_evidence(text, timestamptz) to nodescope_storage_auditor;

alter default privileges for role nodescope_owner in schema nodescope
  revoke all on tables from nodescope_verifier, nodescope_storage_auditor;
alter default privileges for role nodescope_owner in schema nodescope
  revoke all on sequences from nodescope_verifier, nodescope_storage_auditor;
alter default privileges for role nodescope_owner in schema nodescope
  revoke execute on functions from nodescope_verifier, nodescope_storage_auditor;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0019_operational_auditor_privilege_boundary', 'tracked-in-repository')
on conflict (version) do nothing;

reset role;
