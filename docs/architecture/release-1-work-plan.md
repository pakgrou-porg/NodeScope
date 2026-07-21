# NodeScope Release 1 — Detailed Implementation Work Plan

**Status:** Planning artifact for approval; implementation has not started  
**Author:** Manus AI  
**Date:** July 20, 2026  
**Repository target:** `pakgrou-porg/NodeScope`  
**License target:** Apache License 2.0

> This plan turns the approved architecture proposal into concrete implementation work. It deliberately contains **no calendar estimates**: actual duration depends on access to Framework and Asus, GPU/NPU tooling availability, the real telemetry compression benchmark, and approval-gate decisions. The plan assumes the two-platform Release 1 scope, not the later Susa, MSI, or Pipboy milestones.[1] [2]

## 1. Delivery rules and non-negotiable gates

NodeScope should be built as a sequence of independently testable work packages. A package may be developed only after its prerequisites are satisfied, but documentation and test fixtures should be authored alongside the code rather than deferred to the end. No production credentials, Supabase secrets, CA private keys, or host-specific secrets may enter the public repository.

| Gate | Decision or evidence required | Consequence if not passed |
|---|---|---|
| **G0 — Architecture approval** | Approve the Go + React/TypeScript stack, complete dual-replica Compose deployment, Protobuf + zstd as a benchmark candidate, DDSketch, and offline two-tier PKI. | Do not create the public repository or change Supabase. |
| **G1 — Cloud Free storage feasibility** | Probe-mode captures from Framework and Asus show the selected codec, schema, retention, indexes, and maintenance headroom remain below 80% of the 500 MB Cloud Free limit. | Select reduced retention/cardinality or migrate to Supabase Pro/self-hosted Supabase before committing raw retention. |
| **G2 — Security and privacy readiness** | RLS, invite-only magic links, replica callback continuity, internal CA trust, secret handling, audit protocol, and proxy leakage tests pass. | Do not expose browser, API, MCP, or proxy interfaces outside controlled testing. |
| **G3 — Operational readiness** | Backup restore, primary/secondary takeover, capacity circuit breaker, certificate renewal/rotation, and signed update rollback pass. | Do not declare Release 1 usable for ongoing monitoring. |
| **G4 — Release acceptance** | Sustained dual-host test passes all agreed acceptance criteria. | Publish only a pre-release; continue remediation. |

## 2. Workstream map and critical path

The critical path begins with repository and contract work, proceeds through the Supabase/server foundation and the **storage-feasibility gate**, then reaches the Framework and Asus agents, user interfaces, inference integrations, and operational hardening. Several documentation, test-fixture, and UI tasks can proceed after their APIs are frozen, but they must not bypass a gate.

| Workstream | Primary output | Depends on | Blocks |
|---|---|---|---|
| **WS0 — Approval and project initiation** | Approved decisions, repository creation authorization, secrets-handling plan | None | All implementation work |
| **WS1 — Repository and delivery foundation** | Public repository, CI, release/signing pipeline, code skeleton | WS0 | Builds, tests, deployment artifacts |
| **WS2 — Contracts and data foundation** | Telemetry envelope, OpenAPI, SQL migrations, RLS, test fixtures | WS1 | Server, agents, UI, integrations |
| **WS3 — Replicated server and feasibility gate** | Dual Compose stack, ingestion, probe-mode collectors, capacity decision | WS2 | Full agent scope and production retention |
| **WS4 — Framework native agent** | Fedora native service and AMD GPU/NPU collectors | WS2, WS3 | Framework telemetry acceptance |
| **WS5 — Asus native agent** | ARM64/DGX OS service and UMA/ConnectX inventory | WS2, WS3 | Asus telemetry acceptance |
| **WS6 — Browser, TUI, and CLI** | Fleet console, host detail, admin workflows, terminal client | WS2, WS3, WS4/WS5 interfaces | Human operating experience |
| **WS7 — Proxy, REST, MCP, and AgentZero** | Privacy-safe inference proxy and machine interfaces | WS2, WS3 | Usage reporting and agent access |
| **WS8 — Operations, security, and release hardening** | Alerts, backups, PKI, updates, acceptance suite, release docs | WS3–WS7 | G3 and G4 |

