-- Run only through the shared-project administrator bootstrap path.
-- This creates no tables, modifies no sibling schema, and grants no cross-project access.
\set ON_ERROR_STOP on

begin;

do $$
begin
  if not exists (select 1 from pg_roles where rolname = 'nodescope_enroller') then
    create role nodescope_enroller noinherit nologin nocreatedb nocreaterole nosuperuser noreplication;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'nodescope_verifier') then
    create role nodescope_verifier noinherit nologin nocreatedb nocreaterole nosuperuser noreplication;
  end if;
  if not exists (select 1 from pg_roles where rolname = 'nodescope_storage_auditor') then
    create role nodescope_storage_auditor noinherit nologin nocreatedb nocreaterole nosuperuser noreplication;
  end if;
end
$$;

commit;
