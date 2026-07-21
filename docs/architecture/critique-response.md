# NodeScope Proposal Critique Response Matrix

**Status:** Design review response  
**Author:** Manus AI  
**Date:** July 20, 2026

The critique is strong and materially improves the proposal. Of its sixteen substantive findings, **fifteen are accepted**, while the GX10 memory point is **accepted with a factual correction**: current NVIDIA documentation says `nvidia-smi` may report `Memory-Usage: Not Supported` on DGX Spark iGPU platforms, so `memory.used` and `memory.free` cannot be treated as guaranteed fields.[1]

| # | Finding | Disposition | Required correction |
|---:|---|---|---|
| 1 | Supabase Free capacity is optimistic | **Accept** | Recast Free as an evaluation target, not an assumed production fit. Treat paid or self-hosted Supabase as a probable outcome if the measured steady-state design cannot stay below 80% of 500 MB. |
| 2 | 16 KiB zstd/Protobuf estimate is unvalidated | **Accept** | Add an early probe-mode storage-feasibility gate using real Framework and Asus captures under idle, representative, and high-load conditions before full agent packaging or freezing the codec and retention commitment. |
| 3 | GX10 UMA reporting needs explicit conflicting-source handling | **Accept with correction** | Build an UMA-specific panel and alert profile. Attempt NVIDIA memory fields only as optional capabilities; official guidance says framebuffer memory usage may be unsupported. Required values are OS `MemAvailable`, `SwapFree`, huge-page state, CUDA/runtime allocatable memory when available, and per-process GPU memory when exposed, each with source and semantics.[1] [2] |
| 4 | Proxy privacy guarantee lacks an error-sanitization contract | **Accept** | Define an allowlist-only error schema and prohibit backend bodies, prompt/response fragments, raw headers, and wrapped body strings from logs, traces, audits, and database records. Add adversarial failure-path leakage tests. |
| 5 | Backup fenced lease is underspecified | **Accept** | Put the authoritative lease in Supabase, use database time and monotonic fencing tokens, renew on a fixed heartbeat, prohibit finalization without the current token, and do not run or finalize backups while Supabase is unavailable. Peer health informs attempts but never grants ownership. |
| 6 | AgentZero 2.5 is a moving target | **Accept** | Pin support to AgentZero v2.5 commit `d1d48bc9c0e6e253e87c354ce757c518820c6e25`; version the adapter independently; run exact-version contract tests; test newer releases on a compatibility branch before declaring support.[5] [6] |
| 7 | Two magic-link callbacks need an explicit continuity design | **Accept** | Generate links with the Framework callback by default regardless of serving replica. Offer the Asus callback only as an explicit emergency choice. Test link request, token exchange, replica switch, and token refresh across both URLs.[7] [8] [9] |
| 8 | Quantile sketch choice is unspecified | **Accept** | Lock DDSketch at R1.0, with a versioned envelope, 1% default relative accuracy, bounded bins, and a documented zero-value policy. DDSketch is fully mergeable and has relative-error guarantees; the Go implementation is Apache-2.0.[3] [4] |
| 9 | ConnectX presence should be detected in Release 1 | **Accept** | Add detection-only inventory for ConnectX PCI device, driver, firmware, interfaces, and link state. RDMA counters and performance telemetry remain post-Release 1. |
| 10 | Internal CA is under-specified | **Accept** | Make PKI a first-class R1.0/R1.1 deliverable: offline root and issuing intermediate, trust distribution, new-host enrollment, leaf renewal, dual-trust root rotation, recovery, expiry alerts, and rollback tests.[10] [11] [12] |
| 11 | Capacity-governor thresholds are undefined | **Accept** | Define default thresholds: 70% advisory, 80% planning ceiling, 85% protective intervention, 90% raw-write circuit breaker, and 95% emergency state. Thresholds remain Administrator-configurable. |
| 12 | Agent clock skew is not handled | **Accept** | Send wall-clock and monotonic offsets, estimate skew server-side, use a configurable ±30-second default tolerance, quarantine out-of-tolerance samples from historical rollups, retain current health using server receipt time, alert, and report time-sync dependencies in preflight. |
| 13 | Mount learning window has no default | **Accept** | Default to two hours of continuous observation for local/network mounts and six hours for named Docker volumes. Anonymous/transient volumes become expected only when attached to a selected monitored container or explicitly accepted. |
| 14 | Windows artifact language contradicts Release 1 scope | **Accept** | Remove the Windows service package/scaffold from Release 1 artifacts. Keep platform-neutral interfaces; add Windows packaging only in the MSI milestone. |
| 15 | Ingestion rate limiting is omitted | **Accept** | Add per-agent and per-IP token buckets tied to effective interval, compressed/decompressed payload ceilings, sample-count validation, global concurrency and byte budgets, `429 Retry-After`, and observable rejection counters. |
| 16 | Audit integrity during partial failure is unspecified | **Accept** | Use transactional audit intents and state mutations in one Supabase transaction, command IDs and idempotent agent application, durable acknowledgement retries, and audit-result completion. No external control action becomes dispatchable until its audit intent and desired state commit together. |

