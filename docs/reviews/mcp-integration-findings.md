# NodeScope MCP Integration Findings

NodeScope uses the official Go SDK for the Model Context Protocol. The SDK documents typed `mcp.AddTool` registration and streamable HTTP transport via `mcp.NewStreamableHTTPHandler`, including bounded request-body options and a stateless mode appropriate for the NodeScope remote HTTPS tool endpoint.[1] [2]

AgentZero documents external MCP setup through **Settings → MCP/A2A → External MCP Servers**. Its URL-based configuration accepts a remote `url` with an `Authorization: Bearer` header. AgentZero advises keeping real keys out of public files and notes that remote servers should use a reachable HTTPS URL from the AgentZero container/network.[3]

| Design decision | Source-backed implication |
|---|---|
| Remote NodeScope MCP endpoint | Use the official Streamable HTTP transport at `/mcp`, bounded to 1 MiB request bodies, with bearer authentication before tool invocation. |
| Role model | Map protected MCP client credentials to NodeScope Viewer, Operator, or Administrator roles; server-side tool authorization is mandatory. |
| AgentZero configuration | Provide a secret-free `mcp.json.example`; actual token belongs in protected AgentZero storage/environment. |
| Network and TLS | Framework is the preferred remote endpoint, Asus is the explicit fallback, and the AgentZero environment must trust the NodeScope internal CA. |

[1]: https://github.com/modelcontextprotocol/go-sdk "Official Model Context Protocol Go SDK"
[2]: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp "MCP Go SDK package documentation"
[3]: https://github.com/agent0ai/agent-zero/blob/main/docs/guides/mcp-setup.md "AgentZero MCP setup guide"
