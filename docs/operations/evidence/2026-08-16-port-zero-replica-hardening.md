# Port-Zero Ingestion Replica Endpoint Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`ea02b9da88685549420e10131d714970631106ae`](https://github.com/pakgrou-porg/NodeScope/commit/ea02b9da88685549420e10131d714970631106ae) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31954851226`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31954851226). |
| Expected result | Primary and secondary ingestion replica endpoints using `:0` or zero-padded variants such as `:000` must fail during configuration validation. Valid explicit trusted-LAN HTTPS endpoints must remain valid. |
| Observed result | Focused agent configuration tests passed for ordinary and zero-padded port-zero rejection. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint configuration only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `ea02b9d` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> A configured ingestion replica must name a connectable service endpoint. Port zero is never a valid outbound replica destination.
