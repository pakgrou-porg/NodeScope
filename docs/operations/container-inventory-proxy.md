# Container Inventory Proxy

NodeScope agents **do not** connect to `/var/run/docker.sock`, join the `docker` group, or mount a Docker socket. Docker socket access is effectively root-equivalent and is outside the agent’s least-privilege boundary. Container inventory is disabled by default and becomes available only through an administrator-approved, narrowly scoped HTTPS helper.

## Agent Configuration

Set the following values in the agent’s protected environment file only after the helper is available. The agent uses the same CA and optional client certificate configured for telemetry ingestion, so the proxy endpoint must be covered by the internal PKI identity presented to the agent.

| Variable | Required when inventory is enabled | Constraint |
|---|---:|---|
| `NODESCOPE_DOCKER_INVENTORY_ENABLED` | Yes | Must be exactly `true`; its default is `false`. |
| `NODESCOPE_CONTAINER_INVENTORY_PROXY_URL` | Yes | Absolute `https://` URL to the approved helper’s fixed inventory endpoint. |
| `NODESCOPE_CA_CERT_PATH` | Recommended | Internal CA bundle that validates the proxy identity. |
| `NODESCOPE_TLS_CLIENT_CERT_PATH` and `NODESCOPE_TLS_CLIENT_KEY_PATH` | Yes | Agent client identity; both variables must be set together whenever inventory is enabled. |
| `NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES` | No | Comma-separated approved IDs or names for alert selection only. |

An enabled inventory collector without an HTTPS proxy URL and paired client certificate/key fails configuration validation. A reachable proxy that returns a non-200 response produces explicit unavailable inventory evidence rather than falling back to the Docker socket.

## Fixed Response Contract

The helper must expose a read-only `GET` endpoint and return only the schema below. NodeScope limits the response to **1 MiB** and **1,024** containers, rejects duplicate IDs, unknown JSON fields, embedded NUL bytes, invalid UTF-8, and oversized text. The proxy must not include environment variables, command lines, labels containing secrets, mounts, networks, raw Docker inspect data, or Docker API errors.

```json
{
  "containers": [
    {
      "containerId": "a stable runtime identifier",
      "name": "vllm",
      "image": "registry.example/vllm:tag",
      "state": "running",
      "health": "healthy"
    }
  ]
}
```

`health` may be empty or `unreported`; NodeScope preserves that as unavailable health evidence instead of fabricating a health value. All other fields are required. The helper should use a dedicated, read-only local mechanism appropriate to the host and restrict its serving identity to the NodeScope agent certificate or a dedicated network policy.

## Verification and Rollback

Run the agent’s preflight command after configuration. It will describe the `container_inventory_proxy` prerequisite and will never offer Docker-group membership or a socket mount as remediation. After a successful collection cycle, verify that the envelope contains `container.inventory.available` with source `inventory-proxy` and no socket-related provenance.

To disable the feature, set `NODESCOPE_DOCKER_INVENTORY_ENABLED=false` and restart the native agent. No Docker permissions or agent-side cleanup is required because the agent never opens the socket.
