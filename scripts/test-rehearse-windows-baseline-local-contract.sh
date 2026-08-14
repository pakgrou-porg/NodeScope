#!/usr/bin/env bash
# Ensure Windows stays explicitly unsupported until its live qualification gates
# are satisfied.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

harness=scripts/rehearse-windows-baseline-local.sh
for required in \
  'GOOS=windows GOARCH="$architecture" go build' \
  'windows_baseline_capability_boundary":"locally validated' \
  'native_windows_execution":"live environment gate' \
  'signed_windows_installer":"live environment gate' \
  'windows_update_and_rollback_rehearsal":"live environment gate' \
  'msi_rtx_5080_qualification":"live environment gate' \
  'lm_studio_qualification":"live environment gate' \
  'Keep Windows enrollment disabled'; do
  grep -Fq "$required" "$harness" || {
    echo "Windows baseline rehearsal must retain $required" >&2
    exit 1
  }
done

echo "Windows baseline readiness contract passed."
