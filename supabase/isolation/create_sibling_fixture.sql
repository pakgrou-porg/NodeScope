-- Disposable fixture representing a sibling application such as TTRPG-OCR.
-- Execute only in the shared-project isolation rehearsal, then remove with
-- drop_sibling_fixture.sql. It deliberately grants no NodeScope role access.
\set ON_ERROR_STOP on

begin;
drop schema if exists nodescope_isolation_fixture cascade;
create schema nodescope_isolation_fixture;

create table nodescope_isolation_fixture.documents (
  id uuid primary key default gen_random_uuid(),
  title text not null,
  payload jsonb not null default '{}'::jsonb
);

create or replace function nodescope_isolation_fixture.count_documents()
returns bigint
language sql
security invoker
set search_path = nodescope_isolation_fixture, pg_catalog
as $$
  select count(*) from nodescope_isolation_fixture.documents
$$;

revoke all on schema nodescope_isolation_fixture from public, anon, authenticated, service_role, nodescope_owner, nodescope_runtime, nodescope_migrator, nodescope_runtime_login, nodescope_migrate_login;
revoke all on all tables in schema nodescope_isolation_fixture from public, anon, authenticated, service_role, nodescope_owner, nodescope_runtime, nodescope_migrator, nodescope_runtime_login, nodescope_migrate_login;
revoke all on all functions in schema nodescope_isolation_fixture from public, anon, authenticated, service_role, nodescope_owner, nodescope_runtime, nodescope_migrator, nodescope_runtime_login, nodescope_migrate_login;

commit;
