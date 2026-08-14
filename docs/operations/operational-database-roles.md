# Routine Verifier and Storage-Auditor Database Roles

NodeScope uses two independent routine database identities so status verification and storage feasibility evidence do not require `nodescope_owner`, the server runtime identity, the agent enrollment identity, or the migration identity. The **verifier** can invoke only metadata-safe ingestion-status functions. The **storage auditor** can invoke only the receipt-time storage-evidence function. Neither role receives table, sequence, DDL, migration, enrollment, or sibling-schema privileges.

> `nodescope_verifier` and `nodescope_storage_auditor` are **no-login group roles**. Their corresponding `*_login` roles authenticate, inherit only their group role, and must be created by the shared-project administrator. The dedicated NodeScope migrator must not be granted `CREATEROLE` merely to create routine logins.

## Access boundary

| Routine identity | Login role | Permitted NodeScope interface | Explicitly prohibited |
| --- | --- | --- | --- |
| Verifier | `nodescope_verifier_login` | `nodescope.host_ingestion_status(text)` and `nodescope.fleet_ingestion_status()` | Direct table or sequence access; storage evidence; enrollment; writes; DDL; migrations |
| Storage auditor | `nodescope_storage_auditor_login` | `nodescope.storage_probe_evidence(text, timestamptz)` | Direct table or sequence access; host/fleet status; enrollment; writes; DDL; migrations |

The three permitted functions are schema-local `SECURITY DEFINER` functions with a fixed `nodescope, pg_temp` search path. They return summary evidence only; they never return credential digests, raw agent credentials, database URLs, prompt content, response content, or sibling-application data.

## Administrator bootstrap — confirmation-required

This is a protected-environment change. Perform it only after the migration has passed the disposable shared-Supabase fixture gate and an administrator has approved the live activation. The following commands are for the shared-project administrator, not a NodeScope agent, server replica, auxiliary agent, migration runner, or browser console.

```bash
set -euo pipefail
cd /path/to/NodeScope

# Creates no-login groups only. This is idempotent and must run before login creation.
psql "$NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL" \
  --no-psqlrc -v ON_ERROR_STOP=1 \
  -f supabase/isolation/create_operational_roles.sql

# Creates restricted login identities but deliberately does not supply passwords.
psql "$NODESCOPE_SHARED_PROJECT_ADMIN_DATABASE_URL" \
  --no-psqlrc -v ON_ERROR_STOP=1 \
  -f supabase/isolation/create_operational_login_roles.sql
```

Set each login password only through a secret-safe interactive administrative channel after the SQL transaction succeeds. Do not place a password in a command argument, SQL file, shell history, Git repository, Portainer stack variable, browser form, chat transcript, or terminal output. Store the resulting complete connection strings in separate root-owned credential files or a managed secret store. The verifier and storage-auditor URLs must use different passwords.

## Connection and validation

Point `NODESCOPE_VERIFIER_DATABASE_URL` at `nodescope_verifier_login`; point `NODESCOPE_STORAGE_AUDITOR_DATABASE_URL` at `nodescope_storage_auditor_login`. A verifier URL must not be reused for the storage auditor, a replica, a migrator, or enrollment.

```bash
sudo install -d -o root -g root -m 0700 /etc/nodescope/credentials
sudo install -o root -g root -m 0600 /dev/null /etc/nodescope/credentials/verifier.env
sudo install -o root -g root -m 0600 /dev/null /etc/nodescope/credentials/storage-auditor.env

# Populate the two files manually through the approved secret mechanism.
# Each file contains exactly one distinct connection-string variable.
```

Run the existing native commands only from a controlled administrator shell after their matching credential file is sourced. The commands call only their permitted function; do not substitute `nodescope_owner` if a routine credential fails.

```bash
set -a
. /etc/nodescope/credentials/verifier.env
set +a
sudo /usr/local/bin/nodescope-verify --slug framework --require-fresh

set -a
. /etc/nodescope/credentials/storage-auditor.env
set +a
sudo /usr/local/bin/nodescope-storage-evidence \
  --slug framework \
  --since "$(date -u -d '72 hours ago' +%Y-%m-%dT%H:%M:%SZ)" \
  --collection-interval-seconds 5 \
  --require-complete
```

The live activation record must demonstrate all four conditions: each login can execute only its assigned function; each login is denied direct NodeScope table access; each login is denied sibling-schema access; and no routine procedure uses `nodescope_owner`. If a routine credential is lost or suspected exposed, disable or rotate only the affected login through the shared-project administrator, preserve redacted evidence, and do not fall back to an owner connection.

After explicit authorization, the administrator can capture the required denial evidence through the disposable-fixture gate below. It creates and then removes `nodescope_isolation_fixture`; it does not alter NodeScope telemetry, hosts, agents, or configuration. Never run it against an unmanaged sibling schema.

```bash
set -a
. /etc/nodescope/credentials/shared-project-admin.env
. /etc/nodescope/credentials/verifier.env
. /etc/nodescope/credentials/storage-auditor.env
set +a

./scripts/verify-operational-role-denials.sh
```
