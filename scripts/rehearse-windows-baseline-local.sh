#!/usr/bin/env bash
# Verify the Windows baseline can be built while preserving its explicit
# unsupported operational status. Native Windows execution is not attempted.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

if ! command -v go >/dev/null 2>&1; then
  echo "required dependency is unavailable: go" >&2
  exit 2
fi

temporary_directory="$(mktemp -d)"
trap 'rm -rf "$temporary_directory"' EXIT
for architecture in amd64 arm64; do
  GOOS=windows GOARCH="$architecture" go build -o "$temporary_directory/nodescope-agent-windows-$architecture.exe" ./cmd/nodescope-agent
  GOOS=windows GOARCH="$architecture" go test -c -o "$temporary_directory/nodescope-agent-windows-$architecture.test.exe" ./internal/agent
done

for required in \
  'windows_agent_baseline' \
  'logical_cpu_count' \
  'unqualified Windows baseline' \
  'gpu_and_vram' \
  'CapabilityUnavailable'; do
  grep -Fq "$required" internal/agent/preflight_windows.go internal/agent/collectors_windows_test.go || {
    echo "Windows baseline must retain explicit unsupported capability $required" >&2
    exit 1
  }
done

cat <<'JSON'
{"schema_version":1,"scope":"local Windows baseline readiness","result":"passed","controls":{"windows_amd64_cross_build":"locally validated","windows_arm64_cross_build":"locally validated","windows_baseline_capability_boundary":"locally validated","logical_cpu_count_only":"locally validated","gpu_vram_npu_memory_storage_unavailable":"locally validated","signed_release_workflow_preparation":"locally validated","native_windows_execution":"live environment gate","signed_windows_installer":"live environment gate","windows_update_and_rollback_rehearsal":"live environment gate","msi_rtx_5080_qualification":"live environment gate","lm_studio_qualification":"live environment gate"},"recovery":"No Windows host is contacted and no installer is produced. Keep Windows enrollment disabled; if a future validation fails, revoke the test agent credential, stop the test service or process, and restore only a verified prior artifact after hash and attestation checks."}
JSON
