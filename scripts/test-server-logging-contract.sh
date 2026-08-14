#!/usr/bin/env bash
# Server diagnostics must pass through the bounded logger so error bodies,
# endpoint details, credentials, and session values cannot reach console output.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

matches="$(grep -R -nE 'console\.(log|error|warn|info|debug)' server --include='*.ts' --exclude='logger.ts' || true)"
if [[ -n "$matches" ]]; then
  echo "direct server console diagnostics are prohibited:" >&2
  printf '%s\n' "$matches" >&2
  exit 1
fi

for required in 'sensitiveKey' 'value.slice(0, 128)' 'console[level](entry)'; do
  grep -Fq "$required" server/_core/logger.ts || {
    echo "bounded server logger must retain $required" >&2
    exit 1
  }
done

echo "Server logging contract passed."
