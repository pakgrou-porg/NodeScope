# Safe Container-Inventory Proxy URL Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; configuration validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No inventory proxy, Docker daemon, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run 'TestLoadConfig(RequiresHTTPSInventoryProxyForDockerOptIn|RejectsUnsafeContainerInventoryProxyURL)$' -count=1` and `./scripts/release-readiness-check.sh` |
| Expected result | A direct HTTPS inventory-proxy URL remains accepted when the explicit mTLS requirements are met; credential-bearing, query-bearing, and fragment-bearing URLs fail configuration validation; aggregate readiness passes. |
| Observed result | The focused configuration test passed for credentials, query, and fragment rejections; the full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | URL validation constrains configuration input only. It does not prove live proxy identity, mTLS connectivity, inventory results, Docker authorization, host qualification, or operational acceptance. |
| Rollback or recovery | Do not embed secrets or routing metadata in an inventory-proxy URL. Move credentials to the paired mTLS files, use a direct HTTPS path, rerun configuration validation, and retain the prior approved proxy configuration as appropriate. |

> The URL parser preserves fragments before validation so fragment-bearing inputs cannot be silently normalized into an accepted configuration.
