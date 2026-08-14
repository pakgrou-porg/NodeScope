# Independent Review Remediation Record

**Scope:** independently checked review findings supplied on 2026-08-14. **Environment:** local source and deterministic test suite. **Production changes:** none.

| Review item | Independent result | Remediation or evidence | Remaining limitation / recovery |
| --- | --- | --- | --- |
| Ingestion rate window does not reset | **Not confirmed.** The current implementation replaces `startedAt` when one minute elapses. | Added `TestResetsPerAgentRateWindowAfterOneMinute` to lock the correct fixed-window behavior. | The policy remains a fixed per-agent window, not a sliding window. Change only with a separate rate-policy design and load test. |
| `Runner.Run` discards collection errors | **Confirmed.** | Runner now reports the error type for every failed collection cycle and warns on collector-specific unavailable evidence; `TestRunnerReportsCollectionCycleFailureWithoutStoppingPeriodicLoop` proves reporting while periodic operation stays resilient. | Logs intentionally omit error strings to avoid endpoint or credential-adjacent disclosure. Operators need the host journal for the error type. |
| Duplicate role models and rank maps | **Confirmed.** | MCP role is now an alias of `auth.Role`; MCP and REST call `auth.Role.Allows` rather than independent rank maps. | External role assignment and live browser-RBAC remain separate environment gates. |
| Control API single-host lookup scans fleet | **Confirmed.** | Shared service now exposes `HostStatus`; REST host route calls the direct service boundary. Postgres filters by `host_slug`; memory implementation retains test behavior. | Database query plan and real host-cardinality performance require environment evidence. |
| Docker probe binary unused | **Not confirmed.** | `deploy/compose/compose.yaml` uses `/app/nodescope-probe` in the server healthcheck. No code removal is appropriate. | Container health execution remains part of the dual-replica live drill. |
| Missing `go vet` in CI | **Confirmed.** | Deterministic local readiness and Go CI now run `go vet ./...`; `scripts/test-ci-quality-contract.sh` prevents removal. | Static analysis does not replace real-host telemetry, browser E2E, or replica drills. |

The review correctly identifies live end-to-end agent-to-server-to-database-to-console, actual browser authentication, real inference streaming, Framework hardware, replica, backup restore, and Windows MSI qualification as remaining **environment gates**. The operational release ledger tracks those boundaries without treating fixture-driven preview data as operational proof.

## Commands and observed results

```text
go test ./internal/ingest ./internal/agent ./internal/mcpserver ./internal/controlapi
go vet ./...
```

The focused remediation test command passed before this record was written. The aggregate release-readiness command remains the final local validation and is required before publication.

## Recovery

All changes are source-only. If a regression appears, revert the remediation commit, restore the prior release tag or checkpoint, and rerun the aggregate readiness suite before re-publishing. Do not bypass the new static-analysis gate.
