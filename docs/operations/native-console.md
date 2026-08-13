# Native CLI and TUI Operations

NodeScope provides native, read-only terminal views for the fleet. The `nodescope-cli` command writes **table**, **JSON**, or **NDJSON** output for SSH-friendly automation. The `nodescope-tui` command renders the same fleet status interactively and refreshes from one to sixty seconds. Both clients deliberately retrieve only fleet metadata: host identity, platform, freshness, server receipt time, and metric-quality counts. They never request, display, or persist inference prompts or responses.

## Read paths

The console selects one read path per invocation. HTTPS is preferred when an endpoint is supplied; otherwise an SSH target is used; otherwise the client uses the local verifier database URL. This ordering lets a workstation use the role-checked server API without distributing database credentials.

| Mode | Required configuration | Use case | Trust boundary |
|---|---|---|---|
| Authenticated HTTPS | `--endpoint` and `--credential-file` | A terminal client outside the server host | The server validates a Viewer-or-higher credential and HTTPS trust. |
| Local SSH relay | `--ssh-target` | A workstation that can reach a host with `nodescope-cli` installed | The local SSH client authenticates to the host; the remote CLI supplies metadata-only JSON. |
| Local verifier database | `NODESCOPE_VERIFIER_DATABASE_URL` | An administrator shell on a controlled server host | The read-only verifier database identity is used directly. |

The HTTPS endpoint must be a plain `https://host` base URL without embedded user information, query parameters, fragments, or bearer tokens. NodeScope reads the credential only from the designated file and rejects redirects. For an internal certificate authority, pass `--ca-file` with the CA certificate in PEM format. If no CA file is supplied, the operating system trust store is used.

## HTTPS examples

Place a role-scoped API credential in a root-owned or user-owned file with restrictive permissions. Do not put it on a command line, in a shell history, or in a URL.

```bash
chmod 600 "$HOME/.config/nodescope/viewer.credential"

nodescope-cli \
  --endpoint https://nodescope.framework.lan \
  --credential-file "$HOME/.config/nodescope/viewer.credential" \
  --ca-file /etc/nodescope/pki/root-ca.pem \
  --format ndjson

nodescope-tui \
  --endpoint https://nodescope.framework.lan \
  --credential-file "$HOME/.config/nodescope/viewer.credential" \
  --ca-file /etc/nodescope/pki/root-ca.pem \
  --refresh 5s
```

The equivalent environment variables are `NODESCOPE_CONTROL_API_URL`, `NODESCOPE_CONTROL_API_CREDENTIAL_FILE`, and `NODESCOPE_CONTROL_API_CA_FILE`. Flags override those variables. The HTTPS credential grants only the authority assigned by NodeScope; a Viewer credential is sufficient for the fleet view.

## SSH relay examples

The SSH relay mode runs a read-only JSON request on the remote host. It does not copy the invoking machine's API credential to that host. Ensure the remote `nodescope-cli` has a separately configured verifier database or HTTPS read path appropriate for that host.

```bash
nodescope-cli --ssh-target framework.lan --format table
nodescope-tui --ssh-target framework.lan --refresh 10s
```

Use `--ssh-command` only when the remote CLI is installed at a non-default path or must be started through an approved wrapper. The wrapper must append `--format json` unchanged and must not print banners or diagnostic text to standard output, because the relay expects a single JSON array.

## Local verifier database mode

On a controlled server host, provide the narrow verifier connection string through an environment variable rather than a command line.

```bash
export NODESCOPE_VERIFIER_DATABASE_URL='postgresql://…'
nodescope-cli --format json
nodescope-tui --refresh 5s
```

This mode is intentionally a fallback for local administration. It must use the dedicated verifier role or a comparably narrow read-only interface; do not reuse a migrator or owner credential.

## Failure behavior

The clients do not synthesize status if a configured path fails. The CLI exits nonzero and reports a bounded error. The TUI displays **FLEET STATUS UNAVAILABLE** and explicitly states that no substitute values are shown. A remote API response lacks the database-only collection interval and clock-offset fields, so the terminal clients display those values as `n/a` or `—`, rather than fabricating a zero value.
