# Migration 0015 Production Apply Record

**Migration:** `0015_terminal_fleet_status.sql`. **Authority:** explicit user authorization received before persistent application. **Environment:** approved shared NodeScope Supabase project. **Content access:** none; this record contains schema and privilege evidence only.

> The migration was applied only after the dedicated migrator reran the disposable sibling-schema denial gate and executed the exact SQL in a rolled-back preflight. The post-apply verification was repeated independently through the least-privilege migrator path.

## Pre-apply state

| Check | Expected result | Observed result |
| --- | --- | --- |
| Source control | Migration is tracked, regular, direct under `supabase/migrations`, and clean. | **Passed.** |
| Migration ledger | `0015_terminal_fleet_status` is absent. | **Passed** (`false`). |
| Target function | `nodescope.fleet_ingestion_status()` is absent. | **Passed** (`false`). |
| Sibling denial | Both runtime and migrator login paths are denied tested read, DML, DDL, execution, and replacement operations on the disposable sibling fixture. | **Passed** for all tested operations. |
| Rollback preflight | Exact SQL executes under the dedicated migrator inside an explicit transaction and rolls back. | **Passed.** |

## Persistent apply and independent verification

| Check | Expected result | Observed result |
| --- | --- | --- |
| Migration application | Dedicated migrator applies the selected migration after gates pass. | **Passed.** |
| Migration ledger | `0015_terminal_fleet_status` is recorded. | **Passed** (`true`). |
| Target function | `nodescope.fleet_ingestion_status()` exists. | **Passed** (`true`). |
| NodeScope RLS | Every NodeScope table has RLS enabled. | **Passed.** |
| Ownership and generic roles | NodeScope objects retain the dedicated owner and generic Supabase API roles lack NodeScope schema usage. | **Passed.** |
| Cross-schema denial | Runtime retains no sibling-schema create, schema usage, or table privilege; runtime cannot read `auth.users`. | **Passed.** |

## Known limitation

This applies only migration `0015`. It does not activate a server replica, enroll an agent, create user identities, exercise browser authentication, qualify Framework hardware, test production inference streaming, create a backup, or validate an isolated restore. The new fleet-ingestion function is present but has not yet received real host telemetry.

## Recovery

No destructive down migration is defined for `0015`. If the function must be withdrawn before dependent release activation, pause control-plane deployment, use a separately approved migration that removes only the function and migration-ledger record after impact review, then rerun the dedicated shared-isolation verifier. Do not alter sibling schemas, generic roles, or `auth` objects as a migration recovery shortcut.

## Procedures and evidence

The apply procedure is [`scripts/apply-nodescope-migration.sh`](../../../scripts/apply-nodescope-migration.sh); the isolation verifier is [`supabase/isolation/verify_shared_isolation.sql`](../../../supabase/isolation/verify_shared_isolation.sql). Related pre-apply evidence is the [rollback preflight](2026-08-13-migrator-rollback-preflight.md) and [aggregate fixture](2026-08-13-aggregate-shared-supabase-fixture.md).
