-- Create or rotate the two credentialed NodeScope database roles.
-- Invoke with psql variables runtime_password and migrator_password. The values
-- are never committed; use high-entropy URL-safe secrets.
\set ON_ERROR_STOP on

begin;

do $$
begin
  if not exists (select 1 from pg_roles where rolname = 'nodescope_runtime_login') then
    create role nodescope_runtime_login login noinherit nocreatedb nocreaterole nosuperuser noreplication;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'nodescope_migrate_login') then
    create role nodescope_migrate_login login noinherit nocreatedb nocreaterole nosuperuser noreplication;
  end if;
end
$$;

alter role nodescope_runtime_login password :'runtime_password';
alter role nodescope_migrate_login password :'migrator_password';

grant nodescope_runtime to nodescope_runtime_login;
grant nodescope_migrator to nodescope_migrate_login;
-- The migration login may explicitly SET ROLE nodescope_owner for schema-local
-- DDL. nodescope_owner owns only the nodescope schema and has no membership in
-- any TTRPG-OCR or shared application role.
grant nodescope_owner to nodescope_migrate_login;

alter role nodescope_runtime_login set search_path = nodescope, pg_catalog;
alter role nodescope_migrate_login set search_path = nodescope, pg_catalog;

commit;
