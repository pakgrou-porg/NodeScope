# Dedicated-Migrator Rollback Preflight Record

**Recorded:** 2026-08-13. **Scope:** clean tracked NodeScope migration `0015_terminal_fleet_status.sql`. **Change authority:** transaction rollback preflight only. **Persistent migration changes:** none.

> This exercise proved that the dedicated migrator can execute the selected clean migration after assuming `nodescope_owner`, while the outer transaction rolls back the function and migration-ledger insert. It is not production-apply authorization.

## Candidate selection

The schema-local migration ledger recorded migrations `0001` through `0014`. Candidate `0003_schema_migration_history.sql` was rejected before execution because its `nodescope.schema_migrations` target already existed. The first tracked, unrecorded candidate was `0015_terminal_fleet_status.sql`.

| Check | Expected result | Observed result |
| --- | --- | --- |
| Source control | Candidate is a clean tracked regular SQL file. | **Passed** |
| Migration ledger before preflight | `0015_terminal_fleet_status` is unrecorded. | **Passed** |
| Fleet-status function before preflight | `nodescope.fleet_ingestion_status()` is absent. | **Passed** |
| Dedicated-migrator execution | Migration runs inside an explicit transaction with `ON_ERROR_STOP`. | **Passed** |
| Rollback verification | The ledger entry remains unrecorded and the fleet-status function remains absent after rollback. | **Passed** |

## Safety declaration

The initial `0003` candidate selection check stopped before migration execution because the migration-history table already existed. The successful `0015` preflight ran only between `BEGIN` and `ROLLBACK`. No NodeScope function, migration-ledger row, table, role, privilege, telemetry row, sibling-schema object, or deployment setting persisted.

## Remaining production-apply prerequisites

The read-only role preflight and disposable sibling-schema denial gate have passed. A production apply still requires a distinct explicit authorization, confirmation that `0015` remains clean and unrecorded immediately before application, the required pre-apply sibling-denial gate, and a post-apply shared-isolation verification using the dedicated migrator.

## Related procedures

Use this evidence with the [read-only preflight record](2026-08-13-shared-supabase-readonly-preflight.md), [sibling-schema denial record](2026-08-13-sibling-schema-denial-gate.md), [activation-gate register](../activation-gates.md), and [migration application procedure](../../../scripts/apply-nodescope-migration.sh).
