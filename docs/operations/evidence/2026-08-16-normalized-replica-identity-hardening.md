# Normalized Ingestion Replica Identity Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`bcb33a77270a0e0e7090fa9260d3589f69b770f8`](https://github.com/pakgrou-porg/NodeScope/commit/bcb33a77270a0e0e7090fa9260d3589f69b770f8) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31955779977`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31955779977). |
| Expected result | Primary and secondary ingestion replicas naming the same hostname with ordinary and trailing-dot DNS spellings must fail duplicate detection. Existing canonical handling for scheme case, trailing paths, ports, and IPv6 host syntax must remain stable. |
| Observed result | Focused agent configuration tests passed for the new trailing-dot hostname alias and existing exact and trailing-slash aliases. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint identity only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `bcb33a7` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Ordered failover requires distinct replica destinations. DNS trailing-dot syntax is normalized before comparing configured ingestion endpoint identities.
