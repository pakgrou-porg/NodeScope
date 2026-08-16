# Release-Evidence Manifest Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; manifest parser, assembler hardening, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable release-artifact fixture. No release was created, no artifact was uploaded, and no protected database, deployment host, identity provider, or inference backend was contacted. |
| Command | `./scripts/test-release-evidence-contract.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | A safe fixture under a canonical tag and immutable revision produces a parsed manifest; unsafe archive names are rejected; aggregate readiness passes. |
| Observed result | The safe fixture assembled and parsed successfully; an unsafe archive name was rejected; the full aggregate readiness suite passed. |
| Evidence location | [`assemble-release-evidence.sh`](../../../scripts/assemble-release-evidence.sh), [`verify-release-evidence-manifest.mjs`](../../../scripts/verify-release-evidence-manifest.mjs), and [`test-release-evidence-contract.sh`](../../../scripts/test-release-evidence-contract.sh). |
| Known limitation | Manifest integrity checks do not execute a signed tag, publish a GitHub release, verify a remote attestation, validate an artifact download, or establish operational acceptance. |
| Rollback or recovery | Do not promote a release. Preserve failed output, correct the release builder or evidence generator, rerun the disposable fixture and aggregate suite, then use the prior accepted signed tag as appropriate. |

> The parser enforces structure and local artifact metadata only. Workflow attestation and immutable release publication remain separate signed-release gates.
