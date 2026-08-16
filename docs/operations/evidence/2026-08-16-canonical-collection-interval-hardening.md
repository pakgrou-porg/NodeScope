# Canonical Collection Interval Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No host, credential, certificate, replica, database, container, runtime, or deployment operation was performed. |
| Commit | [`a73e487d724b213cba7cf2d3c76728af022b3239`](https://github.com/pakgrou-porg/NodeScope/commit/a73e487d724b213cba7cf2d3c76728af022b3239) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31964192345`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31964192345). |
| Expected result | `NODESCOPE_COLLECTION_INTERVAL_SECONDS` must reject ambiguous leading-zero values while retaining the existing inclusive one-to-sixty-second policy range. |
| Observed result | Focused configuration testing passed for a leading-zero interval with the expected canonical-decimal rejection. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static local configuration only. It does not change a deployed collection interval, collect a live sample, enroll an agent, contact a replica, or qualify Framework or Asus hardware. Those remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `a73e487` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> The sampling interval is now unambiguous in configuration: values retain the documented one-to-sixty-second range and must use canonical decimal notation.