## Concrete revised defaults

| Control | Revised default |
|---|---|
| Database operating ceiling | **80%** of the Supabase limit, reserving 20% for indexes, MVCC, vacuum, migrations, and bursts. |
| Capacity advisory | **70%**; warn and update the forecast. |
| Protective intervention | **85%**; drop oldest six-hour raw partitions until projected use returns below 80%, subject to configured minimum raw retention. |
| Raw-write circuit breaker | **90%**; stop admitting raw-history chunks while preserving latest state, summaries, configuration, alerts, and transactional audits. |
| Emergency | **95%**; critical alert, reject nonessential telemetry detail, and require Administrator remediation or tier migration. |
| Clock-skew tolerance | **±30 seconds**, configurable. |
| Mount learning | **2 hours** continuous for local/network mounts; **6 hours** for named Docker volumes. |
| DDSketch | **1% relative accuracy**, bounded-bin Apache-2.0 Go implementation, algorithm/version stored in every rollup envelope.[3] |
| Backup lease | Supabase database time; **120-second lease**, renewal every **30 seconds**, monotonically increasing fencing token, randomized secondary acquisition backoff. |
| Certificate warnings | Warning at **30 days**, critical at **14 days**; dual-trust root/intermediate rotation workflow. |
| Ingestion request rate | Dynamic allowance of **2× the expected flush rate**, bounded burst, plus compressed/decompressed byte ceilings and global concurrency control. |

## Storage-feasibility gate

Release 1 must not promise the requested 48-hour raw retention on Supabase Free until real measurements pass. R1.1 uses temporary probe-mode collectors on Framework and Asus—before the full platform agents are packaged—to collect at least 72 hours of representative telemetry, including idle periods, active model loading, sustained inference, container churn, mount changes, and proxy usage. A shorter one-second stress capture measures the upper configuration bound.

The benchmark records p50, p95, and maximum compressed bytes per host-minute; series count; samples per chunk; summary and DDSketch size; index size; dead-tuple behavior; rollup CPU time; query latency; and usage-rollup growth. The envelope records codec, schema version, uncompressed size, compressed size, checksum, sample count, and series count so the format can evolve safely.

The Free-tier design passes only if the projected steady state—including indexes, current state, usage summaries, alerts, audits, and maintenance headroom—remains below 80% of the 500 MB limit. If it does not, the Administrator must choose one of three explicit profiles: shorter raw retention, reduced metric cardinality/sampling, or Supabase Pro/self-hosted Supabase. The proposal should state that a paid or self-hosted tier is **probable** for the full five-host feature set rather than framing it as a remote contingency.

## GX10 UMA presentation contract

