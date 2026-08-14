#!/usr/bin/env bash
# Validate structural safety properties of the enrollment wrapper without a live database.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script="$repository_root/scripts/enroll-or-rotate-agent.sh"
migration="$repository_root/supabase/migrations/0014_least_privilege_agent_operations.sql"

[[ -x "$script" || -f "$script" ]] || { echo "missing enrollment wrapper" >&2; exit 1; }
[[ -f "$migration" ]] || { echo "missing enrollment migration" >&2; exit 1; }
for required in 'NODESCOPE_ENROLLER_DATABASE_URL' 'openssl rand -hex 32' 'digest(:' "nodescope.enroll_or_rotate_agent" 'credential-output already exists or is a symlink' 'chmod 0600' 'agent enrollment completed'; do
  grep -Fq -- "$required" "$script" || { echo "enrollment wrapper missing required control: $required" >&2; exit 1; }
done
for forbidden in 'echo "$credential"' 'printf.*credential.*stdout' 'NODESCOPE_RUNTIME_DATABASE_URL'; do
  ! grep -Eq -- "$forbidden" "$script" || { echo "enrollment wrapper contains forbidden credential behavior: $forbidden" >&2; exit 1; }
done
grep -Fq 'security definer' "$migration" || { echo "enrollment function must remain schema-scoped security definer" >&2; exit 1; }
grep -Fq "'rotate_agent_credential'" "$migration" || { echo "enrollment function must retain audit metadata" >&2; exit 1; }

echo "Agent enrollment and rotation contract passed."
