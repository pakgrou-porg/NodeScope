# Absolute Agent State-Directory Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; absolute-path validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No agent host, service manager, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(RejectsRelativeStateDirectory|)$' -count=1` and `./scripts/release-readiness-check.sh` |
| Expected result | The default state directory remains absolute; a configured relative state path fails before collection or persistence can start; aggregate readiness passes. |
| Observed result | The relative state-directory fixture was rejected; the focused test and full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | Absolute-path validation constrains configuration interpretation only. It does not verify directory ownership, filesystem permissions, host persistence, agent qualification, or operational acceptance. |
| Rollback or recovery | Do not use a relative state path. Set a controlled absolute state directory, apply host permission hardening through the approved installation process, rerun configuration validation, and retain the prior accepted configuration as appropriate. |

> The default path remains platform-aware; this guard prevents a manually configured relative path from binding retry or sequence state to an arbitrary service working directory.
