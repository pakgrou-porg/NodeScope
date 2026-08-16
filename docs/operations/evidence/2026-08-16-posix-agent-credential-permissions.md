# POSIX Agent Credential-File Permission Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; POSIX permission validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with disposable `0600` and `0644` credential-file fixtures. No real credential, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(RejectsPermissiveCredentialFileOnPOSIX|)$' -count=1`; `GOOS=windows GOARCH=amd64 go test -c -o /tmp/nodescope-agent-credential-permissions-windows.test.exe ./internal/agent`; and `./scripts/release-readiness-check.sh`. |
| Expected result | POSIX credential files with group or world access are rejected before reading; private files remain valid; Windows compilation remains compatible; aggregate readiness passes. |
| Observed result | The permissive fixture was rejected; focused POSIX validation and Windows compilation passed; the full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | Permission-bit validation is enforced only where POSIX mode semantics are available. It does not verify Windows ACLs, secret rotation, host qualification, or operational acceptance. |
| Rollback or recovery | Do not weaken the POSIX permission boundary. Correct the credential file to restrictive owner-only permissions, rerun configuration validation, and retain the prior accepted credential configuration as appropriate. |

> Windows compatibility is preserved because Windows ACL semantics are not equivalent to POSIX group/world mode bits.
