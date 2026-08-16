# Link-Local Ingestion Replica Address Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live replica, host, credential, certificate, database, container, runtime, or deployment operation was performed. |
| Commit | [`50de64add7fd81cb8218b7ede620e70a75e27a12`](https://github.com/pakgrou-porg/NodeScope/commit/50de64add7fd81cb8218b7ede620e70a75e27a12) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31964680434`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31964680434). |
| Expected result | Primary and secondary HTTPS ingestion replica URLs must reject IPv4 and IPv6 link-local addresses before native-agent delivery can construct a credentialed request. |
| Observed result | Focused configuration testing passed for primary `169.254.169.254` and secondary `fe80::1` targets with the expected explicit link-local rejection. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates literal static endpoint addresses only. It does not resolve hostnames, make network requests, enroll agents, establish mTLS, or qualify Framework or Asus. DNS and live endpoint validation remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `50de64a` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> Credentialed telemetry delivery is now constrained away from non-routable link-local addresses at static configuration load time.
