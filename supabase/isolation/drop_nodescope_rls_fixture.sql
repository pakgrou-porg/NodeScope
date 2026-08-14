-- Remove the disposable NodeScope RLS fixture.
\set ON_ERROR_STOP on

begin;
set role nodescope_owner;
drop table if exists nodescope.rls_isolation_fixture;
reset role;
commit;
