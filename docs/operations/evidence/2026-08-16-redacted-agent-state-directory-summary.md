# Redacted Agent State-Directory Summary

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; summary redaction, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No agent host, service manager, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run '^TestLoadConfig$' -count=1` and `./scripts/release-readiness-check.sh` |
| Expected result | Redacted configuration output exposes only whether the agent state directory is configured and never serializes its filesystem location; aggregate readiness passes. |
| Observed result | The focused summary regression and complete aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | Summary redaction prevents configuration-summary path disclosure only. It does not change agent filesystem behavior, permission enforcement, host qualification, or operational acceptance. |
| Rollback or recovery | Do not reintroduce raw filesystem locations into redacted output. If detailed troubleshooting requires a path, obtain it directly through an authorized host session and keep it out of generalized diagnostic summaries. |

> The summary now emits `state_directory_configured` rather than a raw local path.
