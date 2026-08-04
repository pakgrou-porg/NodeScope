-- Remove the disposable sibling-application isolation fixture.
\set ON_ERROR_STOP on
begin;
drop schema if exists nodescope_isolation_fixture cascade;
commit;
