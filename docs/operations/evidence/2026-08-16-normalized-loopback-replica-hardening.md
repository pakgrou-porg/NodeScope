# Normalized Loopback Ingestion Replica Host Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live endpoint configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`27063e7e09226adaa48c0c8c4d9391f918428bd3`](https://github.com/pakgrou-porg/NodeScope/commit/27063e7e09226adaa48c0c8c4d9391f918428bd3) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31955311643`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31955311643). |
| Expected result | Primary and secondary ingestion replica endpoints using normalized loopback labels such as `localhost.` or `localhost...` must require the existing explicit development-only loopback opt-in. Trusted-LAN hostnames and addresses must retain normal validation behavior. |
| Observed result | Focused agent configuration tests passed for normalized primary and secondary loopback labels, existing IP loopback labels, and explicit development-only opt-in behavior. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint configuration only. It does not establish Framework or Asus connectivity, mTLS enrollment, authenticated telemetry, replica failover, certificate rotation, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `27063e7` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Loopback endpoint controls normalize DNS trailing-dot forms before evaluation, so a configuration spelling cannot silently bypass the production network boundary.
