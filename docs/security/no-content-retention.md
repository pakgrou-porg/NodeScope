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

## Enforced forwarding behavior

The proxy is intentionally a forwarding boundary rather than a content-inspection service. It temporarily holds request and non-streaming response bytes only as long as required to relay them and extract the three trusted usage counters. The resulting `UsageEvent` is created from an allowlist of route, client, timing, status, and token-count fields; it has no field for text, message arrays, request bytes, response bytes, headers, tool arguments, or credentials.

| Path | Client-visible behavior | Persisted outcome |
|---|---|---|
| Successful non-streaming response | The approved backend response is relayed with content type and no-store caching. | Route, client, timing, status, and trusted token counters only. |
| Successful streaming response | Frames are relayed as they arrive; the proxy records time to first bytes and duration. | Route, client, timing, status, and stream state only. Raw frames are not retained. |
| Backend error response | NodeScope discards the upstream body and returns a generic `502` problem response. | A normalized `backend_error` outcome only. |
| Transport or timeout failure | NodeScope returns a generic `502` problem response without the underlying error text. | A normalized `transport_error` outcome only. |
| Malformed stream read | Bytes already delivered remain on the live connection; no frame is copied into telemetry. | A normalized `stream_error` outcome only. |
| Unexpected panic before headers | NodeScope returns a generic `500` problem response without the panic value. | No panic payload is recorded. |

The proxy forwards only request `Accept`, `Accept-Encoding`, and `User-Agent` headers to an approved backend. On successful responses it forwards only `Content-Type`, then sets `Cache-Control: no-store` itself. Backend cookies, diagnostic headers, authorization material, and arbitrary error headers are never propagated or recorded.

Logging, tracing, audit, and support-export integrations use the shared `MetadataOnlyFanout` adapter. Its only input is `OperationalEvent`, a type with an enforced allowlist of route, client, model alias, opaque backend identifier, status, timing, stream state, token counters, and normalized outcome fields. The adapter type has no content, header, byte, tool-argument, or credential field to serialize.

## Adversarial regression coverage

The local proxy suite uses distinctive canaries in request bodies, successful responses, backend error bodies, backend headers, transport errors, malformed stream fragments, and panic values. Each scenario verifies that the live client response is normalized where required and that the usage recorder contains no canary. The complete release-readiness suite also runs the proxy tests alongside the repository-wide Go, contract, browser, and release policy gates.