## 3. WS0 — Approval and project initiation

This workstream converts the design into an authorized engineering effort. It is intentionally short but cannot be skipped because the repository will be public and the system will eventually receive credentials and operational authority.

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-000 | Record approval decisions | Create `docs/decisions/000-approved-baseline.md` with approved stack, scope, deferrals, Free-tier policy, and planned gate conditions. | G0 | The decision record matches the approved proposal and has no unresolved architectural ambiguity. |
| NOD-001 | Confirm repository identity | Re-check that `pakgrou-porg/NodeScope` is the intended public slug and create the repository only after explicit approval. | NOD-000 | The public Apache-2.0 repository exists with no secrets or host configuration committed. |
| NOD-002 | Establish contribution and security policy | Add `LICENSE`, `NOTICE`, `README`, `SECURITY.md`, `CONTRIBUTING.md`, code-of-conduct policy if desired, issue templates, and a vulnerability-reporting route. | NOD-001 | A contributor can understand the project’s license, supported reporting path, and local-development rules. |
| NOD-003 | Create secret-handling inventory | Define every secret class: Supabase publishable/service keys, database credentials, agent credentials, client API keys, CA keys, leaf keys, GitHub signing identity, and backup paths. Specify storage, rotation, and forbidden locations. | NOD-000 | Security review confirms no secret type lacks an owner, storage location, or rotation/revocation method. |
| NOD-004 | Define target-host access checklist | Document what must be available on Framework and Asus: SSH/admin access, Docker/Portainer access on server hosts, outbound HTTPS, static IPs, a shared backup target, time synchronization, and permission to install native services/trust anchors. | NOD-000 | The checklist is complete before installation work begins. |

### User actions at WS0

The user approves the architecture decisions, confirms the public repository name, provides no secrets in Git or chat, and identifies the shared backup mount/path that will be writable from **both** Framework and Asus. The user also confirms the administrative email used by the bootstrap wizard.

## 4. WS1 — Repository, build, and release foundation

This workstream creates a reproducible development and release environment. It must produce multi-architecture artifacts from the first usable build, even before all collectors exist, because Framework is AMD64 and Asus is ARM64.[1]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-010 | Initialize monorepo layout | Create the Go module, `cmd/`, `internal/`, `telemetry/`, `api/`, `web/`, `supabase/`, `deploy/`, `docs/`, and `integrations/` directories described in the proposal. | NOD-001 | `go test ./...`, web lint/build, and local documentation rendering all run from a clean clone. |
| NOD-011 | Establish Go and web toolchains | Pin Go, Node, package-manager, formatter, linter, vulnerability scanner, and reproducible build versions. Add a developer bootstrap command. | NOD-010 | A new developer can reproduce the toolchain without manually guessing versions. |
| NOD-012 | Add coding standards and ADR process | Define error handling, context use, logging, API versioning, telemetry-schema evolution, migration rules, and architecture-decision record templates. | NOD-010 | Pull requests have clear review rules and design changes have a recorded decision path. |
| NOD-013 | Build multi-architecture artifacts | Configure GitHub Actions Buildx to publish `linux/amd64` and `linux/arm64` OCI images, plus native Linux AMD64/ARM64 binaries. Do not produce Windows artifacts in Release 1. | NOD-010 | A tagged test release produces correctly labeled artifacts for Framework and Asus. |
| NOD-014 | Add provenance and signing | Generate checksums, SBOMs, GitHub provenance attestations, and keyless Sigstore/cosign signatures. Add verification commands to documentation. | NOD-013 | A clean host rejects an altered binary/image and accepts a genuine release artifact. |
| NOD-015 | Create CI quality gates | Run unit tests, race detector where appropriate, static analysis, dependency/license scan, secret scan, OpenAPI validation, migration tests, telemetry compatibility tests, and container image scan. | NOD-011 | A pull request cannot merge when required checks fail. |
| NOD-016 | Add test fixture strategy | Define vendor-command fixtures, agent sample fixtures, synthetic UMA states, mocked runtime APIs, proxy error fixtures, and fake Supabase test data. | NOD-012 | Tests can run without physical hardware while retaining a controlled hardware-validation layer. |

