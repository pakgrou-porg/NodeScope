# NodeScope Migration Application Guard

> **Activation gate.** This procedure is not a substitute for the required shared-Supabase isolation rehearsal, sibling-schema denial checks, or administrator approval.

`scripts/apply-nodescope-migration.sh` accepts exactly one reviewed migration path and performs no database action until its local path-containment checks pass. The argument must resolve to a clean, tracked, regular SQL file directly under `supabase/migrations/` with the `NNNN_name.sql` naming convention. Traversal paths, nested paths, missing files, symlinks, untracked files, and locally modified migration files are rejected before database credentials are consulted.

The script then requires the dedicated NodeScope migrator credentials, verifies both NodeScope logins remain denied from the disposable sibling fixture, runs the requested SQL inside a rolled-back preflight transaction, applies the migration only after those gates pass, and runs the post-apply shared-project isolation verification through that same dedicated migrator connection. It does not reuse an opaque database URL credential for verification. Do not invoke `psql` directly for production migrations.

The local `scripts/test-apply-nodescope-migration-contract.sh` check is credential-free. It proves that traversal, symlink, and untracked SQL inputs fail before any live database gate. The full release-readiness suite executes this check automatically.
