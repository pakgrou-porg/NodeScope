#!/usr/bin/env bash
# Exercise the index verifier against the committed index and a disposable
# malformed copy. No external service or protected environment is contacted.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

for required in \
  'Every future operational evidence record must add or update an index row containing all eight proof fields' \
  'operational evidence index row has an empty proof field' \
  'not operationally accepted' \
  'Operational evidence index verified:'; do
  grep -Fq -- "$required" scripts/verify-operational-evidence-index.sh docs/operations/evidence/index.md || {
    echo "evidence-index contract missing: $required" >&2
    exit 1
  }
done

./scripts/verify-operational-evidence-index.sh

fixture="$(mktemp)"
trap 'rm -f "$fixture"' EXIT
sed '0,/| Cloud sandbox fresh clone |/s//|  |/' docs/operations/evidence/index.md >"$fixture"
if ./scripts/verify-operational-evidence-index.sh --index "$fixture" >/tmp/nodescope-evidence-index-invalid.out 2>&1; then
  echo "evidence-index verifier accepted a row with an empty proof field" >&2
  exit 1
fi
grep -Fq 'operational evidence index row has an empty proof field' /tmp/nodescope-evidence-index-invalid.out

echo "Operational evidence index contract passed."
