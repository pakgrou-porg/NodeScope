# Framework Auxiliary-Agent Validation Runbook

**Purpose:** authorize an auxiliary agent to prepare and validate the Framework NodeScope Portainer stack **without starting a container, creating a credential, changing a database, rotating a certificate, or enrolling a host**. The agent must stop at every confirmation boundary and return a redacted evidence bundle for review.

> The agent may inspect software versions, source revisions, file modes, public certificate metadata, Compose expansion, and service-port availability. It must not print, upload, copy, or request a database password, agent token, private key, certificate authority key, Supabase secret, service-role key, MCP token, or complete environment file.

## Scope and hard stop rules

| Activity | Auxiliary agent may run now | Explicit confirmation required | Never permitted in the agent report |
| --- | --- | --- | --- |
| Inspect Framework prerequisites | Yes | Not applicable | Secrets, private key material, environment values |
| Clone or update reviewed NodeScope source | Yes | Not applicable | Unreviewed branches or an unpinned source revision |
| Create non-secret `deploy/compose/replica.env` | Yes, after user supplies non-secret values | Not applicable | Runtime password or any credential in the file |
| Inspect protected paths and public certificate metadata | Yes | Not applicable | Private-key content or copying `/srv/nodescope/runtime/runtime.env` |
| Run the compose preflight | Yes | Not applicable | Any output containing secret values |
| Create a Portainer Git stack definition | Yes, as a draft only | **Before pressing Deploy** | Portainer stack environment secret values |
| Start, stop, build, pull, or remove a container | No | **Required** | None |
| Run agent enrollment or credential rotation | No | **Required** | Generated agent credential |
| Change database state, run migrations, alter roles, issue/revoke certificates | No | **Required** | Credentials, private keys, or raw database output |

If a command would use `docker compose up`, `docker compose down`, `docker build`, `docker pull`, `docker rm`, `docker system prune`, `portainer deploy`, `scripts/enroll-or-rotate-agent.sh`, `psql` against a state-changing function, or any certificate issuance/revocation command, the auxiliary agent must stop and request confirmation.

## Copy-ready instructions for the auxiliary agent

Give the following block to the auxiliary agent exactly as a task prompt. Replace values in angle brackets only with non-secret values supplied by the owner.

```text
You are validating, not deploying, a NodeScope Framework Portainer stack.

Hard boundaries:
1. Do NOT start, stop, build, pull, remove, or deploy containers.
2. Do NOT run agent enrollment, credential rotation, SQL mutations, migrations, certificate issuance, revocation, or backup commands.
3. Do NOT read, print, upload, or copy secret values, database URLs, runtime.env content, private keys, CA keys, tokens, or Portainer secret values.
4. If any command would cross a hard boundary, stop and report: CONFIRMATION_REQUIRED with the exact action and reason.
5. Return only the redacted evidence template in this runbook.

Target:
- Framework address: 10.116.2.145
- Repository: https://github.com/pakgrou-porg/NodeScope.git
- Required compose path: deploy/portainer/framework-stack.yaml
- Expected preferred replica ID: framework
- Expected secondary endpoint: https://10.116.2.56:8443
- Required Portainer stack mode: Git Repository

Run the numbered safe-now stages only. Stop at the first failed stage. Do not attempt a workaround that weakens permissions, TLS, secret placement, or container hardening.
```

## Safe-now stages

The auxiliary agent must execute the following stages in order. The owner supplies non-secret values only where marked.

### Stage 1 — Host and tooling inventory

```bash
set -euo pipefail
printf 'hostname='; hostnamectl --static 2>/dev/null || hostname
printf 'os='; . /etc/os-release && printf '%s %s\n' "$NAME" "$VERSION_ID"
printf 'architecture='; uname -m
printf 'docker='; docker --version
printf 'compose='; docker compose version
printf 'portainer='; docker ps --format '{{.Names}} {{.Image}}' | grep -i portainer || true
printf 'node_scope_port='; ss -ltn '( sport = :8443 )' || true
```

**Expected result:** Framework is the intended Fedora/Linux host, Docker Engine and Compose v2 are available, Portainer presence is reported, and no unexpected service already owns TCP 8443. A nonempty listener on 8443 is a stop condition unless the owner confirms it is the prior approved NodeScope replica.

### Stage 2 — Reviewed source baseline

```bash
set -euo pipefail
sudo install -d -o root -g root -m 0755 /srv/nodescope/source
sudo git clone https://github.com/pakgrou-porg/NodeScope.git /srv/nodescope/source/NodeScope 2>/dev/null || true
cd /srv/nodescope/source/NodeScope
sudo git fetch --tags origin
sudo git status --short
sudo git rev-parse HEAD
sudo git log -1 --format='%H%n%s'
test -f deploy/portainer/framework-stack.yaml
test -f deploy/portainer/framework-stack.env.example
test -f deploy/portainer/runtime.env.example
test -f scripts/preflight-cloud-replica-compose.sh
```

**Expected result:** a clean reviewed checkout contains all four paths. If the checkout is dirty or missing an artifact, stop and report the revision plus missing path. Do not use an unreviewed local modification.

### Stage 3 — Protected-path and public-certificate inspection

The owner, not the auxiliary agent, must stage the protected files through the approved PKI and secret process. The agent may inspect only metadata and the public certificate:

