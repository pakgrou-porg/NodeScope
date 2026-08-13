# NodeScope Manual Installation Guide: Framework and Asus GX10

> **Classification:** Internal Restricted. This runbook contains deployment topology, database-role names, and operational security controls. Do not publish environment-specific copies.
>
> **Status:** Controlled draft; non-production until the live validation and approval gates in this runbook are closed.
> **Version:** 0.4-draft
> **Publication date:** 2026-08-13
> **Author:** Manus AI  
> **Owner:** NodeScope Administrator  
> **Approver:** Pending designated NodeScope Administrator approval  
> **Canonical source:** `docs/agents/manual-install-framework-asus-v2.md` in the NodeScope repository  
> **Required source revision:** A signed NodeScope release tag or approved full commit SHA; do not use a moving branch name.
> **Deployment record:** For every use, retain the release tag, full source revision, artifact checksum, SBOM digest, attestation reference, operator, owner, approver, date, host, and response outcome in the Internal Restricted deployment appendix.

## 1. Purpose, scope, and release gate

This guide installs the **native NodeScope telemetry agent** on the Framework and Asus GX10 hosts. It is designed for a shared Supabase project in which NodeScope owns only the `nodescope` schema and must not read, write, alter, or migrate TTRPG-OCR objects.

The agent collects only permitted operational telemetry. It does **not** receive Supabase, database-runtime, database-migrator, enrollment, verification, storage-audit, or proxy credentials. It never stores inference prompts, response content, process arguments, environment variables, raw command output, or fabricated hardware values.

> **Release gate.** Do not use this guide for production enrollment until the approved release includes: the hardened installer; systemd credential support; the `0014_least_privilege_agent_operations` migration; dedicated enroller, verifier, and storage-auditor logins; and a published signed artifact, checksum, SBOM, and provenance attestation. Until then, treat every installation as a non-production validation exercise.

| Host | Agent architecture | Collection status | Important data rule |
|---|---:|---|---|
| Framework | `linux/amd64` | CPU, RAM, storage, temperature, processes, and standard Linux data are supported. AMD GPU/NPU metrics on Fedora are **experimental** until an exact tested matrix is published. | Do not describe shared system memory as VRAM. |
| Asus GX10 | `linux/arm64` | DGX host, OS-memory, temperature, selected-service, and approved-container paths are supported subject to host preflight. | `nvidia-smi` dedicated `Memory-Usage` may be `Not Supported`; this is an expected UMA state, not zero and not an error. |

The official current requirements should be retained with the installation record: AMD's published Ryzen AI Linux and ROCm paths identify Ubuntu 24.04 support rather than a qualified Fedora matrix, while NVIDIA states that DGX Spark iGPU systems may report dedicated framebuffer memory as unsupported and recommends OS memory, swap, and huge-page context.[1] [2] [3]

## 2. Non-negotiable safeguards

The following controls apply throughout the procedure.

| Control | Required behavior |
|---|---|
| Secret boundaries | Never put agent bearer tokens in shell history, terminal output, SQL text, ordinary environment variables, `agent.env`, Git, tickets, or chat. |
| Database boundaries | Only an administrator workstation uses the enrollment, verifier, storage-auditor, or migration login. An agent host must never have one. |
| TLS | Use HTTPS for both replicas. Direct PostgreSQL administration uses `sslmode=verify-full` and trusted server identity validation, never `sslmode=disable` or `sslmode=require` alone.[4] |
| Docker | Container inventory is disabled by default. Do **not** add `nodescope` to the `docker` group: Docker documents that this grants root-level privileges.[5] |
| Provenance | Install only an approved signed release/commit with verified checksum, SBOM, and artifact attestation. |
| Hardware values | Render or record unavailable values explicitly. Never synthesize VRAM, especially on UMA platforms. |
| Failure handling | Abort on a failed verification step. Do not continue by substituting placeholders, skipping TLS verification, or manually inventing UUIDs. |

## 3. Operator prerequisites

An administrator performs the privileged preparation from a hardened workstation. The workstation must have `git`, `gh`, `sha256sum`, `systemd-analyze`, an SSH client, and a PostgreSQL client. It must have access to the NodeScope release artifact and to the shared Supabase project through a protected administrator credential mechanism.

