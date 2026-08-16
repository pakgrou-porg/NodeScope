# Supported Path-Prefix Trailing-Slash Ingestion Replica Identity Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`363560092e9be65b0479b7fcbd1f1db8e808528f`](https://github.com/pakgrou-porg/NodeScope/commit/363560092e9be65b0479b7fcbd1f1db8e808528f) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31958521116`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31958521116). |
| Expected result | Supported base-path ingestion replica endpoints that differ only by one or more trailing slashes must fail duplicate detection because the native sender removes those slashes before appending its ingestion routes. |
| Observed result | Focused agent configuration tests passed for a supported path prefix with and without a trailing slash, existing root path aliases, port aliases, and valid endpoint configuration. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); [`sender.go`](../../../internal/agent/sender.go); and the linked GitHub workflow run. |
| Known limitation | The control normalizes only trailing slash forms already normalized by the native sender. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `3635600` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> NodeScope supports replica base-path prefixes. Duplicate detection now mirrors the sender’s trailing-slash normalization, retaining that supported routing behavior without permitting a false failover pair.
