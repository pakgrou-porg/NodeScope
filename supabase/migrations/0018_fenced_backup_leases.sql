-- Fenced database-time lease operations for NodeScope backups. These functions
-- are schema-local and avoid host-clock-based leadership decisions.
set role nodescope_owner;

create or replace function nodescope.acquire_maintenance_lease(
  p_lease_name text,
  p_replica_id text,
  p_ttl_seconds integer
)
returns table(fencing_token bigint, expires_at timestamptz)
language plpgsql security definer set search_path = nodescope, pg_catalog
as $$
declare v_now timestamptz := clock_timestamp();
begin
  if p_ttl_seconds not between 10 and 3600 then raise exception 'lease TTL must be 10..3600 seconds'; end if;
  insert into nodescope.maintenance_leases as lease (lease_name, holder_replica_id, fencing_token, acquired_at, renewed_at, expires_at)
  values (p_lease_name, p_replica_id, 1, v_now, v_now, v_now + make_interval(secs => p_ttl_seconds))
  on conflict (lease_name) do update
  set holder_replica_id = excluded.holder_replica_id,
      fencing_token = case when lease.holder_replica_id = excluded.holder_replica_id then lease.fencing_token else lease.fencing_token + 1 end,
      renewed_at = v_now,
      expires_at = excluded.expires_at
  where lease.expires_at < v_now or lease.holder_replica_id = excluded.holder_replica_id
  returning maintenance_leases.fencing_token, maintenance_leases.expires_at into fencing_token, expires_at;
  if fencing_token is null then raise exception 'maintenance lease is currently held by another replica'; end if;
  return next;
end; $$;

create or replace function nodescope.lease_is_current(
  p_lease_name text, p_replica_id text, p_fencing_token bigint
)
returns boolean
language sql security definer set search_path = nodescope, pg_catalog
as $$
  select exists(
    select 1 from nodescope.maintenance_leases
    where lease_name = p_lease_name
      and holder_replica_id = p_replica_id
      and fencing_token = p_fencing_token
      and expires_at >= clock_timestamp()
  );
$$;

revoke all on function nodescope.acquire_maintenance_lease(text, text, integer), nodescope.lease_is_current(text, text, bigint) from public;
grant execute on function nodescope.acquire_maintenance_lease(text, text, integer), nodescope.lease_is_current(text, text, bigint) to nodescope_runtime;

insert into nodescope.schema_migrations (version, source_checksum)
values ('0018_fenced_backup_leases', 'tracked-in-repository') on conflict (version) do nothing;
reset role;