```bash
set -euo pipefail
for path in \
  /srv/nodescope/runtime/runtime.env \
  /srv/nodescope/certs/server.crt \
  /srv/nodescope/certs/server.key \
  /srv/nodescope/certs/root-ca.pem; do
  printf '%s ' "$path"
  sudo stat -c 'owner=%U:%G mode=%a type=%F' "$path"
done

sudo openssl x509 -in /srv/nodescope/certs/server.crt -noout -subject -issuer -serial -dates -ext subjectAltName
sudo openssl x509 -in /srv/nodescope/certs/root-ca.pem -noout -subject -issuer -serial -dates
```

**Expected result:** `runtime.env` and `server.key` are `root:root` mode `0600`; `server.crt` and `root-ca.pem` are `root:root` mode `0644`. The Framework server certificate must identify the approved Framework name and `10.116.2.145` SAN. Do not run `cat`, `sed`, `grep`, `cp`, `tar`, or `scp` on `runtime.env` or any private key.

### Stage 4 — Non-secret replica configuration

The owner supplies non-secret values for the replica environment file. The auxiliary agent may copy the example and set **only** the following values: Supabase URL and host/database/user/port identifiers, Framework replica ID and role, `https://10.116.2.145:8443` as the primary endpoint, and `https://10.116.2.56:8443` as the secondary endpoint. It must not add `NODESCOPE_RUNTIME_DB_PASSWORD`, any token, or any private-key path value not already represented by the approved template.

```bash
set -euo pipefail
cd /srv/nodescope/source/NodeScope
sudo install -d -o root -g root -m 0700 deploy/compose
sudo install -o root -g root -m 0600 \
  deploy/portainer/framework-stack.env.example \
  deploy/compose/replica.env
sudoedit deploy/compose/replica.env
sudo grep -nE 'NODESCOPE_(REPLICA_ID|REPLICA_ROLE|PRIMARY_ENDPOINT|SECONDARY_ENDPOINT|SUPABASE_URL|RUNTIME_DB_HOST|RUNTIME_DB_PORT|RUNTIME_DB_NAME|RUNTIME_DB_USER)=' deploy/compose/replica.env
sudo grep -nE 'PASSWORD|TOKEN|SECRET|KEY=' deploy/compose/replica.env && exit 1 || true
```

The `sudoedit` step requires the owner or an authorized human to type non-secret values. The final grep must show no secret assignment. If a secret is present, remove it and recreate the file from the example before continuing.

### Stage 5 — Validation-only compose preflight

```bash
set -euo pipefail
cd /srv/nodescope/source/NodeScope
sudo ./scripts/preflight-cloud-replica-compose.sh
```

**Expected result:** the script confirms Docker, Compose v2, protected file types and modes, endpoint ordering, no secret placement in the non-secret environment, and canonical Compose expansion. It must not pull, build, start, stop, or remove any container. Return the final success line or the first failure stage only.

### Stage 6 — Portainer draft review

The auxiliary agent may create or review a **draft** Portainer Git Repository stack only if the owner permits Portainer access. Use repository `pakgrou-porg/NodeScope`, a reviewed commit or approved signed tag, and compose path `deploy/portainer/framework-stack.yaml`. Enter only non-secret stack variables from `framework-stack.env.example`. Do not press **Deploy the stack**.

The agent must return a screenshot or redacted text confirmation of the repository, revision, compose path, and non-secret environment key names. It must not include environment values, runtime.env content, or certificate/key data.

## Mandatory handoff format

The auxiliary agent must return this exact redacted structure after a pass or failure:

```text
FRAMEWORK_NODESCOPE_PREFLIGHT
source_revision: <full Git SHA>
repository_clean: <yes|no>
host_os_architecture: <redacted OS/version and architecture>
docker_version: <version or unavailable>
compose_version: <version or unavailable>
portainer_detected: <yes|no|unknown>
port_8443_state: <free|approved-existing-service|unexpected-listener>
protected_file_modes:
  runtime_env: <owner/mode only>
  server_crt: <owner/mode only>
  server_key: <owner/mode only>
  root_ca: <owner/mode only>
certificate_metadata: <subject/issuer/SAN/expiry only>
replica_env_secret_scan: <pass|fail>
compose_preflight: <pass|fail>
portainer_draft: <not-created|created-not-deployed>
expected_result: <text>
observed_result: <text>
known_limitation: <text>
recovery_taken: <text or none>
confirmation_required_next: <exact next action>
```

Do not include passwords, tokens, keys, raw environment lines, database URLs, complete certificate output, or container logs with query strings in this handoff.

## Confirmation-required next actions

After the handoff passes, the owner must separately approve each action in this order: first Portainer **Deploy the stack** for Framework; then a health/TLS verification; then agent credential enrollment; then native agent installation and authenticated ingestion preflight; then any Asus secondary deployment or failover exercise. A preflight pass is not permission to take any subsequent action.

## Failure handling

If any stage fails, the auxiliary agent must stop. It may correct only a non-secret file mode, missing non-secret configuration, or checked-out source revision **after recording the failure**. It must not weaken `read_only`, `cap_drop`, `no-new-privileges`, TLS settings, certificate mounts, secret file modes, or endpoint ordering. For any certificate, database, Portainer, or Docker anomaly, return the redacted handoff and wait for an owner decision.
