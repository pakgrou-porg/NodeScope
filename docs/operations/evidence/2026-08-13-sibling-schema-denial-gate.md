# Shared-Supabase Sibling-Schema Denial Gate Record

**Recorded:** 2026-08-13. **Scope:** a disposable `nodescope_isolation_fixture` schema representing a sibling application. **Change authority:** authorized fixture create-and-drop isolation exercise. **Fixture cleanup:** verified.

> The fixture deliberately granted no NodeScope access. The exercise used the live dedicated NodeScope runtime and migrator logins, then removed the fixture before this record was created.

## Verified denial results

| Identity | Effective role under test | Denied operations against the disposable sibling schema | Result |
| --- | --- | --- | --- |
| `nodescope_runtime_login` | `nodescope_runtime` | `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `ALTER TABLE`, `DROP TABLE`, function execution, and function replacement | **Passed** |
| `nodescope_migrate_login` | `nodescope_owner` | `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `ALTER TABLE`, `DROP TABLE`, function execution, and function replacement | **Passed** |

The dedicated `scripts/verify-sibling-denials.sh` procedure reported every required operation as denied and completed with `Shared-project sibling-schema denial gate passed.` A subsequent TLS read-only check confirmed `fixture_cleanup=PASSED`.

## Scope and safety declaration

The only live changes were creation and automatic removal of the disposable fixture schema, its table, and its function. No NodeScope production schema, application migration, NodeScope telemetry, sibling-application production object, credential, role, privilege, backup, or deployment setting was modified.

## Remaining migration prerequisites

This gate proves denial on the approved representative sibling fixture. Before any migration is applied, a clean tracked migration must still be selected and separately authorized for the dedicated-migrator rolled-back preflight. Production DDL requires a distinct explicit approval only after that preflight succeeds.

## Related procedures

Use this evidence with the [read-only preflight record](2026-08-13-shared-supabase-readonly-preflight.md), the [activation-gate register](../activation-gates.md), the [sibling-denial procedure](../../../scripts/verify-sibling-denials.sh), and the [migration application procedure](../../../scripts/apply-nodescope-migration.sh).
