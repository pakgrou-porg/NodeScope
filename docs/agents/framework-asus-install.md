# Framework and Asus agent installation

NodeScope Release 1 uses the same native Linux agent on the Framework x86-64 host and the Asus Ascent GX10 ARM64 host. The agent transmits bounded Protocol Buffers plus Zstandard telemetry to the preferred Framework replica and can fail over to Asus only after a transient transport or server failure. Credential rejection fails closed.

> **Prerequisite:** Deploy the two server replicas and internal CA before starting either agent. Agent endpoints must use HTTPS. Install the NodeScope internal CA bundle at the configured local path; do not disable hostname or certificate verification.

| Host | Target architecture | Build command |
| --- | --- | --- |
| Framework | `linux/amd64` | `GOOS=linux GOARCH=amd64 go build -o dist/nodescope-agent-linux-amd64 ./cmd/nodescope-agent` |
| Asus GX10 | `linux/arm64` | `GOOS=linux GOARCH=arm64 go build -o dist/nodescope-agent-linux-arm64 ./cmd/nodescope-agent` |

The installer performs no automatic package installation. It creates the unprivileged `nodescope` service account, protects the state directory, creates a root-readable environment template, and enables the service without starting it until a protected credential file is present.

```bash
sudo ./deploy/agent/install-linux.sh ./dist/nodescope-agent-linux-amd64
sudo /usr/local/bin/nodescope-agent --preflight | jq .
```

Treat the preflight report as the authoritative statement of collector availability. On Framework Fedora, AMD DRM, AMD SMI, XRT, and XDNA/NPU readings are **experimental** until the exact Fedora release, kernel, firmware, ROCm, AMD SMI, XRT, and XDNA versions are qualified in the NodeScope compatibility matrix. Do **not** run package-install commands suggested by older guidance or infer a package from a missing tool. On Asus, use the NVIDIA tooling supplied by DGX OS and preserve unavailable framebuffer memory as explicitly unavailable rather than deriving VRAM from host memory.

Populate `/etc/nodescope-agent/agent.env` only after secure agent enrollment supplies the host UUID, agent UUID, credential-file path, and internal CA certificate. The collection interval defaults to five seconds and accepts only one to sixty seconds. Process and alert-container allowlists are optional comma-separated values.

```dotenv
NODESCOPE_AGENT_ID=<agent UUID>
NODESCOPE_HOST_ID=<host UUID>
NODESCOPE_AGENT_CREDENTIAL_FILE=/run/credentials/nodescope-agent.service/agent-token
NODESCOPE_PRIMARY_ENDPOINT=https://framework.nodescope.lan:8443
NODESCOPE_SECONDARY_ENDPOINT=https://asus.nodescope.lan:8443
NODESCOPE_CA_CERT_PATH=/etc/nodescope-agent/ca.pem
# Set these three values when replicas use NODESCOPE_REQUIRE_AGENT_MTLS=true.
NODESCOPE_REQUIRE_CLIENT_MTLS=true
NODESCOPE_TLS_CLIENT_CERT_PATH=/etc/nodescope-agent/agent.crt
NODESCOPE_TLS_CLIENT_KEY_PATH=/etc/nodescope-agent/agent.key
NODESCOPE_COLLECTION_INTERVAL_SECONDS=5
NODESCOPE_SELECTED_PROCESS_NAMES=LMStudio,vllm
NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES=vllm,agent-zero
```

When client mTLS is required, the agent rejects startup unless the internal CA bundle and paired client certificate/key paths are present. Docker inventory remains disabled unless an approved fixed-schema HTTPS proxy is configured with its own required paired client certificate/key; never add the service account to the `docker` group or expose `/var/run/docker.sock`.

Before any `--once` collection or service start, verify the protected credential and ordered replica connectivity without constructing or persisting telemetry.

```bash
sudo /usr/local/bin/nodescope-agent --ingestion-preflight
```

The command sends an authenticated `GET /api/v1/ingest/preflight` only. Record the returned endpoint, canonical agent and host identities, replica ID, and version. A transient primary failure may use the ordered secondary replica, and a repeatedly failing endpoint is temporarily skipped; never record the credential or bypass TLS verification to make this test pass.

After credential and CA material are present, start the service and inspect only redacted status output.

```bash
sudo systemctl start nodescope-agent
sudo systemctl status nodescope-agent --no-pager
sudo journalctl -u nodescope-agent -n 50 --no-pager
```

The storage-feasibility gate runs only after both agents collect representative workloads for 72 hours. Export each host’s actual compact probe summary and run `nodescope-storage-estimate`; do not approve the two-day raw-retention policy from synthetic samples.
