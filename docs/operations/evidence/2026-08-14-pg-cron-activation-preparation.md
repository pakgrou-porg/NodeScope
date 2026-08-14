# pg_cron Activation Preparation — Local Validation Record

**Recorded:** 2026-08-14. **Implementation commit:** `205a61ce3c08d4897d7ce79f4a9566b69d631483`. **Environment:** Manus local build sandbox. **Shared Supabase extension or schedule change:** not performed.

> This record validates the tracked NodeScope-namespaced scheduling package. It is not evidence that `pg_cron` is enabled, that any shared-project cron job exists, or that a scheduler worker is active.

## Results

| Evidence field | Record |
| --- | --- |
| Change | Hardens the non-applied schedule definition, adds a dedicated migrator-backed read-only preflight, adds deterministic contract coverage, and documents explicit activation and recovery controls. |
| Expected result | Only the three exact `nodescope-*` jobs can be removed or recreated by the tracked schedule package; the preflight validates extension, routine, maintenance-state, job-set, and recent failure prerequisites without mutating the database. |
| Local validation commands | `bash -n scripts/preflight-nodescope-pg-cron.sh scripts/test-pg-cron-activation-contract.sh scripts/release-readiness-check.sh`; `./scripts/test-pg-cron-activation-contract.sh`; `./scripts/release-readiness-check.sh`. |
| Observed result | All commands passed. The role-specific check printed `NodeScope pg_cron activation contract passed.` The aggregate suite printed `NodeScope release-readiness checks passed.` |
| Evidence locations | [`schedule_maintenance.sql`](../../../supabase/operations/schedule_maintenance.sql), [`preflight-nodescope-pg-cron.sh`](../../../scripts/preflight-nodescope-pg-cron.sh), [`pg-cron-activation.md`](../pg-cron-activation.md), and [`test-pg-cron-activation-contract.sh`](../../../scripts/test-pg-cron-activation-contract.sh). |

## Known limitation and activation gate

The protected environment remains unchanged. Before execution, the owner must explicitly authorize the shared-Supabase action; the current tracked source must pass the disposable shared-project fixture gate; the shared-project administrator must run the read-only preflight; and that preflight must report a complete absence or a complete set of NodeScope-owned jobs. No replica, agent, or browser-console identity may schedule work.

## Recovery path

Before activation, discard the un-applied scheduling revision if review identifies an issue. After an approved activation, if the post-activation preflight fails, stop and have the shared-project administrator remove only the exact tracked `nodescope-rollup-minute`, `nodescope-high-water-minute`, and `nodescope-retention-daily` jobs. Preserve redacted evidence, correct the extension, privilege, scheduler, or routine failure, and rerun the preflight before another activation. Do not modify non-NodeScope jobs or substitute a replica-local timer.