## 5. WS2 — Contracts, schema, and authorization foundation

This workstream establishes the interfaces that every later component uses. The key design requirement is that agents never receive direct Supabase write credentials; they communicate only with the validated ingestion API.[1]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-020 | Define metric identity model | Specify host, device, device-instance, metric, unit, source, capability, quality, timestamp, and provenance fields. Define how unavailable, estimated, unsupported, and stale values differ. | NOD-010 | Fixtures and API schemas express all required Framework and Asus metric states without ambiguous zero values. |
| NOD-021 | Define telemetry envelope | Create versioned Protobuf schemas with a codec identifier, schema version, compressed/uncompressed size, checksum, sample count, series count, agent/boot/batch identity, wall-clock time, and monotonic timing metadata. | NOD-020 | Compatibility tests prove old envelopes decode after additive schema changes. |
| NOD-022 | Define DDSketch contract | Integrate a Go DDSketch implementation with 1% default relative accuracy, bounded bins, zero-value behavior, merge tests, and algorithm/version metadata in rollup envelopes. | NOD-021 | Independent sketches merge without invalid p95 calculations or undocumented precision changes. |
| NOD-023 | Define OpenAPI and error contract | Design `/api/v1` resource schemas for fleet/host/history, alerts, baselines, roles, credentials, proxy routes, backups, settings, and audits. Specify pagination, filters, error problem documents, and idempotency headers. | NOD-020 | Generated Go/TypeScript clients compile and contract tests cover all expected error forms. |
| NOD-024 | Design Supabase schema and migrations | Create tables for identity/roles, hosts, devices, metric series, latest state, chunks, rollups, inventory, runtime candidates, routes, alerts, audits, leases, backups, and release channels. Add indexes only where query plans justify them. | NOD-020, NOD-021 | Migrations apply and rollback cleanly on a disposable Supabase project. |
| NOD-025 | Implement RLS and database grants | Separate browser-readable data, authenticated-user policies, server roles, invitation/Admin paths, and migrations. Ensure service-role use is restricted to explicit Auth administration paths. | NOD-024 | Policy tests prove Viewer, Operator, Administrator, anonymous user, and server roles see only permitted data. |
| NOD-026 | Implement transactional control protocol | Write the operation-id, audit-intent, desired-state, agent acknowledgement, idempotent redelivery, and final audit-result schema/functions. | NOD-024 | Fault injection proves no command is dispatchable without an audit intent and no completed command lacks a durable result. |
| NOD-027 | Define retention and capacity schemas | Implement raw-chunk partitioning, 1/5/10-minute rollups, capacity snapshots, governor state, and raw-write circuit-breaker state. | NOD-024 | SQL tests verify retention/rollup paths without deleting configuration, summaries, alerts, or audits. |

## 6. WS3 — Replicated server foundation and storage-feasibility gate

