-- Run only through the shared-project administrator bootstrap path, after
-- create_operational_roles.sql. This bootstrap creates non-owner login roles
-- without embedding a password. Set each password through an interactive,
-- secret-safe administrator channel after this transaction commits.
\set ON_ERROR_STOP on

begin;

do $$
begin
  if not exists (select 1 from pg_roles where rolname = 'nodescope_verifier_login') then
    create role nodescope_verifier_login login inherit nocreatedb nocreaterole nosuperuser noreplication;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'nodescope_storage_auditor_login') then
    create role nodescope_storage_auditor_login login inherit nocreatedb nocreaterole nosuperuser noreplication;
  end if;
end
$$;

grant nodescope_verifier to nodescope_verifier_login;
grant nodescope_storage_auditor to nodescope_storage_auditor_login;

commit;
