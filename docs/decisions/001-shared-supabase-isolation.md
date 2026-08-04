# ADR 001: Shared Supabase project isolation

**Status:** Accepted

## Context

NodeScope and TTRPG-OCR will operate in one Supabase project to avoid a separate paid project. NodeScope must not read, write, alter, migrate, back up, or otherwise operate on TTRPG-OCR data or objects.

Supabase documents that custom schemas can be exposed through API settings, but exposure requires explicit schema and object grants. It also documents that PostgreSQL roles control system access, while RLS is intended for application access, and recommends a distinct database user for each service rather than sharing the `postgres` password.[1] [2]

## Decision

NodeScope owns the dedicated `nodescope` schema only. Its database object owner is `nodescope_owner`; its deployed runtime assumes `nodescope_runtime`; its migration login explicitly assumes `nodescope_owner`; and both login credentials are distinct from the shared `postgres` account.

The shared Supabase `anon`, `authenticated`, and `service_role` roles receive no direct NodeScope schema usage. NodeScope validates shared Supabase JWTs at its own server boundary, then uses its dedicated runtime database role. This protects the `nodescope` data plane from a TTRPG-OCR browser session that merely shares the same Supabase Auth project.

NodeScope stores Auth UUIDs as opaque identifiers and intentionally has no foreign keys, schema grants, or read permission against `auth.users`. Every NodeScope migration is schema-qualified and has a rollback-only preflight. The migration verifier checks RLS, ownership, generic-role denial, sibling schema DDL denial, and sibling table privilege denial. Future migration rehearsals must run through the dedicated migrator login.

## Consequences

The Supabase Auth configuration, redirect allowlist, email delivery configuration, quotas, backups, and project pause behavior remain shared operational concerns. NodeScope’s product configuration and all business data remain isolated. Any TTRPG-OCR schema should be explicit rather than stored in `public`; `public` retains unavoidable baseline PostgreSQL usage but carries no NodeScope table, routine, or DDL privilege.

## References

[1] [Supabase: Using Custom Schemas](https://supabase.com/docs/guides/api/using-custom-schemas)

[2] [Supabase: Postgres Roles](https://supabase.com/docs/guides/database/postgres/roles)

[3] [Supabase: Row Level Security](https://supabase.com/docs/guides/database/postgres/row-level-security)
