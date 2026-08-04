#!/usr/bin/env bash
# Install the verified NodeScope update helper and systemd units from checked inputs.
set -euo pipefail
umask 077

if [[ "${EUID}" -ne 0 ]]; then echo "Run with sudo or as root." >&2; exit 1; fi
if [[ $# -ne 6 ]]; then
  echo "Usage: $0 <update-binary> <update-sha256> <service-unit> <service-sha256> <timer-unit> <timer-sha256>" >&2
  exit 2
fi

update_binary="$1"; update_sha="$2"; service_unit="$3"; service_sha="$4"; timer_unit="$5"; timer_sha="$6"
for value in "$update_sha" "$service_sha" "$timer_sha"; do [[ "$value" =~ ^[a-fA-F0-9]{64}$ ]] || { echo "Expected SHA-256 values are required." >&2; exit 1; }; done
for input in "$update_binary" "$service_unit" "$timer_unit"; do [[ -f "$input" && ! -L "$input" ]] || { echo "Input must be a regular non-symlink file: $input" >&2; exit 1; }; done

stage=/var/lib/nodescope-installer/update-stage
install -d -o root -g root -m 0700 "$stage" /var/lib/nodescope/update
for pair in "binary:$update_binary:$update_sha" "service:$service_unit:$service_sha" "timer:$timer_unit:$timer_sha"; do
  IFS=: read -r kind source expected <<<"$pair"
  target="$stage/$kind"
  install -o root -g root -m 0700 "$source" "$target"
  actual="$(sha256sum "$target" | awk '{print $1}')"
  [[ "${actual,,}" == "${expected,,}" ]] || { echo "Checksum mismatch for $kind." >&2; exit 1; }
done

install -o root -g root -m 0755 "$stage/binary" /usr/local/bin/nodescope-update
install -o root -g root -m 0644 "$stage/service" /etc/systemd/system/nodescope-update.service
install -o root -g root -m 0644 "$stage/timer" /etc/systemd/system/nodescope-update.timer
if [[ ! -f /etc/nodescope-agent/update.env ]]; then
  cat >/etc/nodescope-agent/update.env <<'EOF'
# Approved pinned release manifest. Update these only through the NodeScope
# Administrator workflow after verifying the GitHub release provenance.
# NODESCOPE_UPDATE_VERSION=v0.1.0
# NODESCOPE_UPDATE_ARCHIVE_URL=https://github.com/pakgrou-porg/NodeScope/releases/download/v0.1.0/nodescope_0.1.0_linux_amd64.tar.gz
# NODESCOPE_UPDATE_CHECKSUM_URL=https://github.com/pakgrou-porg/NodeScope/releases/download/v0.1.0/nodescope_0.1.0_linux_amd64.tar.gz.sha256
EOF
  chmod 0600 /etc/nodescope-agent/update.env
fi
systemctl daemon-reload
systemd-analyze verify /etc/systemd/system/nodescope-update.service /etc/systemd/system/nodescope-update.timer
systemctl enable nodescope-update.timer
printf 'Installed the NodeScope verified-update helper. Set the approved pinned manifest in /etc/nodescope-agent/update.env and validate gh attestation access before enabling unattended updates.\n'
