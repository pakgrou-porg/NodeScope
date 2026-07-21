# NodeScope Secret-Handling Policy

NodeScope is a public repository. The repository, issue tracker, CI logs, test fixtures, screenshots, and documentation must never contain a usable credential, private key, full connection string, personally identifying inference payload, or backup content.

## Secret inventory

| Secret class | Used by | Approved storage | Rotation or revocation path | Repository rule |
|---|---|---|---|---|
| Supabase publishable URL/key | Browser console configuration | Deployment environment only | Replace deployment value and redeploy | Never commit a real project URL/key pair in examples. |
| Supabase service role / database password | Server migrations, controlled administration, backup/export | Server-replica secret store or local encrypted secret file | Supabase dashboard rotation and redeployment | Never expose to browser, agent, MCP, TUI, or CI logs. |
| Agent credential | One native agent | Root-readable local service configuration | Revoke/replace from administrator workflow | Store only a server-side keyed digest. |
| Client API key | Inference/proxy caller | Caller secret store | Revoke/rotate from administrator workflow | Reveal plaintext once; store only a keyed digest. |
| Internal CA root/intermediate private key | Certificate issuance and recovery | Offline, protected administrator-controlled medium | Documented dual-trust rotation | Never place in replica images, container layers, or GitHub Actions secrets unless an explicit future decision changes the PKI model. |
| Leaf private key | Framework/Asus service or agent | Protected local certificate directory | Short-lived renewal and replacement | Never commit. |
| GitHub OIDC identity | Release signing | Ephemeral GitHub Actions token | Per-run | Do not introduce a long-lived signing secret without approval. |
| Backup data | Daily export target | Shared protected local mount | Ten-copy pruning and recovery procedure | Do not include in source tree, test fixture, or support bundle. |

## Required safeguards

Every example uses placeholders such as `SUPABASE_URL`, `NODESCOPE_AGENT_TOKEN`, and `/srv/nodescope/secrets`. Runtime configuration is loaded from environment variables or root-readable files outside Git. Tests use synthetic non-secret fixtures.

Logging is structured and allowlist-based. Logs may include opaque request IDs, host IDs, metric counts, outcome codes, durations, and capacity states. Logs must not include HTTP request bodies, response bodies, authorization headers, cookie values, database URLs, backend error bodies, or raw prompt/response content.

Before every public push, CI must run secret scanning. A release is blocked if scans find committed credentials, private keys, tokens, or known high-entropy secrets. Any accidental disclosure is handled as a security incident: revoke or rotate the secret first, remove it from every deployment surface, then follow the repository security reporting process.
