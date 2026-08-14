# Windows MSI RTX 5080 and LM Studio Qualification Procedure

**Status:** Windows is **operationally unsupported**. This procedure prepares a controlled validation exercise; it does not authorize production Windows enrollment, unattended service installation, or release promotion.

> A Windows result cannot be promoted until the signed Windows artifact, installer, update/rollback rehearsal, native execution, RTX 5080, LM Studio, and capability-matrix evidence all exist and are reviewed together.

## Authorization and artifact prerequisites

The administrator must approve the MSI host maintenance window, test credential, exact release tag, full commit SHA, artifact checksum, GitHub attestation verification result, test service identity, and rollback owner. Verify the archive checksum and associated release attestation before copying any binary to the MSI host. Do not treat a cross-compiled binary or the current baseline archive as an installer; a signed installer and update/rollback path are separate acceptance gates.

| Prerequisite | Expected result | Stop condition |
| --- | --- | --- |
| Release artifact | Approved signed tag, checksum, SBOM, provenance, and attestation verify against the deployment record. | Any verification failure or missing published release evidence. |
| Capability report | `nodescope-agent.exe -preflight` reports `windows_agent_baseline` and `logical_cpu_count`; unsupported families are explicit unavailable values. | Any GPU/VRAM/NPU/RAM/storage/process/container result is inferred, zero-filled, or silently omitted. |
| Service identity | The least-privilege Windows service identity and file ACL plan are reviewed. | Credential is stored in history, task XML, source control, or an unprotected environment variable. |
| Transport | Internal CA, client identity, revocable credential, configured two-replica ingestion priority, authenticated preflight, and receipt-time behavior pass. | TLS, hostname, credential, or receipt-time behavior differs from the approved evidence. |

## Controlled MSI execution sequence

First run only `-preflight` and one `--once` collection through the approved native agent binary. Capture the full preflight capability output, host Windows version, NVIDIA driver version, RTX 5080 identifier, LM Studio version and model-serving mode, agent version, release evidence result, TLS identity result, first receipt, and latest evidence quality. The first collection must preserve unqualified GPU, VRAM, NPU, memory, storage, process, and container measurements as unavailable.

After baseline transport is accepted, qualify one collector family at a time. For RTX telemetry, record source command or API, driver version, value semantics, unavailable/error behavior, VRAM scope, and comparison against an approved host-local observation. For LM Studio, validate only the approved runtime endpoint and streaming/control-plane surface without retaining prompts or completions. Test ordered primary-to-secondary ingestion failover and failback with receipts. Do not install a scheduled task or unattended service until all preceding evidence and the service identity review are accepted.

## Installer, update, and rollback acceptance

Before production enrollment, the Windows delivery path must provide a signed installer, checksum and attestation verification, atomic staged replacement, retained verified prior binary, update audit evidence, and a tested rollback. The rollback drill must stop the test service, verify the retained prior artifact hash and attestation, restore it, rerun preflight, run one receipt-confirmed collection, and verify that the newer test credential or configuration did not persist unexpectedly.

## Required evidence and recovery

Record source commit SHA, test command/output, MSI environment, expected result, observed result, evidence path, limitation, and rollback result in `docs/operations/evidence/`. If any hardware reading is fabricated, unsupported, missing quality/provenance, or inconsistent with the declared collector boundary, stop. Keep Windows enrollment disabled, revoke the test credential, stop the test service or process, remove the staged artifact, restore only a verified prior artifact, and return the console to explicit unsupported or unavailable state until remediation passes.

## Local readiness command

Run `./scripts/rehearse-windows-baseline-local.sh` before the live exercise. It builds Windows AMD64 and ARM64 baseline agent targets and confirms the explicit unavailable capability contract; it does not execute Windows code, contact the MSI host, create an installer, or qualify RTX 5080 or LM Studio telemetry.
