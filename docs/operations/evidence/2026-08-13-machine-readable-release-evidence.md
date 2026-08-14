# Machine-Readable Release Evidence Record

**Source commit:** `de45786ae2749fa9748ee86b8c0857e1637259ed`. **Environment:** cloud sandbox, Linux AMD64. **Scope:** deterministic local readiness reporting and tagged-release workflow assembly; no signed tag or GitHub release was created.

> The automation is implemented and locally validated. It must not be interpreted as proof that a release artifact has already been signed, attested, uploaded, or accepted.

## Local evidence result

| Field | Record |
| --- | --- |
| Command | `./scripts/write-release-readiness-report.sh --output <temporary-path>` |
| Expected result | The aggregate readiness suite passes on a clean tree and writes a deterministic JSON report keyed to the committed source revision. |
| Observed result | **Passed.** Report `result` was `passed`, `commit_sha` was `de45786ae2749fa9748ee86b8c0857e1637259ed`, and no generated-contract drift remained. |
| Report contents | Source commit and timestamp, command/expected/observed result, evidence locations, known limitations, and recovery instruction. |
| Recovery | Do not promote a release. Restore the prior accepted signed tag and rerun the report after remediation. |

## Tagged-release workflow boundary

The release workflow now builds Linux and Windows archives, writes and verifies SHA-256 checksum files, generates SPDX SBOMs, creates GitHub Actions provenance/SBOM attestations, assembles `release-evidence.json`, attests that manifest, and only then creates the immutable GitHub Release. The contract test verifies this ordering using disposable artifacts.

## Known limitation

No administrator-approved signed tag has triggered the workflow. Therefore, no GitHub release asset, public attestation, SBOM download, checksum verification against a published archive, or production signer validation has yet been observed. Windows remains operationally unsupported until its separate installer/update/rollback and MSI qualification gates pass.

## Evidence location

Use [`scripts/write-release-readiness-report.sh`](../../../scripts/write-release-readiness-report.sh), [`scripts/assemble-release-evidence.sh`](../../../scripts/assemble-release-evidence.sh), [`scripts/test-release-evidence-contract.sh`](../../../scripts/test-release-evidence-contract.sh), and [`.github/workflows/release.yml`](../../../.github/workflows/release.yml). The dependency state remains in the [operational release ledger](../release-epics.md).
