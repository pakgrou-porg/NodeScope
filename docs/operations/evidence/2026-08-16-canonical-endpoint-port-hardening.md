# Canonical Endpoint Port Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live replica, runtime, proxy, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`b2c6611b0e8f6bfb9ecfbfad189bf7c154586318`](https://github.com/pakgrou-porg/NodeScope/commit/b2c6611b0e8f6bfb9ecfbfad189bf7c154586318) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31962566226`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31962566226). |
| Expected result | Ingestion replica, inference-runtime, and container inventory proxy URLs with leading-zero ports must fail static validation; valid decimal ports must retain the existing range and zero-port controls. |
| Observed result | Focused configuration tests passed for leading-zero replica, runtime, and inventory-proxy ports, together with existing zero and out-of-range coverage. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static URL port syntax only. It does not establish Framework or Asus connectivity, replica failover, runtime discovery, inventory proxy behavior, mTLS enrollment, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `b2c6611` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Endpoint ports are now represented only in unambiguous decimal form. This removes a second spelling for the same network port before static configuration is accepted.
