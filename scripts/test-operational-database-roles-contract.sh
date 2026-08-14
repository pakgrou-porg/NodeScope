#!/usr/bin/env bash
# Keep routine database evidence identities narrow, password-free in source, and
# distinct from owner/migration/enrollment identities.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
migration="$repository_root/supabase/migrations/0019_operational_auditor_privilege_boundary.sql"
bootstrap="$repository_root/supabase/isolation/create_operational_login_roles.sql"
guide="$repository_root/docs/operations/operational-database-roles.md"
denial_gate="$repository_root/scripts/verify-operational-role-denials.sh"

for required in \
  'revoke all on all tables in schema nodescope from nodescope_verifier, nodescope_storage_auditor;' \
  'revoke all on all sequences in schema nodescope from nodescope_verifier, nodescope_storage_auditor;' \
  'revoke all on all functions in schema nodescope from nodescope_verifier, nodescope_storage_auditor;' \
  'grant execute on function nodescope.host_ingestion_status(text) to nodescope_verifier;' \
  'grant execute on function nodescope.fleet_ingestion_status() to nodescope_verifier;' \
  'grant execute on function nodescope.storage_probe_evidence(text, timestamptz) to nodescope_storage_auditor;' \
  'alter default privileges for role nodescope_owner in schema nodescope'; do
  grep -Fq -- "$required" "$migration" || { echo "operational database role migration missing: $required" >&2; exit 1; }
done

for required in \
  'create role nodescope_verifier_login login inherit nocreatedb nocreaterole nosuperuser noreplication;' \
  'create role nodescope_storage_auditor_login login inherit nocreatedb nocreaterole nosuperuser noreplication;' \
  'grant nodescope_verifier to nodescope_verifier_login;' \
  'grant nodescope_storage_auditor to nodescope_storage_auditor_login;'; do
  grep -Fq -- "$required" "$bootstrap" || { echo "operational login bootstrap missing: $required" >&2; exit 1; }
done

if grep -Eqi 'password[[:space:]]+['"'"']|password[[:space:]]*=' "$bootstrap"; then
  echo 'operational login bootstrap must not embed a password' >&2
  exit 1
fi

for required in \
  'must not be granted `CREATEROLE`' \
  'NODESCOPE_VERIFIER_DATABASE_URL' \
  'NODESCOPE_STORAGE_AUDITOR_DATABASE_URL' \
  'direct NodeScope table access' \
  'sibling-schema access' \
  'do not fall back to an owner connection'; do
  grep -Fq -- "$required" "$guide" || { echo "operational database role guide missing: $required" >&2; exit 1; }
done

for required in \
  'NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL' \
  'NODESCOPE_VERIFIER_DATABASE_URL' \
  'NODESCOPE_STORAGE_AUDITOR_DATABASE_URL' \
  'expect_denied nodescope_verifier' \
  'expect_denied nodescope_storage_auditor' \
  'expect_allowed nodescope_verifier' \
  'expect_allowed nodescope_storage_auditor' \
  'select * from nodescope.hosts' \
  'select * from nodescope.metric_samples' \
  'nodescope_isolation_fixture.documents' \
  'nodescope.host_ingestion_status' \
  'nodescope.fleet_ingestion_status' \
  'nodescope.storage_probe_evidence' \
  'Operational verifier and storage-auditor denial gate passed.'; do
  grep -Fq -- "$required" "$denial_gate" || { echo "operational database role denial gate missing: $required" >&2; exit 1; }
done

echo 'Operational database roles contract passed.'