The administrator needs these **non-secret** deployment values. Keep them in an Internal Restricted deployment appendix rather than in this source guide.

| Variable | Example purpose | Where it belongs |
|---|---|---|
| `NODESCOPE_RELEASE_TAG` | Approved signed NodeScope release tag | Deployment appendix |
| `NODESCOPE_RELEASE_COMMIT` | Full approved commit SHA | Deployment appendix |
| `NODESCOPE_FRAMEWORK_HOST` | Framework SSH hostname or address | Deployment appendix |
| `NODESCOPE_ASUS_HOST` | Asus SSH hostname or address | Deployment appendix |
| `NODESCOPE_PRIMARY_ENDPOINT` | Framework replica HTTPS URL | Per-host non-secret configuration |
| `NODESCOPE_SECONDARY_ENDPOINT` | Asus replica HTTPS URL | Per-host non-secret configuration |
| `NODESCOPE_CA_FINGERPRINT` | Approved internal CA SHA-256 fingerprint | Deployment appendix |
| `NODESCOPE_DB_HOSTNAME` | Shared Supabase PostgreSQL hostname | Protected administrator configuration |

The primary and secondary agent endpoints must be distinct credential-free HTTPS destinations. NodeScope rejects user information, query strings, fragments, and canonical duplicates so an apparent failover pair cannot silently collapse to one replica. Native telemetry delivery and authenticated preflight do not follow redirects; a redirect is treated as a failed configured replica rather than an instruction to forward telemetry or its bearer credential elsewhere. The following values are secrets and must remain only in their dedicated password manager, credential store, or mode-`0600` password file: database login passwords, agent bearer tokens, internal CA private keys, and server private keys.

## 4. Verify release provenance before building or installing

