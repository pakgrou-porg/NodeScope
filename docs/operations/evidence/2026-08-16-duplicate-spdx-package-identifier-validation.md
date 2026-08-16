# Duplicate SPDX Package-Identifier Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; duplicate-ID rejection, regression fixture, documentation update, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable archive/SBOM fixture and fake GitHub CLI. No remote attestation, release publication, host installation, deployment, or protected system was contacted. |
| Command | `./scripts/test-verify-agent-release-evidence.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | Valid package identifiers remain accepted; duplicate SPDX package identifiers fail closed; aggregate readiness passes. |
| Observed result | The valid fixture passed; a duplicated `SPDXRef-nodescope` was rejected; the full aggregate readiness suite passed. |
| Evidence location | [`validate-spdx-sbom.mjs`](../../../scripts/validate-spdx-sbom.mjs) and [`test-verify-agent-release-evidence.sh`](../../../scripts/test-verify-agent-release-evidence.sh). |
| Known limitation | Identifier uniqueness improves local SBOM structure validation only. It does not prove package completeness, artifact provenance, remote attestation, release publication, host qualification, or operational acceptance. |
| Rollback or recovery | Do not install or promote an artifact with duplicate SPDX identifiers. Preserve failure output, regenerate a valid SBOM, rerun the local verifier and aggregate suite, and retain the prior accepted artifact as appropriate. |

> The manual evidence path remains non-mutating. Its remote attestation and signed-release boundaries remain separate gates.
