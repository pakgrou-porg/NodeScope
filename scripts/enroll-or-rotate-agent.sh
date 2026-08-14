#!/usr/bin/env bash
# Enroll or rotate one native agent through the least-privilege NodeScope
# enroller role. The raw credential is written only to a caller-selected 0600
# file; stdout contains identifiers and rotation metadata, never the token.
set -euo pipefail
umask 077

usage() {
  cat >&2 <<'USAGE'
usage: scripts/enroll-or-rotate-agent.sh \
  --slug HOST_SLUG --display-name DISPLAY_NAME --platform PLATFORM --address IP_OR_HOST \
  --credential-output ROOT_OWNED_PATH [--expires-days DAYS]

Requires NODESCOPE_ENROLLER_DATABASE_URL for the dedicated nodescope_enroller
login. The script never prints the generated credential. It writes it to the
new credential-output file with mode 0600 after the database function accepts
the enrollment or rotation.
USAGE
  exit 2
}

slug=""
display_name=""
platform=""
address=""
credential_output=""
expires_days="90"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --slug) slug="${2:-}"; shift 2 ;;
    --display-name) display_name="${2:-}"; shift 2 ;;
    --platform) platform="${2:-}"; shift 2 ;;
    --address) address="${2:-}"; shift 2 ;;
    --credential-output) credential_output="${2:-}"; shift 2 ;;
    --expires-days) expires_days="${2:-}"; shift 2 ;;
    *) usage ;;
  esac
done

[[ "$slug" =~ ^[a-z0-9][a-z0-9-]{0,62}$ ]] || { echo "invalid host slug" >&2; exit 1; }
[[ -n "$display_name" && -n "$platform" && -n "$address" ]] || usage
[[ "$expires_days" =~ ^[1-9][0-9]{0,3}$ ]] || { echo "expires-days must be a positive integer" >&2; exit 1; }
[[ -n "$credential_output" && "$(dirname "$credential_output")" != "." ]] || { echo "credential-output must be an absolute or explicit protected path" >&2; exit 1; }
[[ ! -e "$credential_output" && ! -L "$credential_output" ]] || { echo "credential-output already exists or is a symlink" >&2; exit 1; }
[[ -d "$(dirname "$credential_output")" && ! -L "$(dirname "$credential_output")" ]] || { echo "credential-output parent must be a real directory" >&2; exit 1; }
command -v openssl >/dev/null 2>&1 || { echo "openssl is required" >&2; exit 1; }
command -v psql >/dev/null 2>&1 || { echo "psql is required" >&2; exit 1; }
: "${NODESCOPE_ENROLLER_DATABASE_URL:?set the dedicated enroller database URL in a protected environment}"

credential="$(openssl rand -hex 32)"
credential_hint="agent-${credential:0:8}"
temporary_output="${credential_output}.partial.$$"
cleanup() { rm -f -- "$temporary_output"; }
trap cleanup EXIT

# Keep the raw credential off argv. It appears only in psql's standard input,
# the protected partial file, and the database digest expression.
printf '%s\n' "$credential" >"$temporary_output"
chmod 0600 "$temporary_output"

result="$(
  {
    printf "\\set ON_ERROR_STOP on\n"
    printf "\\set slug '%s'\n" "$slug"
    printf "\\set display_name '%s'\n" "${display_name//\'/\'\'}"
    printf "\\set platform '%s'\n" "${platform//\'/\'\'}"
    printf "\\set address '%s'\n" "$address"
    printf "\\set credential '%s'\n" "$credential"
    printf "\\set credential_hint '%s'\n" "$credential_hint"
    printf "\\set expires_days '%s'\n" "$expires_days"
    cat <<'SQL'
select host_id::text, agent_id::text, rotation_version
from nodescope.enroll_or_rotate_agent(
  :'slug',
  :'display_name',
  :'platform',
  :'address'::inet,
  digest(:'credential', 'sha256'),
  :'credential_hint',
  now() + (:'expires_days' || ' days')::interval
);
SQL
  } | psql "$NODESCOPE_ENROLLER_DATABASE_URL" --no-psqlrc --tuples-only --no-align --field-separator='|'
)"

mv -f -- "$temporary_output" "$credential_output"
trap - EXIT
printf 'agent enrollment completed: host=%s credential_file=%s rotation=%s\n' \
  "$slug" "$credential_output" "${result##*|}"
