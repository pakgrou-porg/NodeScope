# Inference No-Content-Retention Boundary

NodeScope may measure inference metadata but must never persist, log, trace, audit, cache, or export inference prompt or response content. This rule applies to successful requests, failures, timeouts, malformed streams, backend errors, retries, failovers, panic recovery, and support bundles.

## Permitted metadata

| Category | Permitted fields |
|---|---|
| Request identity | Opaque request ID, timestamp, client key ID, source host ID, route ID, backend ID, model alias. |
| Performance | End-to-end duration, time to first token, inter-token timing, completion duration, throughput, token counts when provided by a trusted runtime, retry count, and outcome code. |
| Operations | Replica ID, backend health state, response status class, normalized error code, byte counts, and safe connection diagnostics. |

## Prohibited fields

The following must never leave the forwarding process as a persistent or diagnostic artifact: prompt text, response text, tool arguments, message arrays, request body bytes, response body bytes, raw SSE frames, authorization headers, cookies, backend error bodies, or unfiltered backend headers.

## Implementation rules

The proxy uses an allowlist-only telemetry model. It constructs a new metadata event from known safe fields rather than serializing an incoming request or upstream response. The proxy discards upstream error bodies after producing a normalized error code. Logs and traces use request IDs, not body fragments. Database schemas contain no content column and audits store only action parameters that have passed an explicit safe-field filter.

Every proxy release must run adversarial canary tests that place distinctive content in prompts, backend response bodies, malformed stream fragments, timeout exceptions, and panic paths. The test suite searches logs, traces, database fixtures, audit output, and support-bundle output for the canary. Any match blocks the release.
