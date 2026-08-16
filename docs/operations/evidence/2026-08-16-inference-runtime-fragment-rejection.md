# Inference Runtime Endpoint Fragment Rejection

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; fragment-preserving parser change, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No inference runtime, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run '^TestParseInferenceRuntimeEndpointsRejectsFragments$' -count=1` and `./scripts/release-readiness-check.sh`. |
| Expected result | Fragment-bearing inference-runtime endpoint URLs retain their fragment through parsing and fail the existing credential/query/fragment boundary; aggregate readiness passes. |
| Observed result | The fragment-bearing runtime endpoint fixture was rejected; the full aggregate readiness suite passed after retrying a transient protobuf-generator module-proxy EOF. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | URL parser hardening constrains configuration input only. It does not verify runtime identity, endpoint reachability, backend streaming behavior, host qualification, or operational acceptance. |
| Rollback or recovery | Do not use runtime endpoint fragments. Use a direct credential-free HTTP(S) base URL, rerun configuration validation, and retain the prior accepted runtime configuration as appropriate. |

> The inference endpoint parser now preserves fragments so they cannot be silently normalized into an accepted runtime configuration.
