# Inference Proxy Operations

NodeScope provides an OpenAI-compatible proxy for **administrator-approved** model routes. It accepts `POST` requests, authenticates the calling client, resolves the requested model against the approved registry, and forwards only to the selected backend. Client credentials and NodeScope control headers are not forwarded to a runtime backend.

## Runtime approval endpoint boundary

Administrator approval accepts only a credential-free HTTPS base URL or `/v1` endpoint. Plain HTTP is accepted only for `localhost`, `127.0.0.1`, or `::1`. Query strings, fragments, embedded user information, and arbitrary endpoint paths are rejected. The approval API creates an opaque candidate ID and records only the runtime kind plus transport class (`https` or `loopback_http`) in its audit event. It does not place the endpoint location in the candidate ID or audit metadata.

This validation improves the approval boundary but does not authorize a runtime automatically. Operators must still validate actual client access, server identity, route configuration, and health evidence on the intended LAN host before making a backend routable.

## Failover policy

Each route has a preferred backend and may have one ordered secondary backend. The proxy attempts the secondary only when the preferred backend cannot be reached or returns an HTTP `502`, `503`, or `504`. It does **not** retry generic backend application failures such as HTTP `500`, because retrying a request whose server-side execution state is unknown could duplicate work. Backend redirects are never followed: the proxy returns a normalized gateway failure and records only the redirect status metadata, keeping an inference request body within the explicitly approved backend destination.

When fallback succeeds, NodeScope returns the secondary response and records the opaque route backend identity as `<route-id>:secondary`. If fallback transport fails after a retryable primary response, NodeScope retains only the primary status metadata, discards both backend response bodies, and returns the normalized proxy failure. The client never receives an upstream diagnostic body.

## Performance and privacy event

The proxy records a bounded usage event containing the route and model alias, caller identifier, opaque backend identity, HTTP outcome, streaming flag, duration, time to first byte/token, and any provider-supplied OpenAI usage counters from a bounded **non-streaming** response. The event has no fields for prompt text, completion text, request/response bodies, headers, credentials, tool arguments, or endpoint locations.

| Response mode | Available evidence | Deliberately unavailable evidence |
|---|---|---|
| Non-streaming response with OpenAI `usage` | Prompt tokens, output tokens, total tokens, duration, and first-byte/token time | Prompt and response content; headers; backend URL |
| Streaming response | Streaming state, duration, first-byte/token time, normalized outcome | Token counts unless a future privacy-reviewed streaming usage contract is introduced; all stream-frame content |
| Backend or transport failure | Opaque backend identity, status/outcome, duration | Upstream error body and diagnostic text |

> A successful local test is not a runtime qualification. Administrators must verify actual vLLM, llama.cpp, and LM Studio routes, credentials, TLS, failover, and client behavior on the intended LAN hosts before enabling production traffic.

## Operator checks

Use a non-sensitive client request against an approved route, then confirm the route and backend identity from the metadata-only usage view. For a controlled fallback rehearsal, temporarily make the preferred backend return a retryable gateway status and confirm that the request is served by the secondary. Restore the preferred route promptly; do not place prompts, response text, credentials, or backend URLs in incident tickets, command history, or support exports.
