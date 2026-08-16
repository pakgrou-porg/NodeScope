#!/usr/bin/env bash
# Enforce the eight proof fields for every indexed operational claim. This is a
# documentation-integrity check only; it does not promote local evidence to an
# environment result or contact any protected system.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
index="$repository_root/docs/operations/evidence/index.md"

if [[ $# -gt 0 ]]; then
  if [[ "$1" != "--index" || $# -ne 2 ]]; then
    echo "usage: $0 [--index <path>]" >&2
    exit 2
  fi
  index="$2"
fi

if [[ ! -f "$index" ]]; then
  echo "operational evidence index does not exist: $index" >&2
  exit 1
fi

for required in \
  'Commit' \
  'Validation command or procedure' \
  'Environment' \
  'Expected and observed result' \
  'Evidence location' \
  'Known limitation' \
  'Rollback or recovery' \
  'not operationally accepted' \
  'Every future operational evidence record must add or update an index row containing all eight proof fields'; do
  grep -Fq -- "$required" "$index" || {
    echo "operational evidence index missing required boundary: $required" >&2
    exit 1
  }
done

awk -F'|' '
  function trim(value) { gsub(/^[[:space:]]+|[[:space:]]+$/, "", value); return value }
  /^\|/ {
    claim = trim($2)
    if (claim == "Claim and evidence state" || claim ~ /^---/) next
    if (NF != 10) {
      printf "operational evidence index row has %d fields, expected 8 proof fields: %s\n", NF - 2, $0 > "/dev/stderr"
      invalid = 1
      next
    }
    for (field = 2; field <= 9; field++) {
      if (trim($field) == "") {
        printf "operational evidence index row has an empty proof field: %s\n", $0 > "/dev/stderr"
        invalid = 1
      }
    }
    rows++
  }
  END {
    if (rows == 0) {
      print "operational evidence index has no claim rows" > "/dev/stderr"
      exit 1
    }
    exit invalid
  }
' "$index"

echo "Operational evidence index verified: $index"
