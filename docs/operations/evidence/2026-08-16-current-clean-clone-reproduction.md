# Current Baseline Clean-Clone Reproduction

| Evidence field | Record |
| --- | --- |
| Commit | `d0b812990420f78d99638ef85ddb3b7dfd9ae60f` |
| Environment | Fresh external clone at `/tmp/nodescope-clean-clones-20260816/nodescope-clean-d0b812990420f78d99638ef85ddb3b7dfd9ae60f`; no protected environment was contacted. |
| Command | `./scripts/reproduce-clean-clone-readiness.sh --commit d0b812990420f78d99638ef85ddb3b7dfd9ae60f --workspace-root /tmp/nodescope-clean-clones-20260816` |
| Expected result | Clone the exact GitHub revision into an external workspace, install locked dependencies, pass the aggregate readiness suite, and leave the detached checkout clean. |
| Observed result | The procedure completed with `CLEAN_CLONE_REPRODUCTION_PASSED`, the requested detached commit, and `status=clean`. |
| Evidence location | [`scripts/reproduce-clean-clone-readiness.sh`](../../../scripts/reproduce-clean-clone-readiness.sh) and [clean-clone procedure](../clean-clone-reproduction.md). |
| Known limitation | Fresh-clone reproducibility does not establish host telemetry, protected Supabase behavior, real identity/RBAC E2E, replica recovery, PKI revocation, approved runtime streaming, tagged release publication, or operational acceptance. |
| Rollback or recovery | Preserve the failed output, do not promote a release, restore the prior accepted signed tag as appropriate, remediate the failing control, and rerun from a new external clone. |

> The procedure used a new destination, detached the exact published commit, and verified the generated-contract and final working-tree boundaries. The clone is disposable evidence material outside the repository.
