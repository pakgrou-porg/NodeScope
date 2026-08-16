# Fixed GitHub CLI Command-Resolution Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; fixed command resolution, controlled-PATH regression fixture, documentation update, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable archive/SBOM fixture and PATH-scoped fake `gh` command. No remote attestation, release publication, host installation, deployment, or protected system was contacted. |
| Command | `./scripts/test-verify-agent-release-evidence.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | The verifier resolves `gh` from its controlled process PATH, ignores `NODESCOPE_GH_BIN`, and preserves all existing malformed-evidence rejection paths; aggregate readiness passes. |
| Observed result | The fixture passed with `NODESCOPE_GH_BIN=/bin/false` and a PATH-scoped `gh` shim; malformed evidence checks still failed closed; the full aggregate readiness suite passed. |
| Evidence location | [`verify-agent-release-evidence.sh`](../../../scripts/verify-agent-release-evidence.sh) and [`test-verify-agent-release-evidence.sh`](../../../scripts/test-verify-agent-release-evidence.sh). |
| Known limitation | Fixed command selection reduces local environment-override risk only. It does not prove the trustworthiness of the operating-system PATH, perform remote attestation, publish a release, qualify a host, or establish operational acceptance. |
| Rollback or recovery | Do not install or promote an artifact if command resolution fails. Preserve failure output, restore the controlled `gh` installation path, rerun local validation, and retain the prior accepted artifact as appropriate. |

> The verifier remains non-mutating. Its remote attestation and signed-release steps remain separately gated.
