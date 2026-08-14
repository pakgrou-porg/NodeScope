# Shared-Supabase Read-Only Preflight Record

**Recorded:** 2026-08-13. **Scope:** the existing NodeScope Supabase database connection. **Change authority:** read-only preflight only. **Database changes made:** none.

> This record establishes that the current connection can reach the intended NodeScope schema and that the supplied NodeScope login roles resolve to their intended roles. It does not authorize a migration, fixture creation, DML, DDL, backup configuration, or deployment.

## Verified observations

| Control | Read-only observation | Result |
| --- | --- | --- |
| Target database | The approved connection reported database `postgres`. | **Passed** |
| NodeScope schema | Schema `nodescope` exists. | **Passed** |
| NodeScope roles | `nodescope_runtime_login`, `nodescope_migrate_login`, `nodescope_runtime`, and `nodescope_owner` exist; all four reported `super=false`. | **Passed** |
| Runtime identity | `nodescope_runtime_login` is a member of `nodescope_runtime`; after `SET ROLE`, the effective role has `USAGE` on `nodescope`, no `CREATE` on `nodescope`, and no `CREATE` on `public`. | **Passed** |
| Migrator identity | `nodescope_migrate_login` is a member of `nodescope_migrate_login` and `nodescope_owner`; after `SET ROLE nodescope_owner`, the effective role has `USAGE` and `CREATE` on `nodescope`, and no `CREATE` on `public`. | **Passed** |
| Existing non-system schemas | Only standard Supabase schemas were observed: `auth`, `cron`, `extensions`, `graphql`, `graphql_public`, `realtime`, `storage`, and `vault`. | **Informational** |

All queries ran inside `BEGIN READ ONLY … ROLLBACK` transactions over TLS. The preflight did not print credentials, connection strings, bearer tokens, private keys, prompts, or completion content.

## Still required before migration or production activation

| Gate | Why this preflight is insufficient | Required next evidence |
| --- | --- | --- |
| Sibling-schema denial | No TTRPG-OCR or disposable sibling-application schema was present to test. | Authorize the controlled create-and-drop disposable sibling fixture and run `scripts/verify-sibling-denials.sh`; retain its redacted result. |
| Table, routine, and RLS boundary | Schema-level privileges do not prove every object-level denial. | Run the dedicated noninterference suite against the approved fixture before any migration. |
| Migration dry run | No specific tracked migration was selected, parsed, or applied. | Select a clean tracked migration and authorize the dedicated-migrator rolled-back preflight. |
| Production application | This record deliberately did not execute DDL or DML. | Obtain a separate explicit approval after the prior gates pass. |

## Safety declaration

No migration, schema creation, fixture creation, role change, privilege change, DML, DDL, data export, retention job, backup job, or deployment action was performed. The next operation requires explicit authorization because it will create and later remove a disposable sibling fixture, even though the test targets no NodeScope production table.

## Related procedures

Use this evidence alongside the [activation-gate register](../activation-gates.md), the [migration application procedure](../../../scripts/apply-nodescope-migration.sh), and the [sibling-denial procedure](../../../scripts/verify-sibling-denials.sh).
