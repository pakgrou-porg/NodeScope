# NodeScope agent preflight dependencies

NodeScope never installs a missing dependency automatically. The native agent reports a capability as **available**, **unavailable**, **unsupported**, or **experimental**, with evidence only. A preflight result is not an authorization to install a package or expand host privileges.

| Capability | Detection | Current operating boundary | Verification |
| --- | --- | --- | --- |
| Framework AMD GPU telemetry | AMD DRM evidence plus `amd-smi` when present | **Experimental** until the NodeScope compatibility matrix qualifies the exact Fedora release, kernel, firmware, ROCm, and AMD SMI versions. Do not run an inferred package-install command from this guide. | Capture the agent preflight capability only; preserve unavailable or experimental evidence. |
| Framework AMD XDNA NPU telemetry | `xrt-smi` when present | **Experimental** until the exact Fedora, XDNA driver, firmware, XRT, and userspace versions are qualified. Do not guess an XRT package from a missing command. | Capture the agent preflight capability only; preserve unavailable or experimental evidence. |
| Asus DGX host GPU telemetry | `nvidia-smi` on `PATH` and device query succeeds | Use only values exposed by the qualified DGX OS tooling. On UMA systems, unavailable dedicated VRAM remains unavailable; never infer it from host memory. | `nvidia-smi --query-gpu=name,temperature.gpu --format=csv,noheader` when the qualified host image supplies it. |
| Docker/Portainer inventory | Explicit opt-in HTTPS inventory proxy and valid paired client certificate/key | The agent never opens `/var/run/docker.sock`, joins the `docker` group, or falls back to a Docker CLI. Use the approved fixed-schema mTLS proxy only. | Run agent preflight, then validate the proxy through the authenticated ingestion and inventory configuration checks. |

The Framework preflight uses accelerator tools opportunistically and marks unqualified AMD readings as experimental. The Asus preflight uses NVIDIA tools opportunistically and preserves UMA provenance. A missing command produces explicit unavailable evidence, never a fabricated value.

Before an accelerator moves beyond experimental status, the designated administrator must approve a written qualified host/version matrix and record the applicable NodeScope release revision.
