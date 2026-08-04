#!/usr/bin/env bash
# Install a verified NodeScope Linux agent binary and hardened systemd service.
# This script deliberately does not install GPU/NPU/container dependencies.
set -euo pipefail
umask 077

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run this installer with sudo or as root." >&2
  exit 1
fi

if [[ $# -ne 4 ]]; then
  cat >&2 <<'USAGE'
Usage:
  install-linux.sh <agent-binary> <agent-sha256> <unit-file> <unit-sha256>

Use checksums from the approved signed NodeScope release manifest. The installer
rejects symlinks, stages both inputs under root ownership, verifies hashes after
staging, preserves the previous binary, and atomically replaces the live files.
USAGE
  exit 2
fi

binary_source="$1"
expected_binary_sha="$2"
unit_source="$3"
expected_unit_sha="$4"

validate_source() {
  local path="$1"
  local label="$2"
  if [[ ! -f "$path" || -L "$path" ]]; then
    echo "$label must be a regular, non-symlink file: $path" >&2
    exit 1
  fi
}
validate_source "$binary_source" "Agent binary"
validate_source "$unit_source" "Systemd unit"

if ! [[ "$expected_binary_sha" =~ ^[a-fA-F0-9]{64}$ && "$expected_unit_sha" =~ ^[a-fA-F0-9]{64}$ ]]; then
  echo "Expected SHA-256 values must be 64 hexadecimal characters." >&2
  exit 1
fi

systemd_version="$(systemctl --version | awk 'NR==1 {print $2}')"
if [[ ! "$systemd_version" =~ ^[0-9]+$ || "$systemd_version" -lt 250 ]]; then
  echo "NodeScope requires systemd 250 or later for per-service credentials; found ${systemd_version:-unknown}." >&2
  exit 1
fi

stage_root=/var/lib/nodescope-installer/stage
backup_root=/var/lib/nodescope-installer/backups
install -d -o root -g root -m 0700 "$stage_root" "$backup_root" /etc/nodescope-agent/credentials
stage_binary="$(mktemp "$stage_root/agent.XXXXXX")"
stage_unit="$(mktemp "$stage_root/unit.XXXXXX")"
cleanup() { rm -f "$stage_binary" "$stage_unit" "${binary_destination_tmp:-}" "${unit_destination_tmp:-}"; }
trap cleanup EXIT

# Copy first, then hash the root-owned staging copies to defeat source-path races.
cat -- "$binary_source" >"$stage_binary"
cat -- "$unit_source" >"$stage_unit"
chmod 0700 "$stage_binary"
chmod 0600 "$stage_unit"

actual_binary_sha="$(sha256sum "$stage_binary" | awk '{print $1}')"
actual_unit_sha="$(sha256sum "$stage_unit" | awk '{print $1}')"
if [[ "${actual_binary_sha,,}" != "${expected_binary_sha,,}" ]]; then
  echo "Agent binary SHA-256 mismatch; refusing installation." >&2
  exit 1
fi
if [[ "${actual_unit_sha,,}" != "${expected_unit_sha,,}" ]]; then
  echo "Systemd unit SHA-256 mismatch; refusing installation." >&2
  exit 1
fi

if ! id nodescope >/dev/null 2>&1; then
  useradd --system --home-dir /var/lib/nodescope-agent --shell /usr/sbin/nologin nodescope
fi
install -d -o nodescope -g nodescope -m 0700 /var/lib/nodescope-agent
install -d -o root -g nodescope -m 0750 /etc/nodescope-agent
install -d -o root -g root -m 0700 /etc/nodescope-agent/credentials

if [[ -e /usr/local/bin/nodescope-agent ]]; then
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  backup_binary="$backup_root/nodescope-agent.${timestamp}"
  cp --preserve=mode,ownership,timestamps /usr/local/bin/nodescope-agent "$backup_binary"
  sha256sum "$backup_binary" >"${backup_binary}.sha256"
fi

binary_destination_tmp="$(mktemp /usr/local/bin/.nodescope-agent.XXXXXX)"
unit_destination_tmp="$(mktemp /etc/systemd/system/.nodescope-agent.service.XXXXXX)"
install -o root -g root -m 0755 "$stage_binary" "$binary_destination_tmp"
install -o root -g root -m 0644 "$stage_unit" "$unit_destination_tmp"
mv -f "$binary_destination_tmp" /usr/local/bin/nodescope-agent
binary_destination_tmp=""
mv -f "$unit_destination_tmp" /etc/systemd/system/nodescope-agent.service
unit_destination_tmp=""

if [[ ! -f /etc/nodescope-agent/agent.env ]]; then
  cat >/etc/nodescope-agent/agent.env <<'EOF'
# Non-secret NodeScope agent configuration. The bearer token is provided only
# through systemd LoadCredential= and must never be added to this file.
# NODESCOPE_AGENT_ID=
# NODESCOPE_HOST_ID=
# NODESCOPE_PRIMARY_ENDPOINT=https://10.116.2.145:8443
# NODESCOPE_SECONDARY_ENDPOINT=https://10.116.2.56:8443
# NODESCOPE_COLLECTION_INTERVAL_SECONDS=5
# NODESCOPE_CA_CERT_PATH=/etc/nodescope-agent/ca.pem
# NODESCOPE_SELECTED_PROCESS_NAMES=
# NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES=
EOF
  chown root:nodescope /etc/nodescope-agent/agent.env
  chmod 0640 /etc/nodescope-agent/agent.env
fi

systemctl daemon-reload
systemd-analyze verify /etc/systemd/system/nodescope-agent.service
systemctl enable nodescope-agent.service
printf 'NodeScope agent installed from verified inputs. Enroll the host with nodescope-enroll to create /etc/nodescope-agent/credentials/agent-token, configure non-secret agent.env, then run systemd-analyze security nodescope-agent.service before starting the service.\n'
