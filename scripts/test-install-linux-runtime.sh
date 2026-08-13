#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
installer="$repository_root/deploy/agent/install-linux.sh"
fixture="$(mktemp -d)"
trap 'sudo rm -rf "$fixture"' EXIT

cat >"$fixture/systemctl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == "--version" ]]; then printf '%s\n' 'systemd 252'; exit 0; fi
printf '%s\n' "$*" >>"${NODESCOPE_INSTALL_ROOT}/systemctl.log"
EOF
cat >"$fixture/systemd-analyze" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == "verify" && -f "$2" ]]
printf '%s\n' "$*" >>"${NODESCOPE_INSTALL_ROOT}/systemd-analyze.log"
EOF
chmod +x "$fixture/systemctl" "$fixture/systemd-analyze"

printf '%s' 'agent version one' >"$fixture/agent-v1"
printf '%s' '[Unit]\nDescription=agent version one\n' >"$fixture/unit-v1.service"
printf '%s' 'agent version two' >"$fixture/agent-v2"
printf '%s' '[Unit]\nDescription=agent version two\n' >"$fixture/unit-v2.service"

sha() { sha256sum "$1" | awk '{print $1}'; }
revision_one=0123456789abcdef0123456789abcdef01234567
revision_two=89abcdef0123456789abcdef0123456789abcdef

run_install() {
  local binary="$1" unit="$2" tag="$3" revision="$4"
  sudo env \
    NODESCOPE_INSTALL_ROOT="$fixture/root" \
    NODESCOPE_SYSTEMCTL_BIN="$fixture/systemctl" \
    NODESCOPE_SYSTEMD_ANALYZE_BIN="$fixture/systemd-analyze" \
    "$installer" "$binary" "$(sha "$binary")" "$unit" "$(sha "$unit")" "$tag" "$revision"
}

run_install "$fixture/agent-v1" "$fixture/unit-v1.service" v1.2.3 "$revision_one"
metadata="$fixture/root/var/lib/nodescope-installer/metadata/installed.env"
sudo test -f "$metadata"
sudo grep -q '^release_tag=v1.2.3$' "$metadata"
sudo grep -q "^source_revision=$revision_one$" "$metadata"
[[ "$(sudo cat "$fixture/root/usr/local/bin/nodescope-agent")" == 'agent version one' ]]

run_install "$fixture/agent-v2" "$fixture/unit-v2.service" v1.2.4 "$revision_two"
sudo grep -q '^release_tag=v1.2.4$' "$metadata"
sudo grep -q "^source_revision=$revision_two$" "$metadata"
previous_binary="$(sudo awk -F= '/^previous_binary_backup=/{print $2}' "$metadata")"
previous_unit="$(sudo awk -F= '/^previous_unit_backup=/{print $2}' "$metadata")"
previous_metadata="$(sudo awk -F= '/^previous_metadata_backup=/{print $2}' "$metadata")"
[[ -n "$previous_binary" ]] && sudo test -f "$previous_binary"
[[ -n "$previous_unit" ]] && sudo test -f "$previous_unit"
[[ -n "$previous_metadata" ]] && sudo test -f "$previous_metadata"
[[ "$(sudo cat "$previous_binary")" == 'agent version one' ]]
sudo grep -q 'agent version one' "$previous_unit"
sudo grep -q '^release_tag=v1.2.3$' "$previous_metadata"
[[ "$(sudo cat "$fixture/root/usr/local/bin/nodescope-agent")" == 'agent version two' ]]
sudo grep -q 'agent version two' "$fixture/root/etc/systemd/system/nodescope-agent.service"
sudo grep -q '^daemon-reload$' "$fixture/root/systemctl.log"
sudo grep -q '^enable nodescope-agent.service$' "$fixture/root/systemctl.log"

printf '%s\n' 'Linux installer runtime regression checks passed.'
