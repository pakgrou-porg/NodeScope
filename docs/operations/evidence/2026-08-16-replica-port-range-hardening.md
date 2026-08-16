# Ingestion Replica Port-Range Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`0300d4e98c829f3b0e0be337b95eccfb6a64bf07`](https://github.com/pakgrou-porg/NodeScope/commit/0300d4e98c829f3b0e0be337b95eccfb6a64bf07) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31956217565`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31956217565). |
| Expected result | Primary and secondary ingestion replica endpoints using port numbers above `65535` must fail static configuration validation before a connection attempt. Existing rejection of zero-valued ports and valid trusted-LAN HTTPS endpoints must remain unchanged. |
| Observed result | Focused agent configuration tests passed for out-of-range primary and secondary ports, existing zero-valued port rejection, and valid endpoint configuration. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint syntax only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `0300d4e` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Ingestion replica endpoints must name a valid TCP service port. Invalid port numbers are rejected while configuration is still local and non-mutating.
