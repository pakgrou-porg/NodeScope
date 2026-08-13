# Capacity Governor Operations

NodeScope uses a deterministic capacity governor to decide whether raw telemetry batches remain eligible for retention. The governor does not delete data or schedule maintenance by itself. It produces a bounded decision that the shared capacity-status record can publish to all server replicas.

| Mode | Default quota threshold | Raw batches | Summary rollups | Operator action |
|---|---:|---|---|---|
| `normal` | Below 70% | Accepted | Accepted | Continue normal retention. |
| `constrained` | 70% to below 85% | Accepted | Accepted | Investigate growth and compression. |
| `summary_only` | 85% to below 95% | Rejected | Accepted | Stop raw retention and preserve latest state plus summaries. |
| `protective` | 95% or above | Rejected | Accepted | Take capacity action before resuming raw retention. |

The three threshold values must be finite numbers in strictly increasing order between zero and 100. A missing, `NaN`, or infinite value fails validation and produces no decision. This prevents a malformed policy from silently bypassing `summary_only` or `protective` retention safeguards.

Raw retention also requires an explicit `nodescope.capacity_status` record with `accept_raw_batches=true`. If the record is missing or indicates a protective mode, ingestion remains available for latest state and compact idempotency receipts but does not retain a raw batch. This fail-conservative behavior prevents a deleted or uninitialized circuit-breaker record from silently restoring raw retention.

> The configured thresholds are conservative defaults, not a proof of Supabase capacity. Before production use, validate their behavior against the dedicated NodeScope schema with representative Framework and Asus telemetry and the 72-hour receipt-time storage benchmark.
