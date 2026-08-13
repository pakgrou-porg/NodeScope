# Fenced backups and restore rehearsal

NodeScope creates a daily **default** backup containing the `nodescope` schema plus configuration and summary telemetry. It excludes high-volume `raw_metric_samples` and `ingest_receipts`. A **full** backup includes every NodeScope table currently retained. The default retention is ten daily snapshots.

The Framework and Asus replicas both attempt the scheduled job. A database-time fenced lease named `daily_backup` authorizes only one replica to publish a snapshot. The publisher checks its fencing token before archive creation and again immediately before the final atomic rename. If it loses the lease while creating the archive, it removes the partial artifact and refuses publication, preventing a stale or partitioned replica from presenting a backup as current. Archive partial files are created exclusively: a pre-existing regular file or symlink at that path is not overwritten or followed. Staging archives accept regular files only; a symlink, device, socket, FIFO, or other non-regular staged entry aborts the run before it can be read. Treat either condition as an operator investigation item rather than deleting it automatically.

## Shared target requirement

The configured backup directory must be the same mounted storage target on both replicas and must appear at the **same absolute path**. The default is `/var/backups/nodescope`. If a different mount path is used, create a systemd drop-in that grants `nodescope-backup.service` write access to that exact path and test it on both hosts before enabling the timer.

```bash
sudo systemctl edit nodescope-backup.service
```

```ini
[Service]
ReadWritePaths=/mnt/nodescope-backups
```

The directory must be owned by root, mode `0750` or stricter, and have sufficient capacity for ten default snapshots plus one in-progress file. NodeScope does not encrypt backups by current design; filesystem access control is therefore mandatory.

## Backup password file

The backup command uses `pg_dump` with a root-only PostgreSQL password file, never a password-bearing command line. Create `/etc/nodescope-backup/pgpass` with mode `0600` and the PostgreSQL password-file format:

```text
host:port:database:user:password
```

The backup database login must be a dedicated NodeScope read-only role with access only to the `nodescope` schema. Do not use the migrator, runtime, Supabase service-role, or TTRPG-OCR credentials.

## Restore rehearsal

Before relying on scheduled backups, perform a restore rehearsal in a disposable database. Verify that the archive contains a `manifest.json`, the expected fenced token, and a custom-format `nodescope.dump`. Restore only into a non-production schema/database, run the NodeScope read-only verifier, and record the outcome in the audit log. Do not restore into the shared production Supabase project.

## Deferred production gate

The backup code, lease functions, installer, and systemd units are implemented locally. Enabling the migration, creating the dedicated backup login, mounting the shared target, and completing the restore rehearsal remain external acceptance gates.
