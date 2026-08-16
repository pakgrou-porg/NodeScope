# Symlinked Evidence-Input Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; explicit symlink regression coverage, documentation update, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable archive/SBOM fixture, a symlinked archive input, and a fake GitHub CLI. No remote attestation, release publication, host installation, deployment, database, or protected system was contacted. |
| Command | `./scripts/test-verify-agent-release-evidence.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | Direct regular evidence files remain accepted; a symlinked archive input fails before checksum, SPDX, or remote-attestation work; aggregate readiness passes. |
| Observed result | The symlinked archive was rejected as a non-regular input; direct fixture evidence passed; the full aggregate readiness suite passed. |
| Evidence location | [`verify-agent-release-evidence.sh`](../../../scripts/verify-agent-release-evidence.sh), [`test-verify-agent-release-evidence.sh`](../../../scripts/test-verify-agent-release-evidence.sh), and [`verified-releases.md`](../verified-releases.md). |
| Known limitation | Direct-regular-file input validation reduces local path indirection only. It does not prove artifact provenance, remote attestation, release publication, host qualification, or operational acceptance. |
| Rollback or recovery | Do not bypass the regular-file boundary. Place verified direct files in a controlled directory, regenerate checksum sidecars as needed, rerun local verification, and retain the prior accepted artifact as appropriate. |

> The manual evidence path remains non-mutating. Remote attestation and signed-release execution remain separately gated.
