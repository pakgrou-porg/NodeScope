# Approved-Backend Streaming Privacy Validation Procedure

**Purpose:** validate streaming inference behavior through the deployed NodeScope proxy against an administrator-approved vLLM, llama.cpp, or LM Studio backend. This procedure is **not runnable without explicit authorization** because it sends a controlled request to an approved runtime and can require route availability changes.

> The local privacy rehearsal proves that test prompt and response canaries are relayed but never retained in proxy usage, tracing, audit, or support-export metadata. A real backend exercise is still required to establish the same behavior with the deployed streaming implementation.

## Preconditions and safety boundary

The administrator must name the selected approved route, client credential, backend kind, test window, and rollback owner. Use a newly created canary client credential, a pre-approved model, and a non-sensitive deterministic prompt. Do not use personal data, production documents, private model prompts, or a route that is not already approved. Do not log the credential, prompt, completion, backend URL, or internal network address in the evidence record.

| Control | Expected result | Stop condition |
| --- | --- | --- |
| Route approval | Only the named approved runtime route is selected. | Route is discovered, unavailable, or has unreviewed endpoint metadata. |
| Streaming transport | The client receives ordered stream frames and a terminal completion without proxy content transformation. | Frame order, termination, content type, or status is unexpected. |
| Metadata-only observability | Usage, audit, trace, support-export, and error artifacts contain only route/model/client/opaque backend/timing/status metadata. | Prompt, completion, credential, backend URL, private header, or stream frame appears outside the relayed response. |
| Fallback | A controlled retryable backend failure reaches only the approved secondary and does not relay the primary error body. | Fallback contacts an unapproved destination or exposes primary response content. |
| Redirect containment | Backend redirect attempts fail closed. | A redirect target receives prompt, bearer token, or response relay. |

## Live execution and evidence

Run a single streaming request with a canary prompt and an expected short response. Capture client-side timing, first frame time, final frame time, HTTP status, selected opaque backend identifier, route ID, model, client ID, stream flag, and token counts if the backend supplies them. Query or export the proxy usage, audit, trace, and support evidence only through approved administrative paths. Search those redacted artifacts for the exact canary prompt and completion; both must be absent.

Repeat once with the controlled retryable primary condition if the route has an approved secondary. The primary error body must not reach the client or observability artifacts. If a redirect simulation is part of the runtime, verify no target received the request. Record the source commit SHA, environment and backend kind, expected and observed results, evidence locations, known limitation, and recovery action in a new `docs/operations/evidence/` record.

## Recovery and revocation

If any prompt, completion, credential, endpoint, or private header appears in an observable artifact, stop the request, disable the route, revoke the canary client credential, remove the runtime approval, and rotate every credential that could have appeared in the affected path. Preserve only redacted evidence for investigation. Resume only after the unsafe retention path is corrected and both the local rehearsal and the live validation are rerun.

## Local readiness command

Run `./scripts/rehearse-inference-privacy-local.sh` before the live exercise. It tests relay, streaming non-retention, malformed-stream non-retention, retryable fallback, redirect containment, and header allowlisting without contacting an approved runtime.
