# Conventional Loopback Alias Ingestion Replica Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`ddf6b27f047c9c3b586c16bace8265453ffa0557`](https://github.com/pakgrou-porg/NodeScope/commit/ddf6b27f047c9c3b586c16bace8265453ffa0557) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31957109545`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31957109545). |
| Expected result | Static conventional loopback labels, including `localhost.localdomain` and `ip6-localhost`, must require the existing explicit development-only replica endpoint opt-in. The check must not resolve names or contact any host. |
| Observed result | Focused agent configuration tests passed for the added primary and secondary alias forms, existing bare and trailing-dot localhost forms, IP loopback values, and explicit development-only opt-in behavior. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control recognizes a fixed conservative alias set and performs no DNS resolution. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `ddf6b27` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Production replica configuration rejects known static loopback aliases without relying on a resolver or performing network I/O. A development-only opt-in remains required for intentional local testing.
