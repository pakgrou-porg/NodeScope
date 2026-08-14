# Aggregate Shared-Supabase Disposable Fixture Record

**Recorded:** 2026-08-13. **Harness:** `scripts/verify-shared-supabase-fixture.sh supabase/migrations/0015_terminal_fleet_status.sql`. **Environment:** approved shared NodeScope Supabase project. **Protected migration apply:** not performed.

> This disposable exercise composes the required pre-apply database safety controls. It creates only designated isolation fixtures, rolls the selected migration back, and removes the fixtures before returning success.

## Results

| Control | Expected result | Observed result |
| --- | --- | --- |
| Runtime and migrator sibling denial | Both identities are denied `SELECT`, DML, DDL, function execution, and function replacement on a representative sibling schema. | **Passed** for all listed operations. |
| Runtime RLS | `nodescope_runtime` may create its own allowed fixture row only inside a rolled-back probe and may not create another actor’s row. | **Passed**. |
| Catalog isolation | NodeScope ownership, RLS, generic-role denial, sibling-schema denial, and `auth.users` denial verify under the dedicated migrator path. | **Passed**. |
| Agent database boundary | Native agent configuration rejects every supported direct database setting and requires authenticated HTTPS ingestion instead. | **Passed** through `TestLoadConfigRejectsDirectDatabaseConfiguration`. |
| Migration safety | Clean unrecorded migration `0015_terminal_fleet_status.sql` executes only inside `BEGIN`/`ROLLBACK`. | **Passed**. |
| Cleanup and non-persistence | Both fixtures, the `0015` migration-ledger entry, and `nodescope.fleet_ingestion_status()` are absent after the harness. | **Passed** by independent TLS read-only verification. |

## Known limitation

The fixture represents a sibling application and validates the two implemented database login roles. It does not replace a future production apply authorization, an actual TTRPG-OCR object-specific rehearsal, agent-to-server live ingestion, or the separate Supabase Auth and browser-RBAC environment gates.

## Recovery path

The harness cleans fixtures automatically through its trap and explicit teardown. If a future run reports a fixture as present, stop before migration application, execute only the designated fixture cleanup procedure, verify absence through the dedicated migrator’s read-only path, and investigate the failed gate before retrying. If any production apply has occurred, use the approved backup/restore and migration recovery procedure rather than the fixture teardown.

## Related implementation

The [harness](../../../scripts/verify-shared-supabase-fixture.sh), [RLS fixture](../../../supabase/isolation/create_nodescope_rls_fixture.sql), [sibling-denial procedure](../../../scripts/verify-sibling-denials.sh), and [migration application procedure](../../../scripts/apply-nodescope-migration.sh) are all fail-closed on prerequisite or safety-gate failure.
