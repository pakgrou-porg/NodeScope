# Normalized Numeric Ingestion Replica Port Identity Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`178a8187d2a870886f23171f105685ab958d06c8`](https://github.com/pakgrou-porg/NodeScope/commit/178a8187d2a870886f23171f105685ab958d06c8) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31956656837`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31956656837). |
| Expected result | Primary and secondary ingestion replicas naming the same host and numeric port with ordinary and leading-zero port spellings must fail duplicate detection. Existing DNS hostname, path, and valid-port normalization must remain stable. |
| Observed result | Focused agent configuration tests passed for the new leading-zero port alias, exact duplicates, trailing-slash aliases, trailing-dot DNS aliases, and valid endpoint configuration. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint identity only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `178a818` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Ordered failover requires distinct replica destinations. Numeric port spellings are normalized before comparing configured ingestion endpoint identities.
