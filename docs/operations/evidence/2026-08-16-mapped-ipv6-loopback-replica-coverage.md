# IPv4-Mapped IPv6 Loopback Ingestion Replica Coverage

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`90d210e1e6b03668c7cff31591015563605796de`](https://github.com/pakgrou-porg/NodeScope/commit/90d210e1e6b03668c7cff31591015563605796de) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31958075039`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31958075039). |
| Expected result | A static IPv4-mapped IPv6 loopback ingestion endpoint (`::ffff:127.0.0.1`) must be rejected outside the existing explicit development-only loopback opt-in. The check must remain static and must not resolve names or contact an endpoint. |
| Observed result | The added primary-endpoint regression passed using the existing `net.IP.IsLoopback` behavior, confirming no implementation change was required. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config_test.go`](../../../internal/agent/config_test.go); [`config.go`](../../../internal/agent/config.go); and the linked GitHub workflow run. |
| Known limitation | The control covers static IP parsing only and intentionally performs no DNS resolution. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `90d210e` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> The existing static loopback detector already recognizes IPv4-mapped IPv6 loopback addresses. This regression prevents that behavior from being weakened without an explicit test failure.
