# Framework and Asus agent installation

NodeScope Release 1 uses the same native Linux agent on the Framework x86-64 host and the Asus Ascent GX10 ARM64 host. The agent sends bounded Protocol Buffers plus Zstandard telemetry to the preferred Framework replica and falls back to Asus only on transport or server failures. It returns to Framework automatically on the next successful collection attempt.

> **Prerequisite:** Deploy the two server replicas and internal CA before starting either agent. Agent endpoints must use HTTPS, and the local CA certificate must be installed at the path configured in `/etc/nodescope-agent/agent.env`.

| Host | Target architecture | Build command |
| --- | --- | --- |
| Framework | `linux/amd64` | `GOOS=linux GOARCH=amd64 go build -o dist/nodescope-agent-linux-amd64 ./cmd/nodescope-agent` |
| Asus GX10 | `linux/arm64` | `GOOS=linux GOARCH=arm64 go build -o dist/nodescope-agent-linux-arm64 ./cmd/nodescope-agent` |

The installer performs no automatic package installation. It creates the unprivileged `nodescope` service account, protects the state directory, creates a root-readable environment template, and enables the service without starting it until credentials are present.

```bash
sudo ./deploy/agent/install-linux.sh ./dist/nodescope-agent-linux-amd64
sudo /usr/local/bin/nodescope-agent --preflight | jq .
```

The preflight report is the authoritative source for available collectors. If `amd-smi` is absent on Framework, use the matching ROCm package guidance; AMD documents `sudo dnf install amdrocm-amdsmi` for RHEL-family distributions after ROCm prerequisites are complete.[1] If `xrt-smi` is absent, install the Ryzen AI/XRT userspace package matched to the actual Fedora kernel, XDNA driver, and firmware; do not guess an incompatible package.[2] On Asus, use the NVIDIA tooling supplied by DGX OS and treat unavailable framebuffer memory as explicitly unavailable rather than VRAM inferred from host memory.[3]

Populate `/etc/nodescope-agent/agent.env` only after secure agent enrollment supplies the host UUID, agent UUID, random credential, and local CA certificate. The exact collection interval defaults to five seconds and may be set from one to sixty seconds. The process and alert-container allowlists are optional and comma-separated.

```dotenv
NODESCOPE_AGENT_ID=<agent UUID>
NODESCOPE_HOST_ID=<host UUID>
NODESCOPE_AGENT_CREDENTIAL=<one-time generated secret>
NODESCOPE_PRIMARY_ENDPOINT=https://10.116.2.145:8443
NODESCOPE_SECONDARY_ENDPOINT=https://10.116.2.56:8443
NODESCOPE_CA_CERT_PATH=/etc/nodescope-agent/ca.pem
NODESCOPE_COLLECTION_INTERVAL_SECONDS=5
NODESCOPE_SELECTED_PROCESS_NAMES=LMStudio,vllm
NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES=vllm,agent-zero
```

After credentials and the CA are present, start the service and inspect only redacted status output.

```bash
sudo systemctl start nodescope-agent
sudo systemctl status nodescope-agent --no-pager
sudo journalctl -u nodescope-agent -n 50 --no-pager
```

The storage feasibility gate runs after both agents have collected representative workloads for 72 hours. Export each host’s actual compact probe summary and run `nodescope-storage-estimate`; do not approve the two-day raw-retention policy based on synthetic samples.

## References

[1] [AMD SMI installation documentation](https://rocm.docs.amd.com/projects/amdsmi/en/latest/install/install.html)

[2] [AMD Ryzen AI `xrt-smi` NPU management interface](https://ryzenai.docs.amd.com/en/latest/xrt_smi.html)

[3] [NVIDIA DGX Spark Dashboard guide](https://docs.nvidia.com/dgx/dgx-spark/dgx-dashboard.html)
