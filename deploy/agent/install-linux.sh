#!/usr/bin/env bash
# Install a verified NodeScope Linux agent binary and hardened systemd service.
# This script deliberately does not install GPU/NPU/container dependencies.
set -euo pipefail
umask 077

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run this installer with sudo or as root." >&2
  exit 1
fi

if [[ $# -ne 6 ]]; then
  cat >&2 <<'USAGE'
Usage:
  install-linux.sh <agent-binary> <agent-sha256> <unit-file> <unit-sha256> <release-tag> <source-revision>

Use checksums, a pinned release tag, and immutable source revision from the
approved signed NodeScope release evidence. The installer rejects symlinks,
stages inputs under root ownership, verifies hashes after staging, preserves
prior binary/unit/provenance records, and atomically replaces the live files.
USAGE
  exit 2
fi

binary_source="$1"
expected_binary_sha="$2"
unit_source="$3"
expected_unit_sha="$4"
release_tag="$5"
source_revision="$6"
install_root="${NODESCOPE_INSTALL_ROOT:-}"
systemctl_bin="${NODESCOPE_SYSTEMCTL_BIN:-systemctl}"
systemd_analyze_bin="${NODESCOPE_SYSTEMD_ANALYZE_BIN:-systemd-analyze}"

root_path() {
  printf '%s%s' "$install_root" "$1"
}

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
if ! [[ "$release_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z]+)*$ ]]; then
  echo "Release tag must be a pinned vMAJOR.MINOR.PATCH identifier." >&2
  exit 1
fi
if ! [[ "$source_revision" =~ ^[a-fA-F0-9]{40,64}$ ]]; then
  echo "Source revision must be an immutable 40- or 64-character hexadecimal revision." >&2
  exit 1
fi

systemd_version="$($systemctl_bin --version | awk 'NR==1 {print $2}')"
if [[ ! "$systemd_version" =~ ^[0-9]+$ || "$systemd_version" -lt 250 ]]; then
  echo "NodeScope requires systemd 250 or later for per-service credentials; found ${systemd_version:-unknown}." >&2
  exit 1
fi

stage_root="$(root_path /var/lib/nodescope-installer/stage)"
backup_root="$(root_path /var/lib/nodescope-installer/backups)"
metadata_root="$(root_path /var/lib/nodescope-installer/metadata)"
metadata_path="$metadata_root/installed.env"
agent_binary_path="$(root_path /usr/local/bin/nodescope-agent)"
agent_binary_dir="$(dirname "$agent_binary_path")"
unit_path="$(root_path /etc/systemd/system/nodescope-agent.service)"
unit_dir="$(dirname "$unit_path")"
config_dir="$(root_path /etc/nodescope-agent)"
credential_dir="$config_dir/credentials"
state_dir="$(root_path /var/lib/nodescope-agent)"
agent_env="$config_dir/agent.env"

install -d -o root -g root -m 0700 "$stage_root" "$backup_root" "$metadata_root" "$credential_dir"
install -d -o root -g root -m 0755 "$agent_binary_dir" "$unit_dir"
stage_binary="$(mktemp "$stage_root/agent.XXXXXX")"
stage_unit="$(mktemp "$stage_root/unit.XXXXXX")"
cleanup() { rm -f "$stage_binary" "$stage_unit" "${metadata_destination_tmp:-}" "${binary_destination_tmp:-}" "${unit_destination_tmp:-}"; }
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

if [[ -z "$install_root" ]]; then
  if ! id nodescope >/dev/null 2>&1; then
    useradd --system --home-dir /var/lib/nodescope-agent --shell /usr/sbin/nologin nodescope
  fi
  install -d -o nodescope -g nodescope -m 0700 "$state_dir"
  install -d -o root -g nodescope -m 0750 "$config_dir"
else
  # Test roots never create or alter host accounts.
  install -d -o root -g root -m 0700 "$state_dir"
  install -d -o root -g root -m 0750 "$config_dir"
fi

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_binary=""
backup_unit=""
backup_metadata=""
if [[ -e "$agent_binary_path" ]]; then
  backup_binary="$backup_root/nodescope-agent.${timestamp}"
  cp --preserve=mode,ownership,timestamps "$agent_binary_path" "$backup_binary"
  sha256sum "$backup_binary" >"${backup_binary}.sha256"
fi
if [[ -e "$unit_path" ]]; then
  backup_unit="$backup_root/nodescope-agent.service.${timestamp}"
  cp --preserve=mode,ownership,timestamps "$unit_path" "$backup_unit"
  sha256sum "$backup_unit" >"${backup_unit}.sha256"
fi
if [[ -e "$metadata_path" ]]; then
  backup_metadata="$backup_root/installed.env.${timestamp}"
  cp --preserve=mode,ownership,timestamps "$metadata_path" "$backup_metadata"
fi

binary_destination_tmp="$(mktemp "$agent_binary_dir/.nodescope-agent.XXXXXX")"
unit_destination_tmp="$(mktemp "$unit_dir/.nodescope-agent.service.XXXXXX")"
install -o root -g root -m 0755 "$stage_binary" "$binary_destination_tmp"
install -o root -g root -m 0644 "$stage_unit" "$unit_destination_tmp"
mv -f "$binary_destination_tmp" "$agent_binary_path"
binary_destination_tmp=""
mv -f "$unit_destination_tmp" "$unit_path"
unit_destination_tmp=""

metadata_destination_tmp="$(mktemp "$metadata_root/.installed.env.XXXXXX")"
cat >"$metadata_destination_tmp" <<EOF
# Root-owned NodeScope installation provenance. Do not edit manually.
schema_version=1
installed_at_utc=$timestamp
release_tag=$release_tag
source_revision=${source_revision,,}
agent_binary_sha256=${actual_binary_sha,,}
unit_sha256=${actual_unit_sha,,}
previous_binary_backup=$backup_binary
previous_unit_backup=$backup_unit
previous_metadata_backup=$backup_metadata
EOF
chmod 0600 "$metadata_destination_tmp"
mv -f "$metadata_destination_tmp" "$metadata_path"
metadata_destination_tmp=""

if [[ ! -f "$agent_env" ]]; then
  cat >"$agent_env" <<'EOF'
# Non-secret NodeScope agent configuration. The bearer token is supplied only
# through the root-managed credential file and must never be added here.
# NODESCOPE_AGENT_ID=
# NODESCOPE_HOST_ID=
# NODESCOPE_AGENT_CREDENTIAL_FILE=/etc/nodescope-agent/credentials/agent-token
# NODESCOPE_PRIMARY_ENDPOINT=https://framework.nodescope.lan:8443
# NODESCOPE_SECONDARY_ENDPOINT=https://asus.nodescope.lan:8443
# NODESCOPE_COLLECTION_INTERVAL_SECONDS=5
# NODESCOPE_CA_CERT_PATH=/etc/nodescope-agent/ca.pem
# NODESCOPE_REQUIRE_CLIENT_MTLS=false
# NODESCOPE_TLS_CLIENT_CERT_PATH=
# NODESCOPE_TLS_CLIENT_KEY_PATH=
# NODESCOPE_SELECTED_PROCESS_NAMES=
# NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES=
EOF
  if [[ -z "$install_root" ]]; then
    chown root:nodescope "$agent_env"
  else
    chown root:root "$agent_env"
  fi
  chmod 0640 "$agent_env"
fi

$systemctl_bin daemon-reload
$systemd_analyze_bin verify "$unit_path"
$systemctl_bin enable nodescope-agent.service
printf 'NodeScope agent installed from verified release %s (%s). Inspect %s before startup, enroll the host to create %s/agent-token, configure non-secret agent.env, then run systemd-analyze security nodescope-agent.service.\n' "$release_tag" "${source_revision,,}" "$metadata_path" "$credential_dir"
