# NodeScope

**NodeScope** is a production-oriented fleet observability and management console for heterogeneous local compute systems. It combines a polished desktop-first browser console, a standalone SSH-friendly TUI and CLI, native host agents, a dual-replica control plane, and a privacy-preserving inference telemetry proxy.

> **Status:** Release 1 implementation has begun. The public repository currently contains the engineering foundation and design contracts; it is not yet ready to monitor a production fleet.

## Release 1 targets

Release 1 targets two systems: a Framework Desktop with AMD Ryzen AI Max+ 395 and a Framework-hosted primary server replica, plus an ASUS Ascent GX10 running the secondary replica. Both replicas use the same multi-architecture Compose/Portainer deployment and share a Supabase project with TTRPG-OCR through strict `nodescope`-schema, role, credential, migration, and RLS isolation. A controlled Windows MSI baseline can transmit logical CPU-count evidence while keeping unqualified resource families explicitly unavailable; see the [Windows MSI baseline guide](docs/agents/windows-msi-install.md). Full MSI telemetry qualification, plus Susa and Pipboy support, remains future work.

The project tracks CPU, RAM, platform-accurate GPU or unified-memory data, NPU capability, temperatures, storage and mounts, selected process/service health, all-container inventory, runtime health, inference timing and token metrics, alerts, and replica health. The console never invents VRAM values. Every memory reading carries provenance, freshness, and an explicit quality state.

## Core guarantees

| Area | Guarantee |
|---|---|
| **Data truthfulness** | Stale and unavailable data are rendered explicitly. GX10 unified-memory values remain distinct by source and semantics. |
| **Authorization** | Every server-side action uses fleet-wide Viewer, Operator, or Administrator authorization. |
| **Privacy** | Prompt and response content are never persisted, logged, traced, audited, or exported. |
| **Agent safety** | Native agents open no inbound port, use revocable credentials, and report missing dependencies instead of installing them. |
| **High availability** | Framework is the preferred replica; Asus is secondary. Agents and inference callers fail over and later fail back. |
| **Evidence-based retention** | Supabase Cloud Free is an evaluation target. Raw retention is finalized only after a real Framework/Asus capacity benchmark. |

## Planned architecture

```text
Native agents ──HTTPS──> Framework server replica ──┐
                           Asus server replica ──────┼──> Shared Supabase project / nodescope schema only
Browser / TUI / CLI ──HTTPS or local socket─────────┘
Inference clients ──OpenAI-compatible proxy────────> Approved runtime backends
```

The core is written in Go. The browser console is React and TypeScript. Server replicas are published as one multi-architecture OCI image and deployed through Docker Compose or Portainer. Agents and the terminal client are native Linux AMD64/ARM64 binaries. The agent also cross-builds as a deliberately narrow Windows AMD64/ARM64 baseline; it is not a claim of qualified Windows GPU, VRAM, NPU, storage, temperature, process, or container telemetry.

## Development workstreams

The work begins with repository governance, typed contracts, Supabase security, a dual-replica server, and a storage-feasibility gate. It then delivers Framework and Asus agents, the console and terminal interfaces, inference integrations, and operational hardening. The detailed plan is available in [`docs/architecture/`](docs/architecture/) once the repository implementation documents are committed.

## Operations

The [container inventory proxy guide](docs/operations/container-inventory-proxy.md) defines the proxy-only, fixed-schema inventory path. NodeScope agents never mount or query the Docker socket directly.

The [Fedora accelerator qualification boundary](docs/operations/fedora-accelerator-qualification.md) explains why Framework AMD GPU and XDNA NPU readings remain explicitly experimental until an exact tested version matrix exists.

The [native CLI and TUI operations guide](docs/operations/native-console.md) describes metadata-only table, JSON, and NDJSON output through authenticated HTTPS, local SSH relay, or a narrow local verifier role.

The [inference proxy operations guide](docs/operations/inference-proxy.md) records the approved-route, retryable-status failover, metadata-only performance, and no-content-retention boundary for inference callers.

The [capacity governor operations guide](docs/operations/capacity-governor.md) describes deterministic raw-retention admission states and its fail-closed threshold validation.

The [alert rule operations guide](docs/operations/alert-rules.md) describes role-checked fleet and host targeting, evidence-quality constraints, and the production threshold validation boundary.

## Security

NodeScope is a public repository. Do not commit credentials, private keys, telemetry exports, backup data, prompts, responses, or production configuration. Read [SECURITY.md](SECURITY.md) before reporting a vulnerability and [docs/security/secret-handling.md](docs/security/secret-handling.md) before configuring a deployment.

## License

NodeScope is licensed under the [Apache License 2.0](LICENSE).
