#!/usr/bin/env bash
# Verify the reviewable repository baseline remains explicit about exclusions,
# license consistency, release workflow provenance, and clean-clone recovery.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
inventory="$repository_root/docs/operations/evidence/2026-08-16-repository-recovery-inventory.md"
release_plan="$repository_root/docs/operations/release-plan.md"

for required in \
  'every changed or untracked path' \
  '.manus-logs/' \
  '.project-config.json' \
  'node_modules/' \
  '.env*' \
  'Apache-2.0' \
  'git status --short' \
  'clean-clone reproduction'; do
  grep -Fq -- "$required" "$inventory" || { echo "repository recovery inventory missing: $required" >&2; exit 1; }
done

for required in \
  'Implemented' \
  'Locally validated' \
  'Environment validated' \
  'Operationally accepted' \
  'Repository provenance and reproducibility' \
  'Shared-Supabase isolation and migration safety' \
  'Framework Linux canary' \
  'Console authentication and RBAC' \
  'Seventy-two-hour canary' \
  'Windows is a separately gated product track'; do
  grep -Fq -- "$required" "$release_plan" || { echo "release plan missing: $required" >&2; exit 1; }
done

grep -Fq 'Apache-2.0' "$repository_root/package.json" || { echo 'package metadata is missing Apache-2.0' >&2; exit 1; }
grep -Fq 'Apache License' "$repository_root/LICENSE" || { echo 'LICENSE is not Apache License text' >&2; exit 1; }
test -f "$repository_root/.github/workflows/release.yml" || { echo 'release workflow is missing' >&2; exit 1; }
grep -Fq 'actions/attest@v4' "$repository_root/.github/workflows/release.yml" || { echo 'release workflow lacks artifact attestation' >&2; exit 1; }
grep -Fq 'anchore/sbom-action@v0' "$repository_root/.github/workflows/release.yml" || { echo 'release workflow lacks SBOM generation' >&2; exit 1; }

for ignored_path in '.manus-logs/devserver.log' '.project-config.json' 'node_modules/package.json' '.env'; do
  git -C "$repository_root" check-ignore -q "$ignored_path" || { echo "expected ignored path is not ignored: $ignored_path" >&2; exit 1; }
done

echo 'Repository recovery contract passed.'
