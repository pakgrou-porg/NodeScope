# Query-Bearing Ingestion Replica Endpoint Rejection Coverage

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`ed59aae850462040388d52ca0569835f3de585f5`](https://github.com/pakgrou-porg/NodeScope/commit/ed59aae850462040388d52ca0569835f3de585f5) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31958967650`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31958967650). |
| Expected result | A query-bearing primary or secondary ingestion replica URL must fail static validation because the native sender appends fixed ingestion route paths to the configured base endpoint. |
| Observed result | The added parser-level query-string regression passed against the existing static URL guard. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); [`sender.go`](../../../internal/agent/sender.go); and the linked GitHub workflow run. |
| Known limitation | This control rejects unsupported query-bearing static configuration; it does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `ed59aae` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> The native agent constructs fixed route paths beneath each configured replica base URL. Query-bearing bases are rejected before any request can be constructed.
