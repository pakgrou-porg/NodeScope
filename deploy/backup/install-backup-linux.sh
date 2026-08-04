#!/usr/bin/env bash
# Install the NodeScope fenced-backup helper from verified release inputs.
set -euo pipefail
umask 077
if [[ "${EUID}" -ne 0 ]]; then echo "Run with sudo or as root." >&2; exit 1; fi
if [[ $# -ne 6 ]]; then
  echo "Usage: $0 <backup-binary> <binary-sha256> <service-unit> <service-sha256> <timer-unit> <timer-sha256>" >&2
  exit 2
fi
binary="$1"; binary_sha="$2"; service="$3"; service_sha="$4"; timer="$5"; timer_sha="$6"
for checksum in "$binary_sha" "$service_sha" "$timer_sha"; do [[ "$checksum" =~ ^[a-fA-F0-9]{64}$ ]] || { echo "Expected SHA-256 values are required." >&2; exit 1; }; done
for input in "$binary" "$service" "$timer"; do [[ -f "$input" && ! -L "$input" ]] || { echo "Input must be a regular non-symlink file: $input" >&2; exit 1; }; done
stage=/var/lib/nodescope-installer/backup-stage
install -d -o root -g root -m 0700 "$stage" /var/backups/nodescope /etc/nodescope-backup
for pair in "binary:$binary:$binary_sha" "service:$service:$service_sha" "timer:$timer:$timer_sha"; do
  IFS=: read -r kind source expected <<<"$pair"
  target="$stage/$kind"
  install -o root -g root -m 0700 "$source" "$target"
  actual="$(sha256sum "$target" | awk '{print $1}')"
  [[ "${actual,,}" == "${expected,,}" ]] || { echo "Checksum mismatch for $kind." >&2; exit 1; }
done
install -o root -g root -m 0755 "$stage/binary" /usr/local/bin/nodescope-backup
install -o root -g root -m 0644 "$stage/service" /etc/systemd/system/nodescope-backup.service
install -o root -g root -m 0644 "$stage/timer" /etc/systemd/system/nodescope-backup.timer
if [[ ! -f /etc/nodescope-backup/backup.conf ]]; then
  cat >/etc/nodescope-backup/backup.conf <<'EOF'
# Root-owned, non-secret configuration. The password is stored only in /etc/nodescope-backup/pgpass.
# The same NODESCOPE_BACKUP_DIRECTORY must be mounted on Framework and Asus for safe backup failover.
# NODESCOPE_REPLICA_ID=framework
# NODESCOPE_BACKUP_DIRECTORY=/var/backups/nodescope
# NODESCOPE_BACKUP_MODE=default
# NODESCOPE_BACKUP_PGHOST=db.example.internal
# NODESCOPE_BACKUP_PGPORT=5432
# NODESCOPE_BACKUP_PGDATABASE=postgres
# NODESCOPE_BACKUP_PGUSER=nodescope_backup_login
EOF
  chmod 0600 /etc/nodescope-backup/backup.conf
fi
printf 'Create /etc/nodescope-backup/pgpass with mode 0600, mount the same backup target on both replicas, and add a systemd drop-in if the backup target is not /var/backups/nodescope. The fenced lease prevents concurrent publication.\n'
systemctl daemon-reload
systemd-analyze verify /etc/systemd/system/nodescope-backup.service /etc/systemd/system/nodescope-backup.timer
systemctl enable nodescope-backup.timer
