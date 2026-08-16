# Direct Internal-CA Certificate File Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No CA certificate, private key, credential, protected host, database, container, or deployment operation was performed. |
| Commit | [`4b3dde5998bec1a9ef5273f11b9407819426fd29`](https://github.com/pakgrou-porg/NodeScope/commit/4b3dde5998bec1a9ef5273f11b9407819426fd29) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31953992425`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31953992425). |
| Expected result | Native mTLS trust loading must reject a directory, symlink, or other non-regular internal-CA certificate object before reading trust material. A direct regular CA certificate path must retain TLS 1.3 and PEM validation behavior; Windows compatibility must remain intact. |
| Observed result | Focused agent tests passed for direct regular-file semantics, directory rejection, and POSIX symlink rejection. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`http_client.go`](../../../internal/agent/http_client.go); [`http_client_test.go`](../../../internal/agent/http_client_test.go); and the linked GitHub workflow run. |
| Known limitation | The control reduces accidental or configured symlink/non-regular CA paths but is not a replacement for CA issuance, certificate rotation, trust-store ownership, mTLS enrollment, or protected-environment validation. No real CA or host was exercised. |
| Rollback or recovery | Revert commit `4b3dde5` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the internal-CA trust-loading path. |

> Native mTLS loads internal CA trust material only from direct regular files. Certificate material remains excluded from source control, logs, evidence, and preflight reports.