The public source repository is [`pakgrou-porg/NodeScope`](https://github.com/pakgrou-porg/NodeScope), but a branch tip is not an approved production release. GitHub artifact attestations record a build's repository, commit SHA, workflow, and trigger; signed tags can be independently verified; Go can verify cached module integrity.[6] [7] [8]

Create an empty working directory on the administrator workstation and use the release values from the deployment appendix.

```bash
set -euo pipefail
umask 077

export NODESCOPE_RELEASE_TAG='REPLACE_WITH_APPROVED_SIGNED_TAG'
export NODESCOPE_RELEASE_COMMIT='REPLACE_WITH_APPROVED_FULL_COMMIT_SHA'

git clone --no-checkout https://github.com/pakgrou-porg/NodeScope.git nodescope-release
cd nodescope-release
git fetch --tags --force
git tag -v "$NODESCOPE_RELEASE_TAG"
git checkout --detach "$NODESCOPE_RELEASE_COMMIT"
test "$(git rev-parse HEAD)" = "$NODESCOPE_RELEASE_COMMIT"
test -z "$(git status --porcelain)"

go mod verify
gofmt -d $(find cmd internal telemetry -name '*.go' -print) | tee /tmp/nodescope-gofmt.diff
test ! -s /tmp/nodescope-gofmt.diff
go test ./...
pnpm install --frozen-lockfile
pnpm contracts:check
pnpm check
pnpm test
```

The release owner must provide the expected SHA-256 values, SBOM location, and artifact-attestation verification command in the release manifest. Verify them before copying either binary or unit file to a host.

```bash
# Replace filenames and expected repository/commit policy values with the signed-release manifest.
sha256sum -c SHA256SUMS
# Example only; use the published artifact and approved owner/repository policy.
gh attestation verify ./nodescope-agent-linux-amd64 --repo pakgrou-porg/NodeScope
```

> Do not run `gofmt -w` as a verification step. A formatter that changes source invalidates the clean-tree and approved-revision evidence.

## 5. Platform preflight before installation

Build an architecture-specific binary from the pinned revision only when a verified release artifact is unavailable for the approved validation exercise.

```bash
# From the verified detached NodeScope checkout.
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=true -o dist/nodescope-agent-linux-amd64 ./cmd/nodescope-agent
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -buildvcs=true -o dist/nodescope-agent-linux-arm64 ./cmd/nodescope-agent
sha256sum dist/nodescope-agent-linux-amd64 dist/nodescope-agent-linux-arm64 deploy/agent/systemd/nodescope-agent.service
```

Copy the correct binary, its release-manifest checksum, the systemd unit, and its release-manifest checksum to a protected administrator staging directory on the target. Do not install directly from `/tmp`, a writable shared directory, or an unverified working tree.

```bash
# Framework example; substitute the deployment-appendix hostname.
scp dist/nodescope-agent-linux-amd64 admin@"$NODESCOPE_FRAMEWORK_HOST":/var/tmp/nodescope-staging/
scp deploy/agent/systemd/nodescope-agent.service admin@"$NODESCOPE_FRAMEWORK_HOST":/var/tmp/nodescope-staging/

# Asus example; substitute the deployment-appendix hostname.
scp dist/nodescope-agent-linux-arm64 admin@"$NODESCOPE_ASUS_HOST":/var/tmp/nodescope-staging/
scp deploy/agent/systemd/nodescope-agent.service admin@"$NODESCOPE_ASUS_HOST":/var/tmp/nodescope-staging/
```

On each host, run the agent preflight from the staged binary before enabling service collection.

```bash
sudo /var/tmp/nodescope-staging/nodescope-agent-linux-ARCH --preflight
```

Record the output as capability states only. Do not upload raw command output to NodeScope. A missing tool must render a capability as unavailable and include a remediation reference; the agent must not install dependencies automatically.

After the protected credential file, endpoint identities, and CA trust are configured, validate authenticated replica connectivity before any `--once` collection or service start.

```bash
sudo /var/tmp/nodescope-staging/nodescope-agent-linux-ARCH --ingestion-preflight
```

This sends an authenticated `GET /api/v1/ingest/preflight` only. The server authenticates the credential and returns the selected endpoint, canonical agent/host identity, replica identity, and version; it does not decode, queue, or persist telemetry. A credential rejection fails closed. Transient failures may use the ordered secondary replica, while a repeatedly failing endpoint is temporarily skipped by the sender circuit. Record only the returned identity evidence; never record the credential or bypass TLS validation.

### 5.1 Framework AMD GPU and NPU status

On Fedora, classify AMD SMI, XRT, and XDNA/NPU collection as **experimental** until the NodeScope release's compatibility appendix explicitly qualifies the exact Fedora release, kernel, firmware, ROCm, AMD SMI, XRT, and XDNA versions. Do not use unverified `dnf` package guesses from this guide. The documented AMD NPU installation workflow presently targets Ubuntu packages and should not be transposed to Fedora without qualification.[1] [2]

### 5.2 Asus GX10 UMA status

For Asus, a dedicated GPU framebuffer-memory metric may be `not_supported`. This is expected on DGX Spark UMA hardware. Preserve separately labeled OS `MemAvailable`, `SwapFree`, huge-page data, and per-process GPU-memory values if exposed. Do not convert any of them into a VRAM percentage or a synthetic free-VRAM number.[3]

### 5.3 Docker inventory status

Leave Docker inventory disabled unless an approved least-privilege integration exists. The current safe default is no Docker socket access. Do not add the NodeScope account to the `docker` group, mount `/var/run/docker.sock`, or use a broad root helper merely to populate inventory.[5]

## 6. Install the verified binary and service atomically

The hardened installer requires four arguments: the binary, its expected SHA-256, the systemd unit, and its expected SHA-256. It rejects symlinks, stages root-owned copies, verifies staged checksums, preserves the previous binary, and atomically replaces live files.

```bash
sudo /path/to/install-linux.sh \
  /var/tmp/nodescope-staging/nodescope-agent-linux-ARCH \
  'REPLACE_WITH_AGENT_BINARY_SHA256' \
  /var/tmp/nodescope-staging/nodescope-agent.service \
  'REPLACE_WITH_UNIT_FILE_SHA256'
```

The target must run systemd 250 or newer because the service uses `LoadCredential=`. This mechanism gives the service a per-activation credential file, avoids environment propagation, and can later use encrypted credentials where the target policy permits it.[9] [10]

Confirm the service definition before enrollment.

```bash
sudo systemd-analyze verify /etc/systemd/system/nodescope-agent.service
sudo systemd-analyze security nodescope-agent.service
sudo systemctl cat nodescope-agent.service
```

Record `systemd-analyze security` findings and any justified exemptions in the deployment appendix. Do not weaken `ProtectSystem`, `NoNewPrivileges`, or credential isolation simply to improve a score.

## 7. Install the internal CA and verify both replica endpoints

Install only the approved public CA certificate or the approved internal CA certificate. Verify its SHA-256 fingerprint against the deployment appendix before trusting it.

```bash
sudo install -o root -g nodescope -m 0640 /var/tmp/nodescope-staging/nodescope-ca.pem /etc/nodescope-agent/ca.pem
sha256sum /etc/nodescope-agent/ca.pem
```

Verify **both** endpoints before enrollment. The hostname in each command must match a certificate Subject Alternative Name; do not use `-k`, `InsecureSkipVerify`, or an IP address unless that IP address is explicitly present in the certificate SAN.

```bash
# Repeat once for the primary and once for the secondary endpoint.
openssl s_client \
  -connect PRIMARY_HOSTNAME:8443 \
  -servername PRIMARY_HOSTNAME \
  -CAfile /etc/nodescope-agent/ca.pem \
  -verify_return_error </dev/null

curl --fail --silent --show-error \
  --cacert /etc/nodescope-agent/ca.pem \
  https://PRIMARY_HOSTNAME:8443/healthz
```

Repeat the same checks for the secondary endpoint. Record certificate Subject, SAN, serial, SHA-256 fingerprint, and expiry date. A successful TLS handshake alone is insufficient: verify endpoint health and the intended server identity.

## 8. Secure enrollment and rotation

### 8.1 Required implementation boundary

Enrollment must use the dedicated `nodescope_enroller` login and the schema-local `nodescope.enroll_or_rotate_agent(...)` function introduced by migration `0014_least_privilege_agent_operations`. The function obtains the canonical host ID from the slug, preserves the existing agent ID on rotation, increments the rotation version, stores only a credential digest, and inserts an audit event.

The enrollment utility must generate the bearer token locally, hash it locally, send **only the SHA-256 digest** to the database, and atomically write the plaintext token to `/etc/nodescope-agent/credentials/agent-token` with root-only permissions. It must not print the token or accept it in a flag, SQL variable, or normal environment variable.

> **Implementation gate.** If the `nodescope-enroll` utility and migration `0014` are not in the approved release, stop here. Do not fall back to manual SQL or manually paste a bearer token into `agent.env`.

### 8.2 Administrator database session

The administrator workstation uses a temporary protected password-file path or interactive authentication. Do not export `PGPASSWORD`. PostgreSQL requires an appropriately protected password file and supports server identity verification with `sslmode=verify-full`.[4] [11]

```bash
# PGPASSFILE must already be created from a protected secret source, mode 0600,
# and removed after the enrollment session. Do not put its contents in this guide.
export PGPASSFILE=/root/.config/nodescope/pgpass.enroller
export PGSSLMODE=verify-full
export PGSSLROOTCERT=/root/.config/nodescope/supabase-ca.pem
chmod 0600 "$PGPASSFILE" "$PGSSLROOTCERT"

# Confirm host identity before invoking any NodeScope enrollment command.
psql "host=REPLACE_WITH_DB_HOST port=5432 dbname=postgres user=nodescope_enroller" \
  -c 'select current_user, session_user, ssl from pg_stat_ssl where pid = pg_backend_pid();'
```

### 8.3 Enrollment command shape

The approved release will provide a command with this behavior. The command below is an interface contract, not a request to place a token in the shell.

```bash
sudo nodescope-enroll \
  --host-slug framework \
  --display-name Framework \
  --platform 'fedora-linux-amd64' \
  --address 'REPLACE_WITH_STATIC_HOST_ADDRESS' \
  --credential-path /etc/nodescope-agent/credentials/agent-token \
  --expires-in 90d
```

The successful command returns only the non-secret canonical `host_id`, stable `agent_id`, rotation version, expiry time, and a non-secret credential hint. Populate the non-secret configuration file from those returned identifiers.

```bash
sudoedit /etc/nodescope-agent/agent.env
```

Use this shape, with real values from enrollment and deployment appendix:

```dotenv
NODESCOPE_AGENT_ID=REPLACE_WITH_STABLE_AGENT_ID
NODESCOPE_HOST_ID=REPLACE_WITH_CANONICAL_HOST_ID
NODESCOPE_AGENT_CREDENTIAL_FILE=/etc/nodescope-agent/credentials/agent-token
NODESCOPE_PRIMARY_ENDPOINT=https://PRIMARY_HOSTNAME:8443
NODESCOPE_SECONDARY_ENDPOINT=https://SECONDARY_HOSTNAME:8443
NODESCOPE_COLLECTION_INTERVAL_SECONDS=5
NODESCOPE_CA_CERT_PATH=/etc/nodescope-agent/ca.pem
# Set all three values when replicas set NODESCOPE_REQUIRE_AGENT_MTLS=true.
NODESCOPE_REQUIRE_CLIENT_MTLS=true
NODESCOPE_TLS_CLIENT_CERT_PATH=/etc/nodescope-agent/agent.crt
NODESCOPE_TLS_CLIENT_KEY_PATH=/etc/nodescope-agent/agent.key
NODESCOPE_SELECTED_PROCESS_NAMES=lmstudio,vllm
# Docker inventory remains disabled unless a separately approved narrow proxy/helper is deployed.
NODESCOPE_DOCKER_INVENTORY_ENABLED=false
NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES=
```

Confirm that `agent.env` contains no value beginning with a bearer token or database URL, and that the persistent credential file is root-only.

```bash
sudo stat -c '%U:%G %a %n' /etc/nodescope-agent/agent.env /etc/nodescope-agent/credentials/agent-token
sudo grep -nE 'CREDENTIAL|PASSWORD|DATABASE_URL|SUPABASE' /etc/nodescope-agent/agent.env && exit 1 || true
```

The `NODESCOPE_AGENT_CREDENTIAL_FILE` reference is allowed because it contains no secret. When replica policy requires agent mTLS, `NODESCOPE_REQUIRE_CLIENT_MTLS=true` requires the internal CA and both client certificate paths before the agent can start. Never disable certificate or hostname verification to work around an internal PKI configuration error.

Credential rotation uses the same `nodescope-enroll` command and host slug. It rotates only the credential, preserves the stable agent identity, records an audit event, atomically replaces the credential file, and requires a service restart. Revocation, expiry, and rotation must be visible in NodeScope audit and administration views.

## 9. Start, inspect, and validate non-content telemetry

Start the service only after CA validation and enrollment complete.

Before starting the service, verify credential authentication and ordered replica connectivity without constructing or persisting telemetry.

```bash
sudo /usr/local/bin/nodescope-agent --ingestion-preflight
```

The command sends an authenticated `GET /api/v1/ingest/preflight` only and returns the selected endpoint, canonical agent and host identities, replica identity, and version. Credential rejection fails closed; transient failures can use the ordered secondary replica, while a repeatedly failing endpoint is temporarily skipped.

```bash
sudo systemctl daemon-reload
sudo systemctl restart nodescope-agent.service
sudo systemctl status --no-pager nodescope-agent.service
sudo journalctl -u nodescope-agent.service -n 100 --no-pager
```

The logs may contain health state, endpoint label, status class, and non-secret identifiers. They must not contain bearer tokens, request bodies, protobuf payloads, prompt content, response content, process arguments, environment values, database URLs, or raw tool output.

The native sender must test the preferred endpoint on each cycle, fail over only for transient transport or server failures, and fail back to the preferred endpoint when healthy. It should use exponential retry delay with jitter and circuit breaking; client errors such as invalid credentials must fail closed rather than spill traffic to another endpoint.

## 10. Least-privilege verification and storage evidence

Routine verification uses the dedicated `nodescope_verifier` login, not `nodescope_owner`. The native verifier invokes the narrow schema-local status function and evaluates the server receipt timestamp, metric-state cardinality, and stale-state evidence. `count(m.metric_name)` remains intentional: a host with no metric rows must report zero, not a misleading outer-join row count.

```bash
# Keep the verifier URL only in a root-managed environment file or credential
# source; never paste it into shell history or a shared runbook.
export NODESCOPE_VERIFIER_DATABASE_URL='REPLACE_WITH_VERIFIER_DATABASE_URL'

sudo /usr/local/bin/nodescope-verify \
  --slug framework \
  --max-receipt-age-seconds 15 \
  --require-fresh \
  --output-dir /var/lib/nodescope-evidence/reports \
  | tee /var/lib/nodescope-evidence/latest-framework-host-verification.json
```

The verifier emits server receipt-time evidence and atomically writes a dynamically named, mode-`0600` report when an output directory is supplied. `--require-fresh` exits nonzero if the server receipt is missing or exceeds the configured freshness window, current metric state is absent, or stale metric state is present. Any unavailable or stale count is a data-quality state, not zero utilization. `clock_offset_seconds` is observed by the server and must be investigated when its quality is stale. The server rejects envelope or metric observation timestamps more than one minute ahead of server receipt time; correct host time synchronization rather than retrying a rejected future-dated batch, because such a batch could otherwise block later valid latest-state updates.

## 11. Seventy-two-hour storage-feasibility gate

Do not finalize two-day raw retention on estimates alone. Run the agent on Framework and Asus with representative selected processes and approved collection paths for at least 72 continuous hours. The window begins only after both hosts show accepted server receipts.

Use the `nodescope_storage_auditor` login and the server receipt time, not agent-provided observation time. The evidence function reports completeness, maximum gap, compressed payload percentiles, cardinality, and physical table/index growth.

```bash
set -euo pipefail
umask 077

# Store this exact URL only in a root-managed environment file or credential
# source readable by the storage-auditor workflow, never in shell history.
export NODESCOPE_STORAGE_AUDITOR_DATABASE_URL='REPLACE_WITH_STORAGE_AUDITOR_DATABASE_URL'

sudo /usr/local/bin/nodescope-storage-evidence \
  --slug framework \
  --since "$(date -u -d '72 hours ago' +%Y-%m-%dT%H:%M:%SZ)" \
  --collection-interval-seconds 5 \
  --require-complete \
  --output-dir /var/lib/nodescope-evidence/reports \
  | tee /var/lib/nodescope-evidence/latest-framework-storage-evidence.json
```

The command returns receipt-time evidence and atomically publishes a dynamically named, mode-`0600` JSON report under the selected output directory. `--collection-interval-seconds` accepts only the supported one-to-sixty-second range; an out-of-range value makes the assessment incomplete rather than producing a misleading completeness or gap conclusion. `--require-complete` exits nonzero if receipts are missing, expected or received batches are absent, gap exceeds three collection intervals, metric cardinality is zero, or a size/completeness value is invalid. Run the same command for Asus. The release gate requires the following evidence to be reviewed together: expected versus received batches; completeness percentage; maximum gap; median, p95, and p99 compressed-batch size; metric cardinality; raw and index relation growth; rollup/deletion job timings; and capacity-governor state. Do not extrapolate a final storage plan from one low-load sample or a manually assumed constant.

## 12. Failure, rollback, and decommissioning

If service startup fails, stop it before changing any configuration. Inspect `systemctl status`, sanitized journal output, certificate verification, credential-file ownership, and non-secret config syntax. Do not weaken TLS, add the agent to Docker groups, or replace an unavailable metric with a guessed value to make the dashboard appear healthy.

The installer preserves the previous binary under `/var/lib/nodescope-installer/backups/` with a checksum. A rollback is an administrator action: stop the service, verify the retained backup checksum, atomically restore the binary, run `systemd-analyze verify`, and restart. Credential rollback is not automatic; rotate through the enrollment workflow and record the reason in audit history.

To decommission a host, revoke its agent credential through the NodeScope administration or enrollment control, stop and disable the service, remove its persistent credential file, and retain only the required audit record. Do not delete shared-Supabase schemas, global extensions, or TTRPG-OCR data.

## 13. Report response matrix

Treat every command result as operational evidence rather than a dashboard-only status. The operator must retain the generated report path and response outcome in the deployment appendix; an Administrator owns the decision to resume, rollback, or escalate.

| Evidence or report outcome | Required response | Record owner | Escalation boundary |
|---|---|---|---|
| Release evidence verifies checksum, SBOM, attestation, and immutable revision | Continue to the next controlled installation step. | Installing operator | Administrator approves production use only after all Release Gate items are met. |
| Release evidence, source revision, checksum, SBOM, or attestation is missing or mismatched | Stop. Do not install, update, or run an artifact. Obtain a newly verified release bundle. | Installing operator | Administrator and release owner. |
| Authenticated ingestion preflight succeeds on the preferred replica | Record the canonical replica evidence and proceed to one-shot collection only if the deployment gate permits it. | Installing operator | Escalate if the result differs from the deployment appendix. |
| Authenticated preflight reaches only a secondary replica | Record failover evidence; investigate the preferred replica before production approval. | Installing operator | Replica owner and Administrator. |
| Verifier report is not fresh or shows no current metric state | Stop the validation run. Inspect agent service, credential, CA, endpoint, and server receipt evidence; do not fabricate values. | Host operator | Administrator if credentials, replica routing, or audit state are implicated. |
| Storage-evidence report is incomplete, has a receipt gap, lacks cardinality, or has invalid size evidence | Do not approve retention capacity. Extend or restart the benchmark after resolving the collection or database issue. | Storage auditor | Administrator and capacity owner. |
| Metric is unavailable, stale, or experimental | Preserve its explicit quality/provenance. Do not infer zero utilization, VRAM, or readiness. | Host operator | Platform owner when qualification is required. |
| Installer or service verification fails | Stop, retain sanitized evidence, and use the documented verified rollback path. Do not weaken TLS or bypass provenance checks. | Installing operator | Administrator for rollback authorization. |

## 14. Revision history

| Version | Date | Change | Status |
|---|---|---|---|
| 0.1 | 2026-07-23 | Initial manual guide. | Superseded; non-production. |
| 0.2-draft | 2026-07-23 | Added provenance gate, secret-file boundary, least-privilege roles, TLS verification, platform support classification, Docker default-off, bilateral endpoint checks, receipt-time evidence, and document controls. | Pending implementation validation and Administrator approval. |
| 0.3-draft | 2026-08-13 | Replaced manual CSV evidence queries with the native server-receipt-time report command, atomic dynamic JSON evidence, and explicit completeness failure semantics. | Pending qualified 72-hour host validation and Administrator approval. |
| 0.4-draft | 2026-08-13 | Added controlled-draft status, immutable deployment-record fields, receipt-time host-verification response matrix, explicit rollback/escalation outcomes, and runbook governance metadata. | Pending signed-release, qualified-host, and Administrator approval gates. |

## References

[1] [AMD: Ryzen AI Linux Installation Instructions](https://ryzenai.docs.amd.com/en/latest/linux.html)

[2] [AMD: ROCm Radeon and Ryzen Linux support matrix](https://rocm.docs.amd.com/projects/radeon-ryzen/en/latest/docs/compatibility/compatibilityryz/native_linux/native_linux_compatibility.html)

[3] [NVIDIA: DGX Spark known issues](https://docs.nvidia.com/dgx/dgx-spark/known-issues.html)

[4] [PostgreSQL: SSL Support](https://www.postgresql.org/docs/current/libpq-ssl.html)

[5] [Docker: Linux post-installation steps](https://docs.docker.com/engine/install/linux-postinstall/)

[6] [GitHub: Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)

[7] [GitHub: Signing tags](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-tags)

[8] [Go Modules Reference](https://go.dev/ref/mod)

[9] [systemd: System and Service Credentials](https://systemd.io/CREDENTIALS/)

[10] [systemd-creds manual](https://www.freedesktop.org/software/systemd/man/systemd-creds.html)

[11] [PostgreSQL: Password File](https://www.postgresql.org/docs/current/libpq-pgpass.html)
