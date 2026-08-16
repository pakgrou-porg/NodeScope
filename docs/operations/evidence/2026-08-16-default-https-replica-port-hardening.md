# Default HTTPS Ingestion Replica Port Identity Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`664099ddfb8a3b0fbeb81a629039fa3fe8486a70`](https://github.com/pakgrou-porg/NodeScope/commit/664099ddfb8a3b0fbeb81a629039fa3fe8486a70) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31957608621`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31957608621). |
| Expected result | Primary and secondary ingestion replicas naming the same HTTPS host with an omitted port and explicit port `443` must fail duplicate detection. Distinct configured paths must retain their supported base-path routing behavior. |
| Observed result | Focused agent configuration tests passed for the new implicit-default-port alias, existing leading-zero port aliases, DNS trailing-dot aliases, path aliases, and valid endpoint configuration. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint identity only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `664099d` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Ordered failover requires distinct replica destinations. The identity check now treats an omitted HTTPS port as the standard HTTPS service port before comparison.