The server foundation is deployed as the same Compose/Portainer stack on Framework and Asus. It includes the server, proxy, API, MCP surface, alert evaluator, compactor, and static web application as one deployable unit. The early probe-mode collectors exist only to make the capacity decision before full agent packaging.[1]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-030 | Implement server configuration model | Create strict configuration loading for replica identity, primary/secondary order, Supabase endpoint, secrets paths, LAN addresses, callback URLs, backup target, CA paths, rate limits, and release channel. | NOD-010, NOD-023 | Invalid/missing configuration fails closed with actionable diagnostics and no secret output. |
| NOD-031 | Build the complete server process | Implement HTTP server, health/readiness endpoints, static web delivery, API router, ingestion route, query route, control route, SSE updates, proxy route, MCP route, and scheduled workers under one process. | NOD-023, NOD-024 | The process starts locally with dependencies mocked and reports individual subsystem health. |
| NOD-032 | Create Compose and Portainer deployment | Author `compose.yaml`, Portainer environment templates, persistent volume layout, secrets-file conventions, health checks, restart behavior, and staged deployment instructions. | NOD-030, NOD-031 | The same image and Compose specification start successfully on AMD64 and ARM64 test environments. |
| NOD-033 | Implement agent enrollment and ingestion | Authenticate unique agent credentials, validate envelope version/size/checksum, deduplicate `(agent_id, boot_id, sequence)`, update latest state, and persist chunks/one-minute aggregates. | NOD-021, NOD-024, NOD-031 | Duplicate/replayed batches cannot double-count data; invalid payloads are rejected safely. |
| NOD-034 | Implement ingestion-abuse controls | Add per-agent and per-IP token buckets, burst/concurrency limits, compressed/decompressed body limits, sample-count limits, request validation, `429 Retry-After`, rejection metrics, and backoff guidance. | NOD-033 | Load and malformed-input tests show one agent cannot exhaust replica resources or bypass quotas. |
| NOD-035 | Implement probe-mode collectors | Build narrowly scoped temporary collectors for Framework and Asus that collect representative host/runtime/container data needed for compression and cardinality measurement. | NOD-021, NOD-033 | Both hosts emit capability-tagged data without requiring the final full collector set. |
| NOD-036 | Run 72-hour storage benchmark | Capture idle, loading, sustained inference, container churn, mount changes, and proxy use. Run an additional shorter one-second stress capture. Measure compressed sizes, series count, index growth, rollup cost, query latency, dead tuples, and usage-rollup growth. | NOD-032–NOD-035 | A reproducible report contains p50/p95/max values and an evidence-backed storage projection. |
| NOD-037 | Make G1 retention decision | Compare the projection to the 80% Cloud Free ceiling. Choose approved codec, raw duration, sampling/cardinality policy, or paid/self-hosted migration. Record the decision in an ADR and deployment defaults. | NOD-036 | The retention configuration is explicit and the system cannot silently promise unsupported Free-tier storage. |
| NOD-038 | Implement capacity governor | Add 70/80/85/90/95% states, projections, alerts, raw-retention intervention, raw-write circuit breaker, and recovery criteria. | NOD-027, NOD-037 | Simulated capacity states preserve latest state, summaries, configuration, alerts, and audit data as specified. |

## 7. WS4 — Framework native agent

Framework is the primary host and the native-agent canary. The implementation should prefer stable APIs/libraries but may use structured command fallbacks behind isolated adapters when vendor libraries are unavailable.[1]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-040 | Implement agent service core | Add configuration, credentials, ordered endpoints, bounded queue, heartbeats, desired-state pull, command acknowledgement, failover/failback, telemetry batching, and graceful shutdown. | NOD-021, NOD-033 | The agent survives server loss, switches to Asus, returns to Framework after stabilization, and never opens an inbound port. |
| NOD-041 | Implement Framework preflight | Detect Fedora version/architecture, permissions, systemd, Docker/Portainer access, AMD GPU tools, XDNA/NPU tools, runtimes, time synchronization, mounts, and missing dependencies. Generate DNF commands and verification steps only. | NOD-040 | The report accurately distinguishes available, unavailable, degraded, and permission-blocked collectors. |
| NOD-042 | Implement Linux host collectors | Collect CPU, load, RAM, swap, pressure where available, temperatures, filesystems, network mounts, Docker volumes, mount state, and read-only state. | NOD-040 | Fixtures and live checks verify units, source labels, and missing-value behavior. |
| NOD-043 | Implement AMD GPU collector | Integrate AMD SMI library/API where usable; add structured fallback; collect capabilities, utilization, temperature, memory, process, errors, and provenance. | NOD-041 | The collector handles absent/incompatible AMD tooling without crashing or inventing values. |
| NOD-044 | Implement XDNA/NPU collector | Use `xrt-smi` or supported interface to report readiness, firmware, power mode, active contexts, errors, latency, and throughput only when exposed. | NOD-041 | NPU capability and unsupported fields are differentiated correctly. |
| NOD-045 | Implement service/container/runtime discovery | Inventory all containers, identify candidate vLLM/llama.cpp/LM Studio/AgentZero/Kodex processes, and expose approval candidates without automatically enabling alerts or routes. | NOD-042 | The console/API can distinguish inventory from selected monitored targets. |
| NOD-046 | Implement clock-skew handling | Send wall-clock and monotonic metadata; enforce configurable ±30-second tolerance; use receipt time for current status; quarantine out-of-tolerance historical samples and create alerts. | NOD-040 | Clock-offset test fixtures prove incorrect timestamps cannot pollute rollups. |
| NOD-047 | Package systemd service and updater | Produce systemd unit, unprivileged service account where feasible, protected configuration/credential paths, signed update checker, atomic replacement, health rollback, and Framework canary channel. | NOD-014, NOD-040 | A signed update succeeds; a tampered artifact is rejected; a failed update rolls back. |

