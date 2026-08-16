# Legacy Credential Development-Mode Gate

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; explicit development-mode gating, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout with test-only environment values. No real credential, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(AllowsExplicitLegacyEnvironmentCredentialOnlyForDevelopment|RejectsLegacyEnvironmentCredentialOutsideDevelopmentMode)$' -count=1` and `./scripts/release-readiness-check.sh`. |
| Expected result | File-based credentials remain the default; legacy environment credentials require both the existing opt-in and explicit development mode; aggregate readiness passes. |
| Observed result | The explicitly marked development fixture passed; the otherwise identical production-mode fixture was rejected; the full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | This local configuration guard does not verify host environment provenance, secret rotation, systemd unit contents, agent qualification, or operational acceptance. |
| Rollback or recovery | Do not enable a legacy environment credential outside controlled development. Use an absolute private credential file for deployment, rerun configuration validation, and retain the prior accepted secret configuration as appropriate. |

> Production configuration now requires no less than the file-based credential path. Legacy environment credentials require two explicit development-only switches.
