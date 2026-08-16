# Manual Artifact Evidence Hardening

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; the exact-sidecar verifier, structured SPDX parser, regression fixture, documentation, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable archive, checksum, SPDX SBOM, and fake GitHub CLI fixture. No artifact attestation service, release, deployment host, or protected system was contacted. |
| Command | `./scripts/test-verify-agent-release-evidence.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | Valid evidence uses sidecars that exactly name the supplied archive and SBOM; malformed SPDX JSON and a mismatched sidecar filename are rejected; aggregate readiness passes. |
| Observed result | Valid disposable evidence passed; malformed SPDX JSON, revision mismatch, and mismatched sidecar filename were rejected; the full aggregate readiness suite passed. |
| Evidence location | [`verify-agent-release-evidence.sh`](../../../scripts/verify-agent-release-evidence.sh), [`validate-spdx-sbom.mjs`](../../../scripts/validate-spdx-sbom.mjs), and [`test-verify-agent-release-evidence.sh`](../../../scripts/test-verify-agent-release-evidence.sh). |
| Known limitation | This validates local artifact-evidence structure only. It does not verify a real GitHub attestation, download a release, install an agent, update a host, or establish release acceptance. |
| Rollback or recovery | Do not install or promote an unverified artifact. Preserve failure output, correct the checksum sidecar or SBOM, rerun the disposable fixture and aggregate suite, and use the prior accepted signed artifact as appropriate. |

> The verification command remains non-mutating. Remote attestation and signed-release execution remain gated outside this local contract.
