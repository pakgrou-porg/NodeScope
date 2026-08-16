# Direct TLS Client-Certificate File Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No TLS certificate, private key, credential, protected host, database, container, or deployment operation was performed. |
| Commit | [`a4c15b1b919ecedca77b224f4277eef883ec98a6`](https://github.com/pakgrou-porg/NodeScope/commit/a4c15b1b919ecedca77b224f4277eef883ec98a6) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31954395844`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31954395844). |
| Expected result | Native mTLS loading must reject a directory, symlink, or other non-regular client-certificate object before private-key validation or TLS key-pair loading. Existing direct regular client-certificate and private-key policy behavior must remain intact; Windows compatibility must remain intact. |
| Observed result | Focused agent tests passed for direct regular-file semantics, directory rejection, POSIX symlink rejection, and the existing permissive-private-key regression path. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`http_client.go`](../../../internal/agent/http_client.go); [`http_client_test.go`](../../../internal/agent/http_client_test.go); and the linked GitHub workflow run. |
| Known limitation | The control reduces accidental or configured symlink/non-regular client-certificate paths but is not a replacement for certificate issuance, rotation, service-account ownership, mTLS enrollment, or protected-environment validation. No real certificate or host was exercised. |
| Rollback or recovery | Revert commit `a4c15b1` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the client-certificate loading path. |

> Native mTLS loads client certificates only from direct regular files. Certificate material remains excluded from source control, logs, evidence, and preflight reports.
