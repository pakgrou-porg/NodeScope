# Container Inventory Proxy Target Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live proxy, container host, credential, certificate, database, or deployment operation was performed. |
| Commit | [`eb84cb474ae2fdc0fb103697deaed17597ebb814`](https://github.com/pakgrou-porg/NodeScope/commit/eb84cb474ae2fdc0fb103697deaed17597ebb814) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31960826992`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31960826992). |
| Expected result | Container inventory proxy URLs using unspecified IPv4 or IPv6 addresses, port zero, or a port above 65535 must fail static validation before optional inventory requests can be constructed. |
| Observed result | Focused configuration tests passed for unspecified IPv4, unspecified IPv6, port zero, and high-port proxy target forms, alongside existing inventory opt-in and URL-safety tests. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static proxy target configuration only. It does not establish the narrow read-only inventory proxy, Framework or Asus Docker access, mTLS enrollment, inventory evidence quality, replica failover, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `eb84cb4` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> The optional container inventory feature accepts only a valid, connectable HTTPS proxy target; wildcard listener addresses and invalid ports are rejected during configuration loading.
