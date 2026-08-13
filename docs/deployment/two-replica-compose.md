# Two-replica NodeScope deployment

NodeScope Release 1 runs the complete server stack independently on Framework and Asus. Framework is the **preferred** replica at `10.116.2.145`; Asus is the **secondary** replica at `10.116.2.56`. Both use the same shared Supabase project but only the `nodescope` schema and the dedicated NodeScope runtime credential.

## Deployment procedure

Copy `deploy/compose/replica.env.example` to a protected `replica.env` file on each host. Set the role and ID for that host. `NODESCOPE_PRIMARY_ENDPOINT` and `NODESCOPE_SECONDARY_ENDPOINT` must be distinct credential-free HTTPS base URLs; NodeScope rejects embedded user information, query strings, fragments, and duplicate canonical destinations so configured failover cannot silently point to the same replica. Create `/srv/nodescope/certs` and `/srv/nodescope/runtime` with root-owned restrictive permissions. The certificate directory contains the issued server certificate and key. Create `/srv/nodescope/runtime/runtime.env` with `NODESCOPE_RUNTIME_DB_PASSWORD` only; the Compose stack combines it with the non-secret host/user fields in `replica.env` and constructs a percent-encoded runtime connection URL in memory. Neither file is committed.

Deploy the same Compose definition through Portainer Stack or Docker Compose. Operators should deploy the secondary replica first, verify `/healthz`, `/readyz`, and peer connectivity, then deploy the preferred Framework replica. Each inference client and native agent uses Framework as its preferred endpoint and Asus as its application-level fallback.

The Compose file intentionally does not contain any password, service-role key, agent token, client token, CA key, or Supabase API secret. It uses a static non-root image, read-only root filesystem, dropped Linux capabilities, a bounded temporary filesystem, and a local health probe.

## Shared Supabase constraint

The database must be prepared using `supabase/migrations/0001_nodescope_foundation.sql` and then verified with `supabase/isolation/verify_shared_isolation.sql`. The verifier must pass before a replica receives a NodeScope runtime credential. Runtime access is limited to the `nodescope` schema; it must not be configured with the shared `postgres` password or Supabase service-role key.

## Validation

After each deploy, confirm the health and readiness endpoints return HTTP 200. Confirm the console shows the correct preferred/secondary replica roles, and ensure the secondary has the required shared backup target mounted before enabling backup takeover. Certificate paths are expected to be valid and readable only by the process that needs them.
