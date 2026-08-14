# Cloud Replica Canary Deployment Prerequisites

**Status:** deployment preparation only. No cloud host, container, certificate, runtime secret, agent, or route has been deployed through this procedure.

> A single approved cloud host may validate one complete NodeScope server-stack replica. It cannot establish a dual-replica failover claim. A second independently reachable complete replica is required before agent failover, failback, backup lease handoff, or replica acceptance is asserted.

## Required host preparation

| Boundary | Requirement | Preflight evidence |
| --- | --- | --- |
| Host access | Administrator-approved SSH or equivalent host-operation path; no credential is committed or pasted into repository configuration. | Access method and rollback owner are recorded outside Git. |
| Container runtime | Docker Engine with Compose v2 is installed and usable by the deployment operator. | `docker compose version` passes in the preflight. |
| Source and image | Clean checkout of the approved GitHub commit or verified GHCR image tag. | Commit SHA or image digest and checksum/attestation result are recorded. |
| Replica configuration | A root-controlled `deploy/compose/replica.env` defines a nonblank replica ID, valid role, and distinct credential-free HTTPS primary/secondary endpoints. | The preflight rejects missing, identical, non-HTTPS, credential-bearing, query-bearing, or fragment-bearing endpoints. |
| Secrets and certificates | Protected runtime secret file, certificate directory, certificate, and private key exist as regular local files. Protected files are not world-readable. | The preflight validates presence, type, and permissions without reading or printing values. |
| Runtime isolation | Compose preserves read-only filesystem, dropped capabilities, no-new-privileges, tmpfs, read-only certificate/runtime mounts, and the in-image health probe. | Deterministic contract checks the compose manifest. |

## Staged execution

1. The administrator stages only the approved checkout, `replica.env`, protected runtime secrets, and issued server certificate on the designated cloud host.
2. The operator runs `scripts/preflight-cloud-replica-compose.sh` from the checkout. The script validates Docker/Compose, endpoint boundaries, local runtime and certificate file type/permissions, secret-placement bans, and `docker compose config --quiet`. It does **not** build, pull, start, stop, or remove a container.
3. The operator records the preflight result and requests a separate authorization to deploy one canary replica. That authorization must identify the exact image digest or source commit, host, bind address, TLS port, certificate serial, runtime-secret version identifier, and rollback owner.
4. Before a second replica or any agent route is enabled, repeat the preflight on that independent host and obtain separate authorization for the failover drill.

## Recovery and stop conditions

Stop before deployment if a protected file is missing, symlinked, or world-readable; the endpoints are not distinct credential-free HTTPS URLs; Compose validation fails; the host lacks Docker Compose v2; the health probe is absent; or the approved image/source evidence differs. Remediate the staging state outside Git, rerun the preflight, and retain only redacted output. A preflight failure must never be bypassed by disabling TLS, moving credentials into `replica.env`, weakening filesystem permissions, or changing the compose security controls.

## Local validation

`./scripts/test-preflight-cloud-replica-compose-contract.sh` verifies the deployment-preflight safety contract locally. `./scripts/preflight-cloud-replica-compose.sh` is intentionally host-bound and requires an approved staged cloud environment before it can perform the final Compose expansion.
