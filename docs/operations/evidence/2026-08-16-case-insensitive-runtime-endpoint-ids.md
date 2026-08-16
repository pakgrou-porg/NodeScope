# Case-Insensitive Runtime Endpoint Identity Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; canonical endpoint-ID validation, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No inference runtime, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run '^TestParseInferenceRuntimeEndpointsRejectsCaseInsensitiveDuplicateIDs$' -count=1` and `./scripts/release-readiness-check.sh`. |
| Expected result | Exact duplicate endpoint IDs and IDs that differ only by letter case fail before runtime routing or discovery configuration can start; aggregate readiness passes. |
| Observed result | The `Local-VLLM` and `local-vllm` fixture was rejected; the full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | Canonical endpoint-ID validation constrains configuration identity only. It does not verify runtime identity, endpoint reachability, backend streaming behavior, host qualification, or operational acceptance. |
| Rollback or recovery | Do not configure case-conflicting endpoint IDs. Select one unique canonical ID, rerun configuration validation, and retain the prior accepted runtime configuration as appropriate. |

> Endpoint ID validation now canonicalizes case for uniqueness while preserving the operator-supplied ID for display.
