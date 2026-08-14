-- Disposable NodeScope-owned fixture for proving runtime RLS enforcement.
-- Execute only through scripts/verify-shared-supabase-fixture.sh.
\set ON_ERROR_STOP on

begin;
set role nodescope_owner;

drop table if exists nodescope.rls_isolation_fixture;
create table nodescope.rls_isolation_fixture (
  actor text not null,
  payload text not null
);
alter table nodescope.rls_isolation_fixture enable row level security;
alter table nodescope.rls_isolation_fixture force row level security;

revoke all on nodescope.rls_isolation_fixture from public, anon, authenticated, service_role;
grant select, insert, update, delete on nodescope.rls_isolation_fixture to nodescope_runtime;

create policy runtime_actor_only on nodescope.rls_isolation_fixture
  for all to nodescope_runtime
  using (actor = current_user)
  with check (actor = current_user);

reset role;
commit;
