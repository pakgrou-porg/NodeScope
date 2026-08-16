# Direct TLS Private-Key File Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No TLS certificate, private key, credential, protected host, database, container, or deployment operation was performed. |
| Commit | [`0b530949f52538b80d8cf2134b99c029ba2173ae`](https://github.com/pakgrou-porg/NodeScope/commit/0b530949f52538b80d8cf2134b99c029ba2173ae) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31953647246`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31953647246). |
| Expected result | Native mTLS loading must reject a directory, symlink, or other non-regular private-key object before permission checks or TLS key-pair loading. A direct regular private key must retain the existing POSIX permission guard; Windows compatibility must remain intact. |
| Observed result | Focused agent tests passed for direct regular-file semantics, directory rejection, POSIX symlink rejection, and existing private-key permission behavior. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`http_client.go`](../../../internal/agent/http_client.go); [`http_client_test.go`](../../../internal/agent/http_client_test.go); and the linked GitHub workflow run. |
| Known limitation | The control reduces accidental or configured symlink/non-regular private-key paths but is not a replacement for OS-level key ownership, permission, service-account, PKI issuance, certificate rotation, or protected-environment enrollment controls. No real key or host was exercised. |
| Rollback or recovery | Revert commit `0b53094` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the mTLS key-loading path. |

> Native mTLS loads client private keys only from direct regular files. Key material remains excluded from source control, logs, evidence, and preflight reports.
