# Drizzle Scaffold Audit Record

**Scope:** reported unused-ORM review finding. **Environment:** current NodeScope source tree and authenticated-console code path. **Decision:** retain the Drizzle dependencies and schema scaffold.

> The audit did not treat a package declaration as proof of runtime use. It traced actual imports from the OAuth and session services into the user persistence helper before deciding whether cleanup was safe.

| Finding | Evidence | Decision |
| --- | --- | --- |
| `drizzle-orm`, `drizzle-kit`, `drizzle.config.ts`, and `drizzle/` are present. | Package metadata and tracked scaffold files declare the tooling. | **Retained pending authentication migration.** |
| Browser/session identity path is active. | `server/_core/oauth.ts` uses `db.upsertUser`; `server/_core/sdk.ts` uses `db.getUserByOpenId` and `db.upsertUser`. | **Retained.** Removing Drizzle now would break the current authenticated console session and user synchronization path. |
| NodeScope telemetry data path uses Go and the dedicated Supabase schema. | The current NodeScope control-plane telemetry code does not depend on Drizzle. | **Not an authorization to remove the scaffold.** The browser identity path remains separate and active. |

## Known limitation and future recovery

The current browser authentication bridge is template-derived and distinct from the planned environment-validated Supabase magic-link flow. Drizzle can be removed only as part of a dedicated, tested authentication migration that replaces `server/db.ts`, `drizzle/schema.ts`, and all active OAuth/session callers atomically. That migration must include a real invite-only browser E2E run, role migration plan, rollback to the retained user store, and a clean-clone readiness result.

No runtime or dependency was removed in this audit. The recovery path is therefore no-op: retain the current pinned dependency set until the separate authentication migration is explicitly authorized and verified.
