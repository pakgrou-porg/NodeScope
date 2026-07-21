# NodeScope

**NodeScope** is a production-oriented fleet observability and management console for heterogeneous local compute systems. It combines a polished desktop-first browser console, a standalone SSH-friendly TUI and CLI, native host agents, a dual-replica control plane, and a privacy-preserving inference telemetry proxy.

> **Status:** Release 1 implementation has begun. The public repository currently contains the engineering foundation and design contracts; it is not yet ready to monitor a production fleet.

## Release 1 targets

Release 1 targets two systems: a Framework Desktop with AMD Ryzen AI Max+ 395 and a Framework-hosted primary server replica, plus an ASUS Ascent GX10 running the secondary replica. Both replicas share a dedicated Supabase project and use the same multi-architecture Compose/Portainer deployment. Later support will cover Susa, MSI, and Pipboy.

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
                           Asus server replica ──────┼──> Dedicated Supabase project
Browser / TUI / CLI ──HTTPS or local socket─────────┘
Inference clients ──OpenAI-compatible proxy────────> Approved runtime backends
```

The core is written in Go. The browser console is React and TypeScript. Server replicas are published as one multi-architecture OCI image and deployed through Docker Compose or Portainer. Agents and the terminal client are native Linux AMD64/ARM64 binaries.

## Development workstreams

The work begins with repository governance, typed contracts, Supabase security, a dual-replica server, and a storage-feasibility gate. It then delivers Framework and Asus agents, the console and terminal interfaces, inference integrations, and operational hardening. The detailed plan is available in [`docs/architecture/`](docs/architecture/) once the repository implementation documents are committed.

## Security

NodeScope is a public repository. Do not commit credentials, private keys, telemetry exports, backup data, prompts, responses, or production configuration. Read [SECURITY.md](SECURITY.md) before reporting a vulnerability and [docs/security/secret-handling.md](docs/security/secret-handling.md) before configuring a deployment.

## License

NodeScope is licensed under the [Apache License 2.0](LICENSE).
