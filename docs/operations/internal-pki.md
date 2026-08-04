# Internal PKI for Framework and Asus replicas

NodeScope uses an offline root CA to issue separate **replica server certificates** and **agent client certificates**. The root private key must not reside on a NodeScope replica, agent host, container image, repository, or backup. Store it offline in protected administrator-controlled storage.

## Certificate inventory

| Certificate | Subject and SAN requirements | Extended key usage | Default validity |
|---|---|---|---|
| Offline root | `NodeScope Offline Root CA` | CA signing and CRL signing | 10 years |
| Framework replica | Framework hostname plus `10.116.2.145` IP SAN | Server authentication | 90 days |
| Asus replica | Asus hostname plus `10.116.2.56` IP SAN | Server authentication | 90 days |
| Agent certificate | Stable agent identity such as `agent-framework` | Client authentication | 90 days |

A replica certificate must include every exact DNS name and IP address agents use. Agents verify the replica certificate against the root CA and do not permit insecure TLS fallback. Replicas may require and verify agent client certificates when `NODESCOPE_REQUIRE_AGENT_MTLS=true`.

## Offline root initialization

On an administrator workstation only:

```bash
nodescope-pki init-root \
  --output-directory /secure/offline/nodescope-root \
  --common-name "NodeScope Offline Root CA" \
  --years 10
```

Move `root-ca-key.pem` to offline protected storage immediately. Copy only `root-ca.pem` to Framework, Asus, and enrolled agent hosts.

## Issuing replica certificates

```bash
nodescope-pki issue --kind replica --common-name framework \
  --ca-certificate /secure/offline/nodescope-root/root-ca.pem \
  --ca-key /secure/offline/nodescope-root/root-ca-key.pem \
  --certificate-output framework-replica.pem \
  --key-output framework-replica-key.pem \
  --dns-san framework.nodescope.lan --ip-san 10.116.2.145 --days 90
```

Issue Asus with its own common name and IP SAN. Set replica keys to root-readable `0600` and certificate files to `0644`. Configure the Docker Compose secret or protected bind mount; never include a private key in an image or environment variable.

## Issuing agent certificates

```bash
nodescope-pki issue --kind agent --common-name agent-framework \
  --ca-certificate /secure/offline/nodescope-root/root-ca.pem \
  --ca-key /secure/offline/nodescope-root/root-ca-key.pem \
  --certificate-output agent-framework.pem \
  --key-output agent-framework-key.pem --days 90
```

The agent service account reads its client certificate and key from root-managed paths. The bearer credential remains a separate systemd credential file, so certificate revocation and API-credential rotation are independent.

## Renewal and rotation

NodeScope surfaces warnings at 30 and 14 days before expiry. Renew server and agent certificates before expiry, validate both endpoints, reload the server/agent, and retain the prior certificate/key until the new material is verified. The current implementation uses short-lived certificates and replacement; a formal CRL/OCSP path remains a post-Release-1 operational extension.

## External acceptance gate

No internal CA has been initialized or deployed yet. The owner must approve the final hostnames/IP SANs, offline-key custody, and renewal operator before mTLS is activated on Framework and Asus.