## 8. WS5 — Asus Ascent GX10 native agent

Asus must receive Release 1 parity for the agreed operational features, but its GB10 unified-memory environment must not be represented as a conventional dedicated-VRAM system.[1]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-050 | Validate ARM64 service packaging | Build, install, run, upgrade, and roll back the agent on DGX OS/Ubuntu ARM64 using the same release identity as Framework. | NOD-013, NOD-040 | The signed ARM64 artifact functions as a native service without Framework-specific assumptions. |
| NOD-051 | Implement DGX host collectors | Collect Arm CPU, OS RAM, swap, huge-page state, temperatures, filesystem/mount state, selected services, container inventory, and runtime candidates. | NOD-050 | All values include units, source, quality, and unavailable-state behavior. |
| NOD-052 | Implement UMA-specific GPU memory model | Present OS `MemAvailable`, `SwapFree`, huge-page state, CUDA/runtime allocatable memory when available, per-process GPU memory when available, and optional vendor fields side by side. | NOD-051 | No UI/API field claims generic “VRAM free”; contradictory sources are visible and explained. |
| NOD-053 | Implement UMA alert profile | Use sustained OS-memory pressure, swap activity, and runtime allocation failure rather than generic low-VRAM percentage thresholds. | NOD-052 | Synthetic pressure tests trigger the intended alert without false GPU-memory alerts. |
| NOD-054 | Implement ConnectX detection-only inventory | Detect PCI identity, driver, firmware, interfaces, and link state; do not implement RDMA performance counters in Release 1. | NOD-051 | The system reliably reports presence/absence and clearly labels the feature as inventory-only. |
| NOD-055 | Implement Docker/Portainer and vLLM discovery | Discover all containers and candidate model services, preserve administrator approval before monitoring/route activation, and report dependency gaps. | NOD-051 | Container inventory and selected-container alerts behave independently. |
| NOD-056 | Create Asus hardware fixtures and live validation | Capture representative real outputs from supported DGX OS interfaces, including unsupported memory states, and add regression fixtures. | NOD-052–NOD-055 | The collector remains testable when the physical host is unavailable. |

## 9. WS6 — Browser console, TUI, and CLI

The human interface is desktop-first and opinionated: a fleet overview and host-detail workflow rather than a custom dashboard builder. Remote terminal use requires a revocable credential; local use after SSH may use the protected local socket pathway.[1]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-060 | Implement web application shell | Build React routing, authenticated session bootstrap, role-aware navigation, API client, cache policy, error boundaries, accessibility baseline, and desktop layout primitives. | NOD-023, NOD-031 | An invited user can sign in and sees only permitted navigation/actions. |
| NOD-061 | Implement deterministic magic-link flow | Generate Framework callback links by default, support explicit Asus emergency callback, exchange token hash, use relative post-auth routes, prevent self-signup, and test cross-replica session continuity. | NOD-025, NOD-030 | Link requests through both replicas work; a session created through either callback remains usable on both replicas. |
| NOD-062 | Build fleet overview | Show host freshness, alert state, CPU, RAM/UMA, GPU/NPU availability, temperatures, storage, selected services, inference activity, and replica health with clear stale/unavailable rendering. | NOD-033, NOD-040, NOD-050 | The primary human user can diagnose which host/system needs attention from one page. |
| NOD-063 | Build host detail views | Create tabs for hardware, dedicated/UMA memory semantics, storage/mounts, processes, all containers, runtimes, inference, alerts, preflight, and history. | NOD-020, NOD-062 | Framework and Asus views correctly render their platform-specific capabilities. |
| NOD-064 | Build administration workflows | Implement role/invite management, agent/client credentials, intervals, storage baseline diff/refresh, alert rules, discovered runtime approval, routes, certificates, backups, capacity, releases, and audits. | NOD-023, NOD-026 | Viewer/Operator/Admin permissions are enforced server-side and client-side affordances match them. |
| NOD-065 | Build TUI | Provide fleet summary, host drill-down, live refresh, alert list, capacity state, and read-only status using the same API contracts as the browser. | NOD-023, NOD-031 | SSH-local and remote credential modes show consistent data with browser views. |
| NOD-066 | Build scriptable CLI | Provide stable table, JSON, and NDJSON output; filtering; time ranges; host selection; exit codes; credential configuration; and local-socket mode. | NOD-023, NOD-031 | Scripts can obtain machine-readable health/history without parsing terminal tables. |

