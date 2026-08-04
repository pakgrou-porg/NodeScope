# NodeScope shared-Supabase bootstrap

NodeScope and TTRPG-OCR operate in one Supabase project while retaining independent application data planes. NodeScope owns only the **`nodescope` PostgreSQL schema**, its schema-local tables, routines, types, migration history, and dedicated database roles. TTRPG-OCR schemas, tables, functions, storage, and configuration are out of scope for every NodeScope migration, credential, backup, and operational command.

> **Do not use a tablespace for application isolation.** In managed Supabase, the practical isolation boundaries are schemas, object ownership, least-privilege database roles, grants, RLS, schema-qualified migrations, and independent credential rotation.

## Required role model

The first NodeScope migration creates three non-login group roles: `nodescope_owner`, `nodescope_migrator`, and `nodescope_runtime`. The bootstrap operator must create two credentialed login roles outside Git:

| Login role | Membership | Intended use | Explicit restriction |
|---|---|---|---|
| `nodescope_migrate_login` | `nodescope_migrator` | Source-controlled NodeScope migrations only | No use by the browser, agent, proxy, TUI, MCP, or AgentZero. |
| `nodescope_runtime_login` | `nodescope_runtime` | Replicated NodeScope server only | No DDL, no ownership, and no privilege on TTRPG-OCR objects. |

The `postgres` project password is a break-glass administrative credential, not a NodeScope application credential. It must not be used by the deployed NodeScope replicas after bootstrap.

## Required configuration before browser access

Apply source-controlled NodeScope migrations in numerical order with the dedicated migration login and a schema-qualified `search_path=nodescope,public`. Verify that every created object is in the `nodescope` schema and owned by `nodescope_owner` before moving on.

Supabase Auth is shared at the project level. Disable public self-signup, enable passwordless magic links, add Framework as the deterministic default callback, and add Asus only as the emergency callback. Those project-level decisions must remain compatible with TTRPG-OCR because they affect the shared Auth service. The NodeScope bootstrap then invites the initial NodeScope Administrator and creates its `nodescope.user_roles` record.

The server may validate Supabase JWTs using the shared project’s publishable key/JWKS. It must connect to PostgreSQL only as `nodescope_runtime_login`; neither the service-role API key nor the `postgres` password belongs in browser, agent, proxy, MCP, CLI, or AgentZero configuration.

## Isolation verification

Before enabling real ingestion, run the `supabase/isolation` verification suite. It must prove that the NodeScope runtime login can operate on `nodescope.*` but receives permission errors for TTRPG-OCR schemas and objects. It must also prove that NodeScope migrations produce no DDL outside `nodescope`, `auth` foreign-key references, and required Supabase system catalog access.

The test suite must additionally prove that an unauthenticated caller cannot read NodeScope data; a Viewer cannot change configuration; an Operator cannot manage users, credentials, or runtime approvals; an Administrator can carry out the documented administration workflows; and direct browser writes cannot create audit or operation records outside the server transaction protocol.

## Migration discipline

Every NodeScope migration is additive, source-controlled, schema-qualified, and reviewed against a disposable shared-project fixture before application. Do not edit an applied migration. Do not execute blanket `GRANT`, `REVOKE`, `ALTER DEFAULT PRIVILEGES`, backup, restore, or schema-discovery commands outside the `nodescope` schema.

RLS remains defense in depth within `nodescope`. The NodeScope server is still the authorization authority for Viewer, Operator, and Administrator semantics. A shared Supabase service role bypasses RLS, so it is intentionally excluded from deployed NodeScope runtime paths.
