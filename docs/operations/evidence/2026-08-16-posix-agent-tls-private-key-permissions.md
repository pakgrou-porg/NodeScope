# POSIX Agent TLS Private-Key Permission Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; runtime private-key permission validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with a disposable permissive private-key fixture. No real certificate, private key, endpoint, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run '^TestNewMTLSHTTPClientRejectsPermissivePrivateKeyOnPOSIX$' -count=1`; `GOOS=windows GOARCH=amd64 go test -c -o /tmp/nodescope-agent-tls-private-key-windows.test.exe ./internal/agent`; and `./scripts/release-readiness-check.sh`. |
| Expected result | POSIX client private keys with group or world access are rejected before TLS key-pair loading; Windows compilation remains compatible; aggregate readiness passes. |
| Observed result | The permissive private-key fixture was rejected; focused POSIX validation and Windows compilation passed; the full aggregate readiness suite passed. |
| Evidence location | [`http_client.go`](../../../internal/agent/http_client.go) and [`http_client_test.go`](../../../internal/agent/http_client_test.go). |
| Known limitation | Permission-bit validation is enforced only where POSIX mode semantics are available. It does not validate certificate content, Windows ACLs, live mTLS connectivity, host qualification, or operational acceptance. |
| Rollback or recovery | Do not weaken the private-key permission boundary. Correct the private key to restrictive owner-only permissions, rerun client construction validation, and retain the prior accepted TLS configuration as appropriate. |

> The permission check occurs before key-pair parsing so an unsafe private key is never passed to the TLS loader.
