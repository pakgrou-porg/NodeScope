# NodeScope Release 1 Gate Plan

This plan converts the working ledger into evidence-based release gates. A code module is not considered complete merely because it compiles; each gate advances only when its required evidence state is achieved.

## Completion states

| State | Meaning | Evidence required |
| --- | --- | --- |
| Designed | Requirements and acceptance criteria are approved. | Decision record and scoped acceptance criteria. |
| Implemented | Source and reviewable tests are present. | Commit, code review, and targeted tests. |
| Locally validated | Disposable or deterministic checks pass without protected infrastructure. | Reproducible local command output. |
| Environment validated | The intended external dependency or real host is exercised. | Redacted host or service evidence with expected and observed results. |
| Operationally accepted | An owner-approved runbook, monitoring, recovery path, and evidence record exist. | Acceptance record including owner, rollback/recovery, and limitation. |

## Release 1 hard gates

| Gate | Current state | Required acceptance evidence | Explicit boundary |
| --- | --- | --- | --- |
| Repository provenance and reproducibility | Locally validated | Clean clone at a named commit runs Go, TypeScript, Vitest, contract, build, SBOM, provenance, checksum, and release-workflow checks. | No protected deployment. |
| Shared-Supabase isolation and migration safety | Fixture validated | Dedicated migrator/runtime/reporting roles; positive and negative sibling-schema, `information_schema`, RLS, upgrade, rollback, and migration-history evidence. | No protected migration until approved. |
| Framework Linux canary | Designed / locally prepared | Authenticated CPU, memory, storage, uptime, network, selected process, and container evidence; quality semantics; disconnect/reconnect/idempotency proof; host qualification matrix. | Framework is the first real host. |
| Asus secondary replica | Designed / locally prepared | Equivalent supported capability matrix plus primary/secondary routing and failover evidence. | Begins after Framework canary acceptance. |
| Replica, PKI, backup, and restore | Locally validated | Primary/secondary outage, failback, certificate rotation/revocation, fenced lease, isolated restore, RPO/RTO, and operator decision-boundary evidence. | No production claim from Compose/source alone. |
| Console authentication and RBAC | Locally prepared | Real magic-link sessions for Viewer, Operator, and Administrator; UI and API denial checks; degraded-replica and session-expiry cases. | No production access claim from fixture-only tests. |
| Seventy-two-hour canary | Designed | Representative fleet/inference load, receipt-time completeness, storage growth, failure-rate, and alert evidence. | Retention settings remain provisional until accepted. |
| Inference privacy and compatibility | Locally validated | Real scoped vLLM, llama.cpp, and LM Studio stream tests for success, timeout, cancel, partial/malformed stream, backend failure, TTFT/token availability, and no-content retention. | No unsupported backend claim. |

## Product tracks

Framework and Asus form the **Linux pilot**. Windows is a separately gated product track and does not block a Linux-only pilot. Windows cannot move beyond Designed until a buildable Windows agent, runtime capability report, signed delivery and rollback path, and real MSI RTX 5080 / LM Studio qualification evidence exist. Every unsupported capability must render an explicit unavailable, unsupported, or permission-denied quality state; it must not be estimated.

## Evidence record requirement

Each operational completion must record the commit SHA, exact test command and result, environment or host, expected and observed result, evidence location, known limitation, recovery path, and approving owner. The project evidence index is the canonical index; this plan supplies the gate semantics.

## Ledger relationship

`todo.md` remains the detailed engineering ledger. This plan controls release claims: a checked source-level item may be Implemented or Locally validated while its corresponding release gate remains blocked on environment validation or operational acceptance.