## 10. WS7 — Inference proxy and agent-facing interfaces

The proxy is the only Release 1 path for consistent client attribution and end-to-end inference timing. It must preserve the absolute content-privacy boundary even when a backend fails unexpectedly.[1] [2]

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-070 | Implement client/API-key management | Create named keys, one-time secret reveal, keyed digest storage, capability sets, revocation, rotation, last-used metadata, and source-host context. | NOD-024, NOD-025 | Revoked/expired keys are rejected on both replicas without exposing key material. |
| NOD-071 | Implement backend registry and approval | Support manual registration plus runtime/Docker discovery candidates. Require Administrator approval before a backend becomes routable. | NOD-045, NOD-055, NOD-064 | Discovery cannot silently expose a new backend to callers. |
| NOD-072 | Implement OpenAI-compatible proxy | Handle request forwarding, streaming, model alias routes, primary backend pinning, optional secondary backend, health checks, failover/failback, and Framework/Asus caller fallback documentation. | NOD-031, NOD-070, NOD-071 | Supported client requests successfully stream through both replicas and route only to approved backends. |
| NOD-073 | Implement privacy-safe telemetry | Measure end-to-end latency, TTFT, prompt/output tokens when reliable, throughput, request outcome, route/backend/model/client/host context, and runtime-native supplements without persisting content. | NOD-072 | Usage rollups reconcile with proxy test requests and never include prompt/response bytes. |
| NOD-074 | Enforce error-sanitization contract | Use typed errors; discard backend bodies/headers; emit allowlisted failure fields only; return generic caller errors with request IDs; test malformed/SSE/timeout/panic/failover paths. | NOD-072 | Canary content is absent from logs, traces, database rows, audits, and support bundles under adversarial failures. |
| NOD-075 | Integrate runtime-specific observation | Scrape/consume vLLM metrics, llama.cpp health/metrics/slots, and LM Studio health/model APIs. Label exact, estimated, unavailable, and runtime-native values correctly. | NOD-071, NOD-073 | Different runtime surfaces produce truthful, provenance-labeled metrics. |
| NOD-076 | Publish REST API and generated clients | Finalize OpenAPI, user/client authentication, role/capability checks, pagination, audit context, idempotency, and usage/history endpoints. | NOD-023, NOD-026 | Browser, CLI, AgentZero tool, and integration tests share the same response contracts. |
| NOD-077 | Implement remote MCP server | Expose curated read, configuration, interval, baseline-refresh, alert, and operational-status tools. Use the same REST policy engine and audit protocol; expose no arbitrary SQL. | NOD-076 | MCP tool calls obey the same authorization and audit rules as browser/API calls. |
| NOD-078 | Implement AgentZero adapter | Build a thin Python tool pinned to AgentZero v2.5 commit `d1d48bc9c0e6e253e87c354ce757c518820c6e25`; add compatibility manifest, exact-image contract tests, and a scheduled newer-version test branch. | NOD-076, NOD-077 | The pinned AgentZero deployment invokes NodeScope without duplicated business logic or authorization drift. |

## 11. WS8 — Operations, security hardening, and Release 1 readiness

This final workstream turns a functioning system into an operable one. It is not “cleanup”: the backup, certificate, capacity, audit, and update tests are release-blocking.

