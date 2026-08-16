#!/usr/bin/env bash
# Validate the stable field set of a locally generated readiness report without
# contacting any deployment, database, host, or external runtime.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <report.json>" >&2
  exit 2
fi

report="$1"
if [[ ! -f "$report" ]]; then
  echo "readiness report does not exist: $report" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
node "$script_dir/verify-release-readiness-report.mjs" "$report"

echo "Release-readiness report verified: $report"
