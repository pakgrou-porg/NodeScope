# Absolute Agent TLS Material-Path Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; absolute TLS material-path validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No CA, client certificate, client key, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(RejectsRelativeTLSMaterialPaths|)$' -count=1` and `./scripts/release-readiness-check.sh` |
| Expected result | Absolute configured CA, client certificate, and client-key paths remain valid; every relative TLS path fails before mTLS or inventory startup; aggregate readiness passes. |
| Observed result | All three relative TLS material-path fixtures were rejected; the focused test and full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | Absolute-path validation constrains configuration interpretation only. It does not verify certificate contents, filesystem ownership or modes, mTLS connectivity, host qualification, or operational acceptance. |
| Rollback or recovery | Do not configure TLS material with relative paths. Use controlled absolute paths through the approved installation process, rerun configuration validation, and retain the prior accepted TLS configuration as appropriate. |

> This rule applies whether TLS material supports direct agent ingestion or Docker inventory collection.
