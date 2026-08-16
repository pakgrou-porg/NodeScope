# Manual Clean-Source-Tree Verification

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; the clean-tree gate, regression contract, aggregate integration, documentation update, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable untracked fixture file. No host install, remote attestation, release publication, deployment, database, or protected system was contacted. |
| Command | `./scripts/test-verify-manual-agent-install-contract.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | Manual offline and release-evidence verification refuse any tracked or untracked source-tree change before optional tool checks or operational actions; aggregate readiness passes. |
| Observed result | The disposable untracked file was rejected before the Go tool check; the contract passed; the full aggregate readiness suite passed. |
| Evidence location | [`verify-manual-agent-install.sh`](../../../scripts/verify-manual-agent-install.sh), [`test-verify-manual-agent-install-contract.sh`](../../../scripts/test-verify-manual-agent-install-contract.sh), and [`release-readiness-check.sh`](../../../scripts/release-readiness-check.sh). |
| Known limitation | Clean-tree verification constrains local source provenance only. It does not verify a remote artifact, attestation, release publication, agent installation, host qualification, or operational acceptance. |
| Rollback or recovery | Do not proceed from a dirty tree. Commit or discard the local changes through normal review, rerun manual verification from a clean tree, and retain the prior accepted artifact or source revision as appropriate. |

> The check runs before optional tool prerequisites so a dirty source tree fails deterministically and no host state can be changed.
