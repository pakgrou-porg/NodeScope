# Operational Database Roles — Local Validation Record

**Recorded:** 2026-08-14. **Implementation commit:** `249fc12f348c3a3609e679f0882a8a465ac44e12`. **Environment:** Manus local build sandbox. **Protected Supabase migration apply:** not performed.

> This record covers the locally buildable access-boundary implementation only. It is not evidence that `0016_operational_auditor_privilege_boundary.sql` or either authenticated routine login exists in the protected shared Supabase project.

## Results

| Evidence field | Record |
| --- | --- |
| Change | Adds `0016_operational_auditor_privilege_boundary.sql`, a password-free shared-project administrator bootstrap for distinct verifier and storage-auditor logins, an authorized disposable-fixture denial gate, an operational runbook, and a deterministic release-readiness contract. |
| Expected result | Routine verification and receipt-time storage evidence can use distinct non-owner logins that are denied direct NodeScope table/sequence access, unassigned function access, and sibling-schema access. |
| Local validation commands | `bash -n scripts/test-operational-database-roles-contract.sh scripts/verify-operational-role-denials.sh scripts/release-readiness-check.sh`; `./scripts/test-operational-database-roles-contract.sh`; `go test ./cmd/nodescope-verify ./cmd/nodescope-storage-evidence`; `./scripts/release-readiness-check.sh`. |
| Observed result | All commands passed. The aggregate suite printed `NodeScope release-readiness checks passed.` The role-specific contract printed `Operational database roles contract passed.` |
| Evidence locations | [`0016_operational_auditor_privilege_boundary.sql`](../../../supabase/migrations/0016_operational_auditor_privilege_boundary.sql), [`create_operational_login_roles.sql`](../../../supabase/isolation/create_operational_login_roles.sql), [`verify-operational-role-denials.sh`](../../../scripts/verify-operational-role-denials.sh), and [`operational-database-roles.md`](../operational-database-roles.md). |

## Known limitation and activation gate

No protected database change occurred. The migration must first pass the tracked disposable shared-Supabase fixture process; the shared-project administrator must then create the two login identities without embedding passwords, set distinct secrets through an approved secret channel, and run the authorized `verify-operational-role-denials.sh` fixture gate. A successful local contract does not prove a Supabase privilege configuration until that authorized environment validation is complete.

## Recovery path

Before a live application, recovery is simply to discard the un-applied migration and bootstrap revision. After an approved live activation, if the denial gate fails or a credential is suspected exposed, stop routine use, revoke the affected login's group membership or disable that login through the shared-project administrator, preserve only redacted evidence, and do not substitute `nodescope_owner`. Investigate and rerun the disposable-fixture gate before any re-enable. Use the established migration and backup/restore procedures, not ad hoc privilege broadening, for an applied schema change.