| ID | Task | Detailed work | Depends on | Done when |
|---|---|---|---|---|
| NOD-080 | Implement alert engine and defaults | Add platform/capability-aware default rules, duration, severity, cooldown, suppression, acknowledgement, recovery, and in-console presentation. Include GX10 UMA, agent freshness, storage, temperature, selected services/containers, inference, replica, Supabase, backup, and certificate rules. | NOD-027, NOD-062 | Defaults activate only when relevant capabilities exist and can be safely edited by Administrators. |
| NOD-081 | Implement learned storage baselines | Add two-hour local/network mount and six-hour named-volume default learning periods; handle transient volumes; baseline diff; Operator/Admin refresh; and audit records. | NOD-026, NOD-042, NOD-051 | Expected-mount alerts do not fire during learning but do fire after a validated disappearance/read-only change. |
| NOD-082 | Implement rollup, retention, and governor jobs | Schedule rollups, delete benchmark-approved expired raw partitions, preserve summaries, run capacity checks, and enforce circuit-breaker behavior. | NOD-027, NOD-037, NOD-038 | Repeated worker execution is idempotent and cannot corrupt p95/rollup data. |
| NOD-083 | Implement fenced backup protocol | Use Supabase database time, 120-second lease, 30-second renewal, monotonic fencing token, token-specific temporary files, final ownership recheck, manifest, checksums, ten-copy pruning, and shared target validation. | NOD-024, NOD-032 | A stale primary cannot publish after partition; secondary takeover and disposable restore succeed. |
| NOD-084 | Implement internal PKI lifecycle | Generate offline root/issuing intermediate, distribute trust, issue leaf certificates, renew leaves atomically, monitor 30/14-day expiry, rotate root/intermediate with dual trust, and document recovery/rollback. | NOD-030, NOD-032 | New-host enrollment, leaf renewal, root rotation, and rollback pass without an untrusted-client outage. |
| NOD-085 | Implement replica self-monitoring | Monitor peer reachability/version, APIs, ingestion, proxy, MCP, Supabase, certificate expiry, backup freshness, shared mount, rate-limit rejections, and capacity states. | NOD-031, NOD-083, NOD-084 | A failure in any critical subsystem becomes visible in both the browser and TUI. |
| NOD-086 | Complete agent-update controls | Add signed channel metadata, Framework canary, health-observation period, rollback, pause behavior, and server-image upgrade detection with Administrator approval. | NOD-014, NOD-047, NOD-050 | Signature/digest failures are rejected and the canary cannot silently roll out a failing build. |
| NOD-087 | Create end-to-end failure test suite | Simulate replica loss, backend failure, Supabase loss, capacity threshold transitions, stale lease, backup mount loss, clock skew, expired/replaced certs, duplicate batches, privacy failures, and AgentZero calls. | NOD-080–NOD-086 | Every known critical failure maps to a documented expected behavior and automated test. |
| NOD-088 | Complete operational documentation | Write setup, manual bootstrap, Supabase configuration, Portainer deployment, agent install, CA trust/renewal/rotation, backup/restore, incident response, upgrades, troubleshooting, API/MCP, and Free-to-Pro/self-hosted migration guides. | All prior work | A technically competent operator can install, recover, and upgrade NodeScope without hidden knowledge. |

## 12. Release acceptance procedure

The Release 1 acceptance test should use Framework as preferred and Asus as secondary for a sustained test period. It should not be shortened into a single happy-path demonstration. The objective is to prove that the complete lifecycle works: collection, storage, query, proxying, action auditing, backup, recovery, and updates.

