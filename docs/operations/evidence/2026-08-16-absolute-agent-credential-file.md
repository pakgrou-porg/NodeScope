# Absolute Agent Credential-File Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; absolute credential-path validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable absolute credential-file fixture. No agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(RejectsRelativeCredentialFile|)$' -count=1` and `./scripts/release-readiness-check.sh` |
| Expected result | An absolute credential-file path remains valid; a relative path fails before any credential read or collection can start; aggregate readiness passes. |
| Observed result | The relative credential-file fixture was rejected; the focused test and full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | Absolute-path validation constrains configuration interpretation only. It does not verify filesystem ownership, file mode, secret rotation, host qualification, or operational acceptance. |
| Rollback or recovery | Do not load credentials from a relative path. Configure a controlled absolute credential-file path through the approved installation process, rerun configuration validation, and retain the prior accepted credential configuration as appropriate. |

> The legacy development-only environment credential exception remains separately explicit; this guard applies whenever a file path is used.
