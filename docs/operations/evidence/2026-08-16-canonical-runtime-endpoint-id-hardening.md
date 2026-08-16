# Canonical Inference Runtime Endpoint ID Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live runtime configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`b9d7619efa45ed994cfbd293ea765169d521b934`](https://github.com/pakgrou-porg/NodeScope/commit/b9d7619efa45ed994cfbd293ea765169d521b934) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31959412670`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31959412670). |
| Expected result | Inference runtime endpoint IDs must reject leading-dot, trailing-dot, and consecutive-dot forms while preserving existing support for letters, numbers, dots, underscores, and hyphens. |
| Observed result | Focused parser tests passed for `.local-vllm`, `local-vllm.`, and `local..vllm`, alongside existing runtime endpoint configuration coverage. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static endpoint IDs only. It does not establish Framework or Asus runtime connectivity, runtime discovery, mTLS enrollment, inference request behavior, replica failover, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `b9d7619` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Runtime endpoint IDs may be used in logs, configuration, and administrative surfaces. Canonical non-path-like IDs prevent ambiguity without restricting the supported local inference runtimes.
