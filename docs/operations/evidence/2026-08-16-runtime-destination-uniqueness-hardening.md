# Inference Runtime Destination Uniqueness Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live runtime configuration, host, credential, certificate, database, container, or deployment operation was performed. |
| Commit | [`abe5cbd95ffe7afc4e282fd89bc9fb66e68f3714`](https://github.com/pakgrou-porg/NodeScope/commit/abe5cbd95ffe7afc4e282fd89bc9fb66e68f3714) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31961448083`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31961448083). |
| Expected result | Two inference runtime entries pointing to the same endpoint under different IDs or equivalent URL spellings must fail static validation to prevent duplicate runtime telemetry collection. |
| Observed result | Focused parser tests passed for a leading-zero local port alias and an implicit-versus-explicit default HTTPS port plus trailing-dot hostname alias. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static runtime destination identity only. It does not establish Framework or Asus runtime connectivity, runtime discovery, mTLS enrollment, inference request behavior, replica failover, or deployment health. All remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `abe5cbd` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Runtime IDs remain distinct administrative labels, but they must not represent duplicate network destinations that would collect the same endpoint twice.
