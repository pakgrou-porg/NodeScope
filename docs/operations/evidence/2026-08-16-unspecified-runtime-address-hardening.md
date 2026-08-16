# Unspecified Inference Runtime Endpoint Address Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live runtime configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`78b942cff324b2e92c3a671308b87d9db84e5eae`](https://github.com/pakgrou-porg/NodeScope/commit/78b942cff324b2e92c3a671308b87d9db84e5eae) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31959850175`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31959850175). |
| Expected result | Inference runtime endpoint URLs using unspecified IPv4 (`0.0.0.0`) or IPv6 (`::`) addresses must fail static validation before local runtime discovery or telemetry can create requests. |
| Observed result | Focused parser tests passed for both unspecified IPv4 and IPv6 endpoint forms, together with existing runtime endpoint validation coverage. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static runtime endpoint configuration only. It does not establish Framework or Asus runtime connectivity, runtime discovery, mTLS enrollment, inference request behavior, replica failover, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `78b942c` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> An unspecified address is appropriate for binding a listener, not for a NodeScope agent runtime target. The agent now fails closed when it appears in static runtime endpoint configuration.
