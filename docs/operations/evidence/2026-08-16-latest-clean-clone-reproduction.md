# Latest Baseline Clean-Clone Reproduction

| Evidence field | Record |
| --- | --- |
| Commit | `31dc31bed9b7b01fd97e761b9f1cde3dae701e94` |
| Environment | Fresh external clone at `/tmp/nodescope-clean-clone-latest-20260816/nodescope-clean-31dc31bed9b7b01fd97e761b9f1cde3dae701e94`. No protected database, deployment host, identity provider, inference backend, or release service was contacted. |
| Command | `./scripts/reproduce-clean-clone-readiness.sh --commit 31dc31bed9b7b01fd97e761b9f1cde3dae701e94 --workspace-root /tmp/nodescope-clean-clone-latest-20260816` |
| Expected result | Clone the exact published revision outside the repository, install locked dependencies, pass the full deterministic readiness suite, and leave the detached clone clean. |
| Observed result | The procedure emitted `CLEAN_CLONE_REPRODUCTION_PASSED`, verified the requested commit, passed the aggregate readiness suite, and reported `status=clean`. |
| Evidence location | [`reproduce-clean-clone-readiness.sh`](../../../scripts/reproduce-clean-clone-readiness.sh) and [clean-clone procedure](../clean-clone-reproduction.md). |
| Known limitation | A fresh local clone does not establish protected Supabase behavior, real identity/RBAC E2E, host telemetry, replica recovery, PKI revocation, isolated restore, approved-backend streaming, release attestation, tagged publication, host qualification, or operational acceptance. |
| Rollback or recovery | Do not promote a release from a failed reproduction. Preserve output, remove the disposable clone, correct the failing control in review, then rerun from a new external workspace against the intended published commit. |

> The output clone was created outside the repository and the procedure's final clean-status check confirmed no generated material or local configuration was retained in the checked revision.
