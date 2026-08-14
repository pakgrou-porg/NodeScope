# Framework Docker and Portainer Deployment

This guide deploys **one complete preferred NodeScope server replica** on Framework. It is a deployment package, not acceptance evidence: Asus, agent routing, replica failover, backup leasing, certificate revocation, and restore remain separate gated exercises.

> The Framework stack publishes TLS on `10.116.2.145:8443` and advertises Asus (`10.116.2.56:8443`) as the ordered secondary endpoint. It does not start until a Docker-capable Framework host, protected runtime files, and an issued Framework replica certificate are available.

## 1. Obtain the reviewed source

In Portainer, create a **Git repository stack** from `pakgrou-porg/NodeScope` and select a reviewed commit or approved signed release tag. Set the compose path to:

```text
deploy/portainer/framework-stack.yaml
```

Git-stack mode is required for the included local build context. For a later verified image deployment, set `NODESCOPE_IMAGE_REF` to the approved image digest and remove the `build:` section only through a reviewed stack revision.

## 2. Stage host-owned protected files

On Framework, create these paths before deploying. Do not place the runtime file, certificate, key, or root private key in Portainer variables, the Git repository, bind-mounted application source, or any backup of source control.

| Host path | Required content | Owner and mode |
| --- | --- | --- |
| `/srv/nodescope/runtime/runtime.env` | `NODESCOPE_RUNTIME_DB_PASSWORD` only, using the least-privilege runtime login password. | `root:root`, `0600` |
| `/srv/nodescope/certs/server.crt` | Framework replica certificate including Framework’s approved hostname and `10.116.2.145` SAN. | `root:root`, `0644` |
| `/srv/nodescope/certs/server.key` | Matching Framework replica private key. | `root:root`, `0600` |
| `/srv/nodescope/certs/root-ca.pem` | Internal root CA public certificate, retained for operators and future agent mTLS staging. | `root:root`, `0644` |

Issue the Framework leaf from the offline CA using the exact procedure in [Internal PKI](internal-pki.md). The offline CA private key must never be copied to Framework.

## 3. Set non-secret Portainer environment values

Copy [`deploy/portainer/framework-stack.env.example`](../../deploy/portainer/framework-stack.env.example) into Portainer’s stack environment editor and replace `YOUR_SHARED_PROJECT` values. The values are configuration, not credentials. Keep `NODESCOPE_REPLICA_ID=framework`, `NODESCOPE_REPLICA_ROLE=preferred`, and distinct credential-free HTTPS primary and secondary endpoints.

Create the protected runtime file from [`deploy/portainer/runtime.env.example`](../../deploy/portainer/runtime.env.example) on the host, not in Portainer. Do not add database passwords, service-role keys, agent credentials, MCP tokens, or private keys to the stack environment editor.

## 4. Run validation before deployment

From the reviewed checkout on Framework, create `deploy/compose/replica.env` using the non-secret values, then run:

```bash
./scripts/preflight-cloud-replica-compose.sh
```

The preflight confirms Docker and Compose v2, the protected file types and permissions, endpoint ordering, secret-placement bans, and canonical compose expansion. It does not pull, build, start, stop, or remove containers. Fix every preflight failure before selecting **Deploy the stack** in Portainer.

## 5. First deployment and verification

After separate approval, deploy the Framework stack in Portainer. Verify that the service is healthy, the container is read-only and non-privileged, and Framework serves only its configured TLS listener. Then run the approved authenticated server and agent preflight sequence. Do not point agents at Framework, add Asus, enable agent mTLS, or declare failover operational until the separate replica and host-qualification gates are approved.

## Rollback

If startup, TLS, database, health, or configuration validation fails, stop the Framework stack in Portainer. Preserve only redacted container status and diagnostic metadata. Restore the previously verified image or Git commit, restore the prior certificate/key pair only if still valid, and rerun the compose preflight before another deployment attempt. Do not weaken `read_only`, `cap_drop`, `no-new-privileges`, certificate mounts, or protected secret-file permissions to bypass a failure.
