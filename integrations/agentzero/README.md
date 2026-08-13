# NodeScope remote MCP for AgentZero 2.5

**Classification:** Internal Restricted. This guide connects AgentZero to the NodeScope remote MCP endpoint; it does not grant database access, host SSH access, or direct access to the inference proxy’s backend credentials.

NodeScope exposes its Model Context Protocol service at `https://<replica>/mcp`. The service uses bearer credentials held in a root-managed MCP configuration file on the NodeScope replica. Each credential maps to an explicit **Viewer**, **Operator**, or **Administrator** role. The server applies authorization before every tool call and records approved control actions through the NodeScope audit protocol.

| Tool | Minimum role | Data or action boundary |
|---|---:|---|
| `nodescope_fleet_status` | Viewer | Freshness, server receipt time, metric counts, and explicit unavailable/stale state only. |
| `nodescope_acknowledge_alert` | Operator | Records an alert acknowledgement and audit event. |
| `nodescope_set_collection_interval` | Operator | Sets a global or host override from 1–60 seconds through the control plane. |
| `nodescope_refresh_storage_baseline` | Operator | Requires an explicit reviewed-diff acknowledgement and creates an auditable requested operation. |

> **Privacy boundary:** NodeScope tools never return inference prompts, completions, request bodies, response bodies, client bearer credentials, agent bearer credentials, database URLs, or certificate private keys.

## Configure AgentZero

In AgentZero, open **Settings → MCP/A2A → External MCP Servers**, then add the content from `nodescope-mcp.json.example`. Put the actual bearer value in AgentZero’s protected secret storage or environment, not in the JSON configuration, screenshots, or chat.

The preferred connection must point at the Framework replica. If the AgentZero deployment supports connection-level fallback, configure the Asus URL as the secondary endpoint. Otherwise, maintain a second disabled configuration entry and enable it only after a Framework failure is visible in NodeScope.

The MCP endpoint must be covered by the NodeScope internal CA. Import the CA public certificate into the AgentZero container or host trust store before applying the connection. Do not disable TLS verification as a workaround.

## Validate after deployment

AgentZero should list four NodeScope tools after the connection becomes healthy. Begin with `nodescope_fleet_status`, verify that returned freshness states match the browser console, and then run one controlled Operator action in a non-production maintenance window. Confirm the paired audit entry in NodeScope before enabling autonomous agent operations.

## Required deferred inputs

This integration cannot be activated until the internal CA, framework/asus HTTPS endpoints, protected MCP client credential, and the NodeScope MCP configuration file are provisioned. The implementation is present in the server; these are deployment and validation gates, not missing application code.

## Local compatibility contract

NodeScope validates the exact configuration shape documented by AgentZero 2.5: a remote HTTPS URL ending in `/mcp` and an `Authorization: Bearer ${NODESCOPE_AGENTZERO_MCP_TOKEN}` header in the secret-free example. The local test suite invokes the actual stateless MCP HTTP handler with that JSON-RPC initialization pattern, rejects missing and invalid bearer credentials, and proves that a Viewer cannot invoke an Operator tool while an Operator can. These checks do not replace the deferred live connection test against the real AgentZero container and NodeScope internal CA.

## Reference

AgentZero documents URL-based MCP connections using a `url` and `Authorization` header in its MCP setup guide.[1]

[1]: https://github.com/agent0ai/agent-zero/blob/main/docs/guides/mcp-setup.md "AgentZero MCP Setup"
