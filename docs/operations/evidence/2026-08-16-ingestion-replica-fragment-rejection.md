# Ingestion Replica Endpoint Fragment Rejection

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; fragment-preserving replica parser change, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No ingestion replica, agent host, deployment environment, database, or external system was contacted. |
| Command | `go test ./internal/agent -run '^TestParseReplicaEndpointRejectsFragments$' -count=1` and `./scripts/release-readiness-check.sh`. |
| Expected result | Primary and secondary ingestion-replica URL fragments are retained through parsing and rejected by the existing credential/query/fragment boundary; aggregate readiness passes. |
| Observed result | The fragment-bearing primary replica fixture was rejected; the full aggregate readiness suite passed. |
| Evidence location | [`config.go`](../../../internal/agent/config.go) and [`config_test.go`](../../../internal/agent/config_test.go). |
| Known limitation | URL parser hardening constrains configuration input only. It does not verify replica identity, mTLS connectivity, failover behavior, host qualification, or operational acceptance. |
| Rollback or recovery | Do not use ingestion-replica URL fragments. Use a direct credential-free HTTPS endpoint, rerun configuration validation, and retain the prior accepted replica configuration as appropriate. |

> The replica parser now preserves fragments so routing metadata cannot be silently normalized into an accepted endpoint.
