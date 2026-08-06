# Windows MSI NodeScope agent baseline

> **Status:** Controlled validation baseline. This guide does **not** authorize unattended production deployment, scheduled-task installation, GPU/NPU collection, or Windows host-management API access.
>
> **Scope:** Windows 11 MSI host with an RTX 5080 or other dedicated GPU, including an LM Studio workload. The binary must originate from an approved NodeScope release or an approved detached source revision.

The current Windows native agent is intentionally narrow. It reuses the NodeScope authenticated sender, endpoint failover order, credential-file support, client-certificate configuration, and sequence envelope path. It reports one positive host fact: the logical CPU count returned by the Go runtime. All other resource families are sent as **explicit unavailable samples**, never as zero, estimated, or inferred values.

## Supported telemetry boundary

| Resource family | Baseline state | Interpretation |
|---|---|---|
| Agent transport and endpoint order | Supported in source | Requires real TLS, credential, and ingestion qualification before production use. |
| Logical CPU count | Supported | Reported by the Go runtime as `cpu.logical_count`. |
| CPU utilization | Unavailable | No Windows performance-counter or WMI collector is enabled. |
| RAM and mounts | Unavailable | No Windows memory or volume API collector is enabled. |
| Temperature | Unavailable | No driver, sensor, or vendor command is invoked. |
| RTX GPU utilization and VRAM | Unavailable | The baseline does not invoke `nvidia-smi`, NVML, WMI, or LM Studio to infer GPU or VRAM values. |
| NPU | Unavailable | The baseline does not invoke Windows AI, vendor, or device APIs. |
| Processes and Docker Desktop inventory | Unavailable | No Windows process scan, Docker socket, named pipe, or privileged helper is used. |

> **Data rule:** An unavailable sample is a deliberate, truthful result. It must remain visibly unavailable in the console and TUI until a Windows collector is hardware-qualified. Do not replace it with `0`, an empty VRAM percentage, or a dashboard-derived approximation.

## 1. Preconditions

An administrator must retain the approved release tag, full commit SHA, artifact checksum, artifact-attestation result, and Sigstore verification result in the deployment record. Use the release verification procedure in [Verified NodeScope releases](../operations/verified-releases.md) when Windows archives are available. Until a signed Windows artifact is published, use a clean detached checkout of the approved source revision only for a controlled validation exercise.

The Windows host needs the following non-secret identifiers and endpoint values supplied through an approved deployment record: `NODESCOPE_AGENT_ID`, `NODESCOPE_HOST_ID`, `NODESCOPE_PRIMARY_ENDPOINT`, and `NODESCOPE_SECONDARY_ENDPOINT`. The agent bearer credential belongs in a protected file; do not place it in PowerShell history, a task XML file, Git, a batch file, or a general-purpose environment variable.

## 2. Build the baseline binary for a controlled validation

Run this only from the approved detached source revision on a trusted build workstation. The command intentionally does not claim that the resulting local build is a release artifact.

```powershell
$ErrorActionPreference = "Stop"
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64" # Use arm64 only on a supported Windows ARM64 host.
go build -trimpath -buildvcs=true -o .\dist\nodescope-agent-windows-amd64.exe .\cmd\nodescope-agent
Get-FileHash .\dist\nodescope-agent-windows-amd64.exe -Algorithm SHA256
```

Copy the approved binary into a protected staging directory such as `C:\ProgramData\NodeScope\staging`. Do not execute from Downloads, a shared profile directory, or a mutable checkout after provenance verification.

## 3. Create protected host directories and credential file

Run the following from an elevated PowerShell prompt. Replace only the credential placeholder through the approved password-manager workflow; do not paste an actual bearer token into a ticket or transcript.

```powershell
$ErrorActionPreference = "Stop"
$root = "C:\ProgramData\NodeScope"
New-Item -ItemType Directory -Force -Path "$root\bin", "$root\state", "$root\credentials" | Out-Null

# Create this file through a protected secret-delivery process.
# Its content is the bearer credential only, with no key name or JSON wrapper.
$credentialFile = "$root\credentials\agent-token"
New-Item -ItemType File -Force -Path $credentialFile | Out-Null
icacls $root /inheritance:r | Out-Null
icacls $root /grant:r "Administrators:(OI)(CI)F" "SYSTEM:(OI)(CI)F" | Out-Null
```

The operator must establish a service identity and precise file ACL strategy before installing any unattended service or scheduled task. That work is intentionally outside this baseline because it needs a Windows-specific hardening review and real-host test.

## 4. Run preflight before any collection

Set only non-secret configuration in the current administrator session. The credential is referenced by file path. If internal TLS is in use, set the CA and client-certificate paths only after their fingerprints and certificate identities are verified through the deployment record.

```powershell
$root = "C:\ProgramData\NodeScope"
$env:NODESCOPE_AGENT_ID = "REPLACE_WITH_STABLE_AGENT_ID"
$env:NODESCOPE_HOST_ID = "REPLACE_WITH_CANONICAL_HOST_ID"
$env:NODESCOPE_AGENT_CREDENTIAL_FILE = "$root\credentials\agent-token"
$env:NODESCOPE_AGENT_STATE_DIRECTORY = "$root\state"
$env:NODESCOPE_PRIMARY_ENDPOINT = "https://REPLACE_WITH_FRAMEWORK_REPLICA/"
$env:NODESCOPE_SECONDARY_ENDPOINT = "https://REPLACE_WITH_ASUS_REPLICA/"

& "$root\bin\nodescope-agent.exe" --preflight
```

The report must contain `windows_agent_baseline` and `logical_cpu_count`. It must show CPU utilization, memory, storage, temperature, GPU/VRAM, NPU, selected process, and container inventory capabilities as unavailable. Abort the exercise if a baseline build reports Linux probe capabilities such as `procfs`, `amd_smi`, `xrt_smi`, `nvidia_smi`, or `docker_socket`.

## 5. Controlled one-shot collection

Only after the endpoint certificate, hostname, agent identity, and revocable credential have been authorized should an administrator run one collection. Preserve the response receipt and the exact preflight output with the deployment record. Do not enable a loop, service, or scheduled task until the primary-to-secondary ingestion failover and receipt-time evidence are qualified on the actual MSI host.

```powershell
& "$root\bin\nodescope-agent.exe" --once
```

A successful command proves only that the agent could construct and transmit the baseline envelope. It does not qualify an RTX 5080, LM Studio, NVIDIA driver tooling, Windows storage, temperature, process, Docker Desktop, or NPU collector.

## 6. Rollback and removal

If preflight, TLS validation, authentication, or ingestion fails, stop the foreground process and retain the failure evidence locally under the approved deployment record. Do not weaken certificate verification, replace the endpoint with HTTP, or switch to an environment credential.

To roll back a controlled binary replacement, restore the prior approved binary from its checksum-verified staging archive, then repeat preflight before another `--once` execution. To remove the baseline validation state entirely, remove the binary and credential file only after the credential has been revoked through the control-plane workflow; never delete the credential first while a future unattended task might still reference it.

## 7. Promotion gate

Windows telemetry remains a validation-only path until all of the following are completed on the actual MSI host: Windows 11 version/driver qualification, RTX 5080 and LM Studio compatibility results, internal TLS and revocable credential proof, authenticated preflight, primary-to-secondary ingestion failover, receipt-time completeness evidence, a least-privilege service identity, a signed Windows release artifact, and a hardware-specific metric compatibility matrix. Until then, the only production-safe presentation is the explicit unsupported or unavailable state documented above.