| Test group | Specific verification |
|---|---|
| **Deployment** | Deploy the exact same signed multi-architecture server image through Compose/Portainer on Framework and Asus. Verify each complete replica reports ready and shares the dedicated Supabase project correctly. |
| **Collection** | Run both native agents at the approved default interval. Confirm host/device provenance, metric quality states, deduplication, rollups, clock-skew behavior, and chart gaps during controlled outage. |
| **GX10 correctness** | Verify that UMA sources remain distinct, unsupported NVIDIA memory fields remain unavailable, and no dashboard/API synthesizes dedicated VRAM metrics. |
| **Capacity** | Execute the chosen retention profile. Exercise 70/80/85/90/95% capacity states in a controlled environment and prove that configuration/summaries/audits survive protective behavior. |
| **Identity** | Invite the initial Administrator, test Viewer/Operator/Admin permissions, disable self-signup, test Framework default callback, Asus emergency callback, and cross-replica session continuity. |
| **Proxy privacy** | Send recognizable canary text through normal and adversarial proxy paths; prove it is absent from every persistent/log/audit/support surface. |
| **Agent interfaces** | Compare browser, TUI, CLI JSON/NDJSON, REST, MCP, and pinned AgentZero results for the same user/key/capability context. |
| **Resilience** | Fail agents over from Framework to Asus and back; fail a backend route; temporarily disconnect Supabase; test stale cache behavior and fail-closed mutations. |
| **Backup and PKI** | Run primary backup, simulate stale lease/partition, perform Asus takeover, restore to a disposable project, renew a leaf, and execute dual-trust rotation rehearsal. |
| **Updates** | Install a signed Framework canary update, reject a modified artifact, force a health rollback, and require an Administrator-approved server upgrade. |

## 13. User decisions and inputs required by phase

| When | User input or decision | Why it is required |
|---|---|---|
| Before NOD-001 | Approve G0 and authorize public repository creation. | The repository, CI, release provenance, and public issue tracker should not exist before approval. |
| Before NOD-030 | Supply deployment secrets through approved local secret paths; confirm Supabase project URL and dashboard access. | Bootstrap cannot apply migrations or configure Auth without a secure deployment channel. |
| Before NOD-032 | Confirm Framework/Asus ports, static addresses, and the chosen local names/IP SANs. | Compose and certificate templates need stable endpoints. |
| Before NOD-036 | Provide access windows and representative workloads on Framework and Asus. | The storage gate must reflect actual metrics and model behavior rather than synthetic estimates. |
| At G1 | Approve the measured storage profile, including any shorter raw retention or paid/self-hosted move. | The retention promise and operational cost depend on evidence. |
| Before NOD-083 | Provide a backup target mounted and writable at the same logical path on both server replicas. | Safe secondary backup takeover is impossible without a reachable shared target. |
| Before NOD-084 | Identify devices/browsers that need the internal root certificate and choose protected storage for offline CA material. | Trust distribution and recovery must be planned before exposing TLS endpoints. |
| Before NOD-071/NOD-072 | Identify initial vLLM, llama.cpp, and LM Studio endpoints to register or discover. | The proxy must route only to explicit, approved backends. |
| At G4 | Review test evidence and approve the Release 1 tag. | Release should be based on demonstrated behavior, not feature checklist completion alone. |

## 14. Explicit Release 1 exclusions

The following are deliberately not implementation tasks in this plan. They remain subsequent milestones so they do not dilute the critical path.

| Deferred item | Earliest follow-on milestone |
|---|---|
| Susa platform agent and third replica | Post-Release 1 platform expansion |
| MSI Windows native service, MSI/RTX 5080 collectors, and LM Studio validation | Post-Release 1 MSI milestone |
| Pipboy Linux/AMD multi-GPU agent | Post-Release 1 Pipboy milestone |
| Restart controls for processes, services, or Docker containers | Post-Release 1 action-control milestone |
| Grafana integration | Post-Release 1 read-only visualization option |
| Mobile-optimized browser experience | Post-Release 1 UX milestone |
| RDMA counters and performance telemetry | Post-Release 1 ConnectX/RDMA milestone |
| Custom dashboard builder | Post-Release 1 UX milestone |

## 15. Immediate next action after approval

After **G0**, the first implementation action should be **NOD-001 through NOD-004, followed by NOD-010 through NOD-016**: create the public repository, establish governance/security files, initialize the Go/React monorepo, and configure CI to produce test artifacts. The first technical proof point should then be **NOD-035/NOD-036 probe-mode collection and the storage-feasibility report**, because that evidence determines whether the requested raw-retention policy is achievable on Supabase Cloud Free.[1] [2]

## References

[1]: NodeScope_Release_1_Proposal_v2.md "NodeScope Release 1 Architecture and Technology Proposal"
[2]: NodeScope_Critique_Response.md "NodeScope Proposal Critique Response Matrix"
