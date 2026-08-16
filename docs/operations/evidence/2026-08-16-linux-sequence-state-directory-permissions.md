# Linux Sequence-State Directory Permission Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; secure state-directory validation, Linux regression coverage, and this evidence record are committed together. |
| Environment | Local Linux NodeScope checkout with a disposable writable state-directory fixture. No agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run '^TestOpenSequenceStoreRejectsGroupOrWorldWritableDirectory$' -count=1` and `./scripts/release-readiness-check.sh`. |
| Expected result | New state directories use restrictive creation mode; existing directories writable by group or others fail before boot-ID read or state-file access; aggregate readiness passes. |
| Observed result | The writable-directory fixture was rejected; the focused Linux validation and full aggregate readiness suite passed. |
| Evidence location | [`state_linux.go`](../../../internal/agent/state_linux.go) and [`state_linux_directory_security_test.go`](../../../internal/agent/state_linux_directory_security_test.go). |
| Known limitation | The directory writability boundary is Linux-specific. It does not verify directory ownership, parent-directory safety, Windows ACLs, host qualification, or operational acceptance. |
| Rollback or recovery | Do not use state directories writable by group or others. Correct the directory mode to an owner-controlled setting, rerun local validation, and retain the prior accepted agent configuration as appropriate. |

> The state store now validates the directory after creation and before reading boot or sequence state, preventing shared writable directories from becoming a persistence boundary.
