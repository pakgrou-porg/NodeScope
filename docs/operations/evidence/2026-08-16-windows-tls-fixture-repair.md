# Windows TLS Fixture Repair

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; platform-aware fixture repair and this evidence record are committed together with the TLS material-path hardening. |
| Environment | Local NodeScope checkout and GitHub Windows runner. No CA, client certificate, client key, deployment host, database, or external operational system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(EnablesDockerInventoryOnlyWithExplicitBoolean|RequiresHTTPSInventoryProxyForDockerOptIn|RequiresInternalCAAndClientCertificateForExplicitMTLS)$' -count=1`; `GOOS=windows GOARCH=amd64 go test -c -o /tmp/nodescope-agent-config-windows.test.exe ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions Windows agent runtime test. |
| Expected result | TLS configuration fixtures use platform-native absolute paths, preserving the relative-path rejection while allowing valid native Windows test fixtures. |
| Observed result | Initial CI identified Linux-style fixture paths as non-absolute on Windows. The fixture helper was made platform-aware; focused validation, cross-compilation, and the full local readiness suite passed. |
| Evidence location | [`config_test.go`](../../../internal/agent/config_test.go) and the GitHub Actions run linked from the publication record. |
| Known limitation | Local and CI fixture compatibility does not prove certificate contents, live mTLS connectivity, host qualification, or operational acceptance. |
| Rollback or recovery | Preserve platform-native absolute test fixtures. If a future path check fails on a supported operating system, add a focused native fixture, rerun local and CI validation, and do not weaken the absolute-path production boundary. |

> The production rule remains fail-closed: all configured TLS material paths must be absolute. Only the test fixtures were corrected to represent valid Windows paths.
