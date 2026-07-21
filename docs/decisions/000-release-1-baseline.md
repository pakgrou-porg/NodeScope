# ADR 000: NodeScope Release 1 Baseline

**Status:** Accepted

## Decision

NodeScope Release 1 is a public Apache-2.0 fleet observability and control plane for Framework and Asus Ascent GX10. The repository uses a Go module for the native agent, server, proxy, TUI, CLI, and operational tooling, plus a React/TypeScript browser console. The browser workspace currently supplies the first console development environment; production server replicas will be distributed as a multi-architecture Docker Compose/Portainer stack for Framework (`10.116.2.145`) and Asus (`10.116.2.56`).

The Supabase project in `us-east-2` is the shared durable system of record. Cloud Free is an evaluation target only. The raw-retention policy is not final until a real Framework and Asus storage-feasibility benchmark demonstrates that the selected envelope, indexes, rollups, audit records, and maintenance headroom fit below the defined capacity ceiling.

## Product invariants

| Invariant | Requirement |
|---|---|
| Memory truthfulness | NodeScope never synthesizes VRAM values. Every memory value includes a source, semantics, freshness, and quality state. GX10 receives a dedicated UMA view. |
| Data availability | Stale and unavailable states are explicitly rendered; unavailable values are never silently represented as zero. |
| Authorization | Every protected server-side operation is evaluated against fleet-wide Viewer, Operator, or Administrator permissions. Client-side rendering is only an affordance; the server is authoritative. |
| Privacy | Prompt and response content must never be persisted, logged, traced, audited, or included in support output. Proxy telemetry is metadata-only. |
| Discovery | Process, runtime, route, and alert candidates require explicit approval before they receive operational significance. |
| Agents | Native agents open no inbound port, use ordered ingestion endpoints, fail over and fail back, and report missing dependencies without installing them. |
| Server replicas | Framework is preferred and Asus is secondary. Both run the complete server stack and remain stateless apart from bounded caches and local operational files. |
| Initial actions | Release 1 permits interval changes and storage-baseline refresh. Process, service, and container restart actions are deferred. |

## Explicit Release 1 scope

The browser console provides a polished fleet overview, platform-aware host detail views, administrator workflows, in-console alerts, and a desktop-first experience. The terminal experience consists of a standalone interactive TUI and scriptable CLI. Machine consumers receive consistent REST, remote HTTPS MCP, and exact-version AgentZero integration surfaces.

Release 1 supports Framework and Asus. Susa, MSI, Pipboy, restart controls, Grafana integration, mobile optimization, Windows service packaging, and RDMA performance telemetry are post-release work.

## Consequences

The project must maintain two independently deployable artifacts: a multi-architecture Docker image for server replicas and native Linux AMD64/ARM64 binaries for the agents and terminal client. The public repository must contain no deployment secrets, client API keys, database passwords, CA private keys, or host-specific confidential configuration.

The work order begins with contracts and the storage benchmark rather than a fabricated dashboard data model. User-facing screens may use fixture data during development, but fixture data must clearly indicate its development-only provenance and may not be presented as customer telemetry.
