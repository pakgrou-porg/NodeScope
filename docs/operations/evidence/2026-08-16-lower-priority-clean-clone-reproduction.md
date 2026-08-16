# Lower-Priority Hardening Baseline Clean-Clone Reproduction

| Evidence field | Record |
| --- | --- |
| Commit | `77997a3ad44f44654ccd251f622fa0b5f81f80be` |
| Environment | Fresh external clone at `/tmp/nodescope-clean-clone-lower-priority-20260816/nodescope-clean-77997a3ad44f44654ccd251f622fa0b5f81f80be`. No protected database, deployment host, identity provider, inference backend, or release service was contacted. |
| Command | `./scripts/reproduce-clean-clone-readiness.sh --commit 77997a3ad44f44654ccd251f622fa0b5f81f80be --workspace-root /tmp/nodescope-clean-clone-lower-priority-20260816` |
| Expected result | Clone the exact published revision outside the repository, install locked dependencies, pass the complete deterministic readiness suite, and leave the detached clone clean. |
| Observed result | The procedure emitted `CLEAN_CLONE_REPRODUCTION_PASSED`, verified the requested commit, passed the aggregate readiness suite, and reported `status=clean`. |
| Evidence location | [`reproduce-clean-clone-readiness.sh`](../../../scripts/reproduce-clean-clone-readiness.sh) and [clean-clone procedure](../clean-clone-reproduction.md). |
| Known limitation | Fresh local reproduction does not establish protected Supabase behavior, real identity/RBAC E2E, host telemetry, replica recovery, PKI revocation, isolated restore, approved-backend streaming, release attestation, tagged publication, host qualification, or operational acceptance. |
| Rollback or recovery | Do not promote a release from a failed reproduction. Preserve output, remove the disposable clone, correct the failing control in review, and rerun from a new external workspace against the intended published commit. |

> The disposable clone remained outside the repository. Its final clean-status check confirms generated output and local configuration were not retained in the tested revision.
