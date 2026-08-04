-- NodeScope transactional audit-and-operation protocol.
-- Additive and schema-qualified; safe alongside TTRPG-OCR in the shared project.

set role nodescope_owner;

create or replace function nodescope.create_operation_with_audit(
  p_actor_type nodescope.actor_type,
  p_actor_id text,
  p_actor_user_id uuid,
  p_action text,
  p_target_type text,
  p_target_id text,
  p_host_id uuid,
  p_desired_state jsonb
)
returns table(operation_id uuid, audit_event_id uuid)
language plpgsql
security definer
set search_path = nodescope, pg_catalog
as $$
declare
  new_audit_event_id uuid;
  new_operation_id uuid;
begin
  if length(trim(coalesce(p_actor_id, ''))) = 0 then
    raise exception 'actor ID is required';
  end if;
  if p_action not in ('set_collection_interval', 'refresh_storage_baseline') then
    raise exception 'unsupported Release 1 operation action: %', p_action;
  end if;
  if length(trim(coalesce(p_target_type, ''))) = 0 then
    raise exception 'target type is required';
  end if;

  insert into nodescope.audit_events (
    actor_type,
    actor_id,
    actor_user_id,
    action,
    target_type,
    target_id,
    outcome,
    metadata
  ) values (
    p_actor_type,
    p_actor_id,
    p_actor_user_id,
    p_action,
    p_target_type,
    p_target_id,
    'intent',
    jsonb_build_object('desired_state', coalesce(p_desired_state, '{}'::jsonb))
  ) returning id into new_audit_event_id;

  insert into nodescope.operations (
    audit_event_id,
    host_id,
    action,
    desired_state,
    state
  ) values (
    new_audit_event_id,
    p_host_id,
    p_action,
    coalesce(p_desired_state, '{}'::jsonb),
    'pending'
  ) returning id into new_operation_id;

  return query select new_operation_id, new_audit_event_id;
end;
$$;

revoke all on function nodescope.create_operation_with_audit(
  nodescope.actor_type,
  text,
  uuid,
  text,
  text,
  text,
  uuid,
  jsonb
) from public, anon, authenticated, service_role;
grant execute on function nodescope.create_operation_with_audit(
  nodescope.actor_type,
  text,
  uuid,
  text,
  text,
  text,
  uuid,
  jsonb
) to nodescope_runtime;

reset role;
