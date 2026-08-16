# Unspecified Ingestion Replica Address Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, or container was changed. |
| Commit | [`581920e3698ca9eafcd1373d022277ad7a8354d8`](https://github.com/pakgrou-porg/NodeScope/commit/581920e3698ca9eafcd1373d022277ad7a8354d8) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31952929729`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31952929729). |
| Expected result | Primary and secondary ingestion replica endpoints with `0.0.0.0` or `[::]` must fail during configuration validation. Explicit configured HTTPS LAN replica endpoints must remain valid. |
| Observed result | Focused agent configuration tests passed for unspecified IPv4 and IPv6 endpoint rejection. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint configuration only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `581920e` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> A replica endpoint must name a connectable host. Wildcard listener addresses are never valid outbound ingestion targets.
