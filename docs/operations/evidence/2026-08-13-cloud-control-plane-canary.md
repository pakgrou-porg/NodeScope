# Cloud Control-Plane Canary Record

**Recorded:** 2026-08-13. **Environment:** current cloud sandbox, Linux AMD64. **Harness:** `scripts/e2e-cloud-control-plane-canary.sh`. **Scope:** NodeScope server control-plane behavior only; this is **not** Framework hardware qualification.

> The cloud instance substitutes for the unavailable Framework network route only to validate transport, authentication, persistence, retry, and evidence semantics. It does not produce AMD GPU, XDNA/NPU, storage, Docker, or real Framework-host collector evidence.

## Results

| Control | Expected result | Observed result |
| --- | --- | --- |
| Server transport | Disposable NodeScope server accepts only TLS 1.3 connections with a client certificate issued by the disposable internal CA. | **Passed.** A request without a client certificate failed its TLS handshake. |
| Agent authentication | Valid bearer credential passes the non-mutating ingestion preflight; an invalid bearer credential returns `401`. | **Passed.** |
| Fresh evidence | The authenticated cloud sample persists as `fresh`, source `cloud-canary`, with explicit semantics and receipt time no earlier than observed time. | **Passed.** |
| Idempotent retry | First delivery returns `accepted`; a byte-identical retry returns `duplicate`; exactly one raw-batch or compact receipt remains. | **Passed** through the fail-conservative compact `ingest_receipts` path. |
| Cleanup | Disposable host, agent, server process, certificates, keys, payload, and temporary directory are removed on exit. | **Passed** by the trap cleanup path. |

## Test command and result

```text
export PATH="/home/ubuntu/.local/go1.25.12/bin:$PATH"
./scripts/e2e-cloud-control-plane-canary.sh

Cloud control-plane canary passed: TLS 1.3 mTLS, bearer authentication,
authenticated preflight, rejected invalid credential, idempotent duplicate
delivery, and fresh persisted receipt-time evidence.
```

## Known limitation

The cloud canary uses an authenticated client certificate and telemetry submission fixture, not a Framework-resident native collector. It cannot qualify the Framework’s Fedora release, AMD GPU/NPU matrix, mount/storage behavior, Docker/Portainer observations, vLLM/llama.cpp/LM Studio runtime discovery, endpoint reachability, or 72-hour raw-retention feasibility. Asus remains deferred until Framework hardware qualification is accepted.

## Rollback and recovery

The harness destroys its server, host, agent, certificate, key, payload, and temporary artifacts automatically. If a run fails, do not reuse its identity or credentials. Rerun only after inspecting the redacted phase diagnostic; if a database fixture remains, remove it through the dedicated migrator and verify its absence using a read-only NodeScope-owner query before proceeding.

## Evidence location

The executable procedure is [`scripts/e2e-cloud-control-plane-canary.sh`](../../../scripts/e2e-cloud-control-plane-canary.sh). Its credential-free structural contract is [`scripts/test-e2e-cloud-control-plane-canary-contract.sh`](../../../scripts/test-e2e-cloud-control-plane-canary-contract.sh). Framework qualification remains listed as a distinct deferred gate in the [operational release ledger](../release-epics.md).
