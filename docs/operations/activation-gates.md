# NodeScope Activation-Gate Register

**Status:** controlled pre-activation register. **Owner:** NodeScope Administrator. **Scope:** Framework primary replica, Asus secondary replica, their native agents, and the shared Supabase-backed control plane.

> A passing local build or CI run validates the repository controls listed below; it does **not** authorize live collection, alerting, routing, or operational action against a host. Activation requires the ordered evidence in this register.

## 1. Validated local controls

| Control area | Current repository boundary | Evidence before activation |
| --- | --- | --- |
| Contracts and cross-platform builds | Go, browser, API/protobuf, Linux AMD64/ARM64, and Windows agent checks run in release readiness and CI. | Record the approved commit SHA and its successful CI run. |
| Telemetry quality and ordering | Unavailable values are explicit; experimental evidence cannot drive automatic alerts; future-dated and lower-quality equal-time observations are rejected or retained conservatively. | Retain the approved release-readiness result and host receipt-time report. |
| Inference privacy and routing | No request/response content is retained; approved routes are opaque; backend and agent HTTP clients refuse redirects. | Review route approvals and preflight output without recording credentials or endpoints in tickets. |
| Agent delivery and local tooling | Native clients support TLS 1.3, CA verification, ordered failover, bounded retry, table/JSON/NDJSON output, and metadata-only runtime observation. | Record per-host preflight and authenticated ingestion-preflight receipts. |
| Browser console | Desktop fleet and host-detail views label evidence quality, history absence, host alerts, runtime approval state, and agent clock-offset provenance. | Validate against live read-only telemetry after database and identity gates pass. |

## 2. Ordered live activation gates

| Gate | Required owner | Acceptance evidence | Release consequence if incomplete |
| --- | --- | --- | --- |
| **A. Shared Supabase isolation** | Database owner and NodeScope Administrator | Dedicated `nodescope` schema, least-privilege runtime and migrator roles, RLS review, sibling-schema denial checks, and a disposable noninterference rehearsal. | Do not apply a production migration or provide a broad owner database credential. |
| **B. Migration and service identity** | Database owner | Clean tracked migration input, dedicated-migrator preflight, post-apply isolation verification, and read-only verifier/storage-auditor path. | Do not start server replicas against the shared instance. |
| **C. Internal PKI and credentials** | NodeScope Administrator | Internal CA record, replica leaf certificates, revocable per-agent credentials, trusted CA distribution, and TLS 1.3 hostname verification. | Do not allow HTTP fallback, certificate bypass, or unattended agent execution. |
| **D. Primary and secondary replicas** | Framework and Asus operators | Complete Compose/Portainer deployment on both hosts, health evidence, distinct credential-free endpoints, primary-to-secondary ingest failover, and preferred-endpoint recovery. | Keep collection in one-shot validation mode only. |
| **E. Framework and Asus enrollment** | Host operators | Canonical host and stable agent identity, authenticated preflight, dependency report, selected-process/container allowlists, and one controlled receipt-time collection. | Do not enable persistent agent service collection. |
| **F. Storage and retention qualification** | Storage auditor and Administrator | A completed 72-hour receipt-time benchmark at the approved interval, including completeness, gaps, cardinality, raw/rollup size, and capacity conclusions. | Do not activate full raw-retention policy or scheduled rollups/retention. |
| **G. Platform qualification** | Hardware owner and Administrator | Approved Framework Fedora GPU/NPU version matrix; Asus DGX/UMA and Docker proxy validation. | Keep unqualified metrics experimental, unavailable, or unsupported; never promote them to alert-driving evidence. |
| **H. Backup and recovery** | Backup owner and Administrator | Shared-target validation, primary lease and secondary takeover exercise, restore drill, and archive integrity record. | Do not represent backup protection as operationally ready. |
| **I. Auth, roles, and operations** | Administrator | Invite-only magic-link flow, role-level authorization review, audit verification, and browser/native/MCP operational action approval. | Restrict to controlled administrator validation; do not authorize autonomous operations. |
| **J. Release activation** | Release owner and Administrator | Signed release tag, checksum/SBOM/provenance verification, approved runbook version, and documented go/no-go approval. | Do not promote the tested source revision to a production release. |

## 3. Required deployment record

The Administrator should create one deployment record per promotion attempt. It must include the commit and signed-tag identifiers, CI URL, approved server and agent artifact checksums, certificate identities and expiry dates, redacted endpoint identities, canonical host and agent IDs, preflight/ingestion-preflight receipts, storage-benchmark artifact path, backup-restore record, qualified platform matrix version, approver, and timestamp.

Secrets are deliberately excluded from the deployment record. Agent bearer tokens, database passwords, and private keys remain only in the designated secret store or mode-restricted credential files.

## 4. Stop conditions

Stop the activation attempt and retain only redacted evidence if any gate returns a credential failure, TLS mismatch, redirect, unknown identity, sibling-schema permission, incomplete storage evidence, missing backup target, unqualified accelerator state, or failed restore. Do not work around a stop condition by widening a role, replacing HTTPS with HTTP, bypassing certificate verification, or synthesizing telemetry.

## 5. Related controlled procedures

This register is the deployment order of record. It is used alongside the [Framework and Asus manual installation guide](../agents/manual-install-framework-asus-v2.md), [agent dependency preflight guide](../agents/preflight-dependencies.md), [alert-rule operations guide](alert-rules.md), and [shared-Supabase isolation controls](../database/shared-supabase-isolation.md).
