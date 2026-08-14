# Least-Privilege Agent Enrollment and Credential Rotation

NodeScope enrolls a native agent through the dedicated `nodescope_enroller` role and the schema-local `nodescope.enroll_or_rotate_agent` function. The function preserves canonical host identity by stable host slug, replaces only the credential digest, increments the agent rotation version, and records audit metadata without storing or returning the raw credential.

> Do not use `postgres`, `nodescope_owner`, the runtime database login, browser environment variables, or an agent-side database connection to enroll an agent. Native agents authenticate to server replicas over HTTPS and do not receive database credentials.

## Prerequisites

Migration `0014_least_privilege_agent_operations` must be applied, and the shared-project bootstrap must have created `nodescope_enroller`. Provide its dedicated database URL only through a protected administrator environment; never paste it into terminal history, Git, Portainer environment values, or an agent configuration file.

## Enroll Framework

From an administrator shell with `NODESCOPE_ENROLLER_DATABASE_URL` sourced from protected storage, generate and stage a Framework credential file:

```bash
sudo install -d -o root -g root -m 0700 /etc/nodescope/credentials
sudo -E ./scripts/enroll-or-rotate-agent.sh \
  --slug framework \
  --display-name Framework \
  --platform fedora-linux-amd64 \
  --address 10.116.2.145 \
  --credential-output /etc/nodescope/credentials/framework-agent.token \
  --expires-days 90
```

The command prints the host slug, credential-file path, and rotation version only. It writes a new random credential to the protected file with mode `0600`; install that file through the native-agent secret-file or systemd-credential path. The raw credential must not enter logs, database audit metadata, shell history, browser UI, MCP output, or any source-controlled configuration.

## Rotate a credential

Run the same command with the same stable `--slug` and a **new credential-output path**. The database host ID stays canonical, the agent ID remains associated with that host, the prior digest stops authenticating immediately, and the audit event records only rotation metadata and the short credential hint. Update the agent’s protected credential file, restart or reload the agent, and verify the authenticated non-mutating ingestion preflight against the preferred replica before removing the old protected file.

## Verification and recovery

Verify identity through `GET /api/v1/ingest/preflight` using the native agent, not direct database access. The response must report the expected agent and host identity without echoing the credential. If rotation fails after the database function accepts a new digest but before the agent receives the protected file, run the wrapper again to create a replacement credential; do not restore a previous digest manually. If a credential is suspected exposed, rotate it immediately, revoke the affected certificate independently if mTLS is enabled, and retain only redacted audit identifiers and rotation versions for incident review.
