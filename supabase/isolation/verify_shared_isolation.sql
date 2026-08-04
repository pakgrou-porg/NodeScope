-- Run as a controlled migration/administrator role after applying NodeScope
-- migrations. This script reads catalog metadata only; it never alters another
-- application schema.

begin;

create temporary table nodescope_isolation_results (
  check_name text primary key,
  passed boolean not null,
  detail text not null
) on commit drop;

insert into nodescope_isolation_results
select
  'nodescope schema exists',
  exists (select 1 from pg_namespace where nspname = 'nodescope'),
  'The dedicated NodeScope schema must exist.';

insert into nodescope_isolation_results
select
  'runtime has no sibling schema privileges',
  not exists (
    select 1
    from pg_namespace namespace
    where namespace.nspname not in ('nodescope', 'public', 'pg_catalog', 'information_schema', 'pg_toast')
      and namespace.nspname not like 'pg_temp_%'
      and has_schema_privilege('nodescope_runtime', namespace.oid, 'USAGE')
  ),
  'nodescope_runtime must not have USAGE on any sibling application schema. PostgreSQL grants harmless public-schema USAGE to PUBLIC by default; table and DDL privileges are checked separately.';

insert into nodescope_isolation_results
select
  'runtime has no sibling schema create privilege',
  not exists (
    select 1
    from pg_namespace namespace
    where namespace.nspname not in ('nodescope', 'pg_catalog', 'information_schema', 'pg_toast')
      and namespace.nspname not like 'pg_temp_%'
      and has_schema_privilege('nodescope_runtime', namespace.oid, 'CREATE')
  ),
  'nodescope_runtime must not create objects outside nodescope.';

insert into nodescope_isolation_results
select
  'runtime has no sibling table privileges',
  not exists (
    select 1
    from information_schema.role_table_grants grants
    where grants.grantee = 'nodescope_runtime'
      and grants.table_schema <> 'nodescope'
  ),
  'nodescope_runtime must not have explicit table privileges outside nodescope.';

insert into nodescope_isolation_results
select
  'generic Supabase API roles have no nodescope schema usage',
  not has_schema_privilege('anon', 'nodescope', 'USAGE')
  and not has_schema_privilege('authenticated', 'nodescope', 'USAGE')
  and not has_schema_privilege('service_role', 'nodescope', 'USAGE'),
  'Shared Supabase generic roles must not directly access nodescope.';

insert into nodescope_isolation_results
select
  'nodescope objects have dedicated owner',
  not exists (
    select 1
    from pg_class relation
    join pg_namespace namespace on namespace.oid = relation.relnamespace
    join pg_roles owner_role on owner_role.oid = relation.relowner
    where namespace.nspname = 'nodescope'
      and relation.relkind in ('r', 'p', 'S', 'v', 'm')
      and owner_role.rolname <> 'nodescope_owner'
  ),
  'All NodeScope relations must be owned by nodescope_owner.';

insert into nodescope_isolation_results
select
  'all nodescope tables have RLS enabled',
  not exists (
    select 1
    from pg_class relation
    join pg_namespace namespace on namespace.oid = relation.relnamespace
    where namespace.nspname = 'nodescope'
      and relation.relkind in ('r', 'p')
      and not relation.relrowsecurity
  ),
  'Every NodeScope table must enable RLS.';

insert into nodescope_isolation_results
select
  'runtime cannot select auth users',
  not has_table_privilege('nodescope_runtime', 'auth.users', 'SELECT'),
  'NodeScope may reference auth user IDs but cannot read auth.users.';

select check_name, passed, detail
from nodescope_isolation_results
order by check_name;

-- Convert any failing row into an error so automated bootstrap cannot proceed.
do $$
declare failure_count integer;
begin
  select count(*) into failure_count
  from nodescope_isolation_results
  where not passed;
  if failure_count > 0 then
    raise exception 'NodeScope shared-project isolation verification failed (% checks)', failure_count;
  end if;
end
$$;

rollback;
