# Inference Runtime Endpoint Port-Range Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live runtime configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`751ea3e6fb8111d3c6d887793e9745f4d41da608`](https://github.com/pakgrou-porg/NodeScope/commit/751ea3e6fb8111d3c6d887793e9745f4d41da608) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31960318631`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31960318631). |
| Expected result | Inference runtime endpoint URLs with port zero or a port above 65535 must fail static validation before local runtime telemetry can create requests. |
| Observed result | Focused parser tests passed for `127.0.0.1:0` and `127.0.0.1:65536`, together with existing runtime endpoint validation coverage. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static runtime endpoint configuration only. It does not establish Framework or Asus runtime connectivity, runtime discovery, mTLS enrollment, inference request behavior, replica failover, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `751ea3e` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> A runtime listener must bind to a valid port, and a NodeScope agent must target one. Invalid port numbers now fail during configuration loading instead of producing a runtime connection error.
