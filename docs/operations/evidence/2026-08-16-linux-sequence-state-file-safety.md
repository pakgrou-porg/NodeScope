# Linux Sequence-State File Safety

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; direct-regular-file validation, no-follow temporary publication, Linux regression coverage, and this evidence record are committed together. |
| Environment | Local Linux NodeScope checkout with disposable state, temporary-state, and symlink fixtures. No agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'Test(OpenSequenceStoreRejectsSymlinkedSequenceState|SequenceStoreRejectsSymlinkedTemporaryState)$' -count=1`; `GOOS=linux GOARCH=amd64 go test -c -o /tmp/nodescope-agent-state-linux.test ./internal/agent`; and `./scripts/release-readiness-check.sh`. |
| Expected result | Linux sequence-state and temporary files reject symlinks and non-regular files; temporary publication uses a no-follow file descriptor; aggregate readiness passes. |
| Observed result | Both symlink fixtures were rejected; focused Linux validation and the full aggregate readiness suite passed. |
| Evidence location | [`state_linux.go`](../../../internal/agent/state_linux.go) and [`state_linux_security_test.go`](../../../internal/agent/state_linux_security_test.go). |
| Known limitation | State-file safety is currently Linux-specific. It does not verify the state-directory owner, Windows file reparse points or ACLs, host qualification, or operational acceptance. |
| Rollback or recovery | Do not bypass direct-file checks. Remove unsafe symlinked state artifacts, restore a direct regular state file in the controlled state directory, rerun local validation, and retain the prior accepted agent configuration as appropriate. |

> Linux temporary state publication now uses `O_NOFOLLOW` in addition to a direct-regular-file precheck, preventing the temporary state path from following a symlink.