NodeScope should display a dedicated **Unified Memory** section rather than a VRAM gauge. It shows OS memory available, swap free, huge-page state, runtime-reported allocatable memory, per-process GPU memory where NVIDIA exposes it, and any optional vendor memory field separately. Every value includes its source and a tooltip describing whether it measures OS reclaimable memory, runtime allocatable memory, or process attribution.

Generic “memory below 10%” alerts do not apply to a single GX10 value. Defaults should use sustained OS memory pressure and swap activity, then annotate runtime allocation failures as separate events. Contradictory values are expected and are shown side by side, not reconciled into a fabricated number.

## Proxy error-sanitization contract

Persistent proxy errors may contain only: request ID, timestamp, client ID, route ID, backend ID, failure-phase enum, transport/error-class enum, HTTP status code, timeout stage, byte counts, latency fields, retry/failover outcome, and `body_dropped=true`. Authorization, cookies, arbitrary request/response headers, URLs with query strings, request bodies, response bodies, and backend-provided error text are prohibited.

Go errors must use typed sentinel/structured errors. Code must not wrap or format response-body bytes into an error. A caller-facing internal error uses a generic problem document and request ID. The test matrix covers rejected requests, backend 4xx/5xx responses, timeouts before headers, stream interruption after partial content, malformed SSE, connection reset, failover exhaustion, panic recovery, and support-bundle export.

## Authoritative backup lease

The lease exists only in Supabase and is acquired in one transaction using database time. Each acquisition increments a fencing token. The holder renews every 30 seconds for a 120-second lease. If Supabase is unavailable, no replica may acquire, renew, or finalize a backup.

A backup writes to a token-specific temporary file. Before atomic rename and manifest publication, the replica must prove it still owns the current token. A stale holder may finish writing its temporary file but cannot publish it; stale files are cleaned later. The peer-health check may delay an acquisition attempt, but it never grants ownership. Secondary takeover also requires the configured backup target to be mounted and writable on both replicas at the same logical path; bootstrap must reject takeover mode otherwise.

## Transactional audit and command protocol

A control mutation begins with one Supabase transaction that inserts the audit intent and the desired command/state using a unique operation ID. If the transaction does not commit, no command is visible to agents and no local cache applies the change. The agent applies each command ID at most once, stores the result locally until acknowledged, and retries the acknowledgement. The server commits the acknowledgement and audit result together. Re-delivery is safe because both dispatch and application are idempotent.

This protocol resolves the partial-failure gap: a durable side effect always has a durable audit intent, and an action result remains pending rather than disappearing when Supabase is temporarily unavailable.

## References

[1]: https://docs.nvidia.com/dgx/dgx-spark/known-issues.html "NVIDIA DGX Spark Known Issues"
[2]: https://nvidia.custhelp.com/app/answers/detail/a_id/5728/~/unexpected-available-memory-reporting-on-dgx-spark "NVIDIA Support — Unexpected Available Memory Reporting on DGX Spark"
[3]: https://github.com/DataDog/sketches-go "DataDog sketches-go — DDSketch Go implementation"
[4]: https://www.vldb.org/pvldb/vol12/p2195-masson.pdf "DDSketch: A Fast and Fully-Mergeable Quantile Sketch"
[5]: https://github.com/agent0ai/agent-zero/releases "Agent Zero Releases"
[6]: https://www.agent-zero.ai/p/docs/extensions/ "Agent Zero Extensions Framework"
[7]: https://supabase.com/docs/guides/auth/redirect-urls "Supabase Auth Redirect URLs"
[8]: https://supabase.com/docs/guides/auth/auth-email-passwordless "Supabase Passwordless Email Logins"
[9]: https://supabase.com/docs/guides/auth/sessions "Supabase User Sessions"
[10]: https://smallstep.com/docs/step-ca/certificate-authority-server-production/ "Smallstep CA Production Considerations"
[11]: https://smallstep.com/docs/step-ca/renewal/ "Smallstep Certificate Renewal"
[12]: https://smallstep.com/docs/step-ca/certificate-authority-core-concepts/ "Smallstep CA Core Concepts"
