# NodeScope agent preflight dependencies

NodeScope never installs a missing dependency automatically. The native agent reports the capability as **available**, **unavailable**, or **unsupported**, along with a copied, platform-specific remediation step and a verification command.

| Capability | Detection | Supported remediation guidance | Verification |
| --- | --- | --- | --- |
| AMD GPU telemetry | `amd-smi` on `PATH` and an AMD DRM device | AMD documents `amdrocm-amdsmi` as the standalone package. On RHEL-family systems, the documented form is `sudo dnf install amdrocm-amdsmi`; select the matching ROCm release and ensure the ROCm binary directory is on `PATH`.[1] | `amd-smi version` |
| AMD XDNA NPU telemetry | `xrt-smi` on `PATH` | AMD documents `xrt-smi examine` as the NPU management interface and supports JSON report output. The installer must be selected for the actual Fedora/XDNA environment rather than guessed by NodeScope.[2] | `xrt-smi examine -f JSON -o /tmp/nodescope-xrt-smi.json` |
| NVIDIA/DGX host GPU telemetry | `nvidia-smi` on `PATH`, device query succeeds | DGX Spark includes its own dashboard. NodeScope uses host-side `nvidia-smi` only for values it exposes; it does not infer dedicated VRAM on unified-memory systems.[3] | `nvidia-smi --query-gpu=name,temperature.gpu --format=csv,noheader` |
| Docker/Portainer inventory | Docker socket is readable and CLI/API responds | Grant the service account read-only Docker access using the local administrator’s approved mechanism. NodeScope does not modify container state in Release 1. | `docker version --format '{{.Server.Version}}'` |

The Framework preflight uses `amd-smi` and `xrt-smi` opportunistically. The Asus preflight uses NVIDIA tools opportunistically and preserves UMA provenance. A missing command produces an explicit unavailable capability, never a fabricated value.

## References

[1] [AMD SMI installation documentation](https://rocm.docs.amd.com/projects/amdsmi/en/latest/install/install.html)

[2] [AMD Ryzen AI `xrt-smi` NPU management interface](https://ryzenai.docs.amd.com/en/latest/xrt_smi.html)

[3] [NVIDIA DGX Spark Dashboard guide](https://docs.nvidia.com/dgx/dgx-spark/dgx-dashboard.html)
