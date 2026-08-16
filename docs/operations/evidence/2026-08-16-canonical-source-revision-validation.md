# Canonical Source-Revision Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; canonical revision enforcement, regression fixture, documentation update, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable archive/SBOM fixture and fake GitHub CLI. No remote attestation, release publication, host installation, deployment, or protected system was contacted. |
| Command | `./scripts/test-verify-agent-release-evidence.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | A canonical 40-character GitHub commit revision passes the local format gate; an unsupported 64-character revision fails closed; aggregate readiness passes. |
| Observed result | The canonical fixture passed; a 64-character revision was rejected before remote resolution; the full aggregate readiness suite passed. |
| Evidence location | [`verify-agent-release-evidence.sh`](../../../scripts/verify-agent-release-evidence.sh) and [`test-verify-agent-release-evidence.sh`](../../../scripts/test-verify-agent-release-evidence.sh). |
| Known limitation | Canonical revision formatting improves local input validation only. It does not prove a remote tag target, artifact provenance, attestation, release publication, host qualification, or operational acceptance. |
| Rollback or recovery | Do not install or promote an artifact with an unsupported source-revision format. Preserve failure output, use the 40-character source revision from verified release evidence, rerun local validation, and retain the prior accepted artifact as appropriate. |

> The manual evidence path remains non-mutating. Remote tag resolution and signed-release execution remain separate gates.
