# Metadata-Only Proxy Event Allowlist

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; exact event-field regression coverage, privacy-boundary documentation, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. No approved inference backend, streaming client, protected database, deployment host, or identity provider was contacted. |
| Command | `go test ./internal/proxy -run TestMetadataOnlyEventFieldAllowlists -count=1` and `./scripts/release-readiness-check.sh` |
| Expected result | Persisted `UsageEvent` and externally fanned-out `OperationalEvent` retain exactly their metadata-only allowlists; any new field, including a content-bearing field, fails the proxy regression test; aggregate readiness passes. |
| Observed result | The exact field-allowlist test passed; the complete deterministic readiness suite passed. |
| Evidence location | [`metadata_only_event_contract_test.go`](../../../internal/proxy/metadata_only_event_contract_test.go), [`types.go`](../../../internal/proxy/types.go), and [`no-content-retention.md`](../../security/no-content-retention.md). |
| Known limitation | Static field allowlists cannot prove real vLLM, llama.cpp, or LM Studio stream behavior, transport interception, external integration behavior, host qualification, or operational acceptance. Those remain approved-runtime environment gates. |
| Rollback or recovery | Do not add a field to either event as a convenience. Review any required telemetry expansion against the privacy policy, update the explicit allowlist only with approval, rerun local and real-streaming gates, and preserve the previous accepted contract until then. |

> Token counters are metadata. Prompt or response text, body bytes, headers, tool arguments, and credentials remain outside both event schemas.
