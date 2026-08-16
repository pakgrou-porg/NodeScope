#!/usr/bin/env bash
# Verify that clean-clone reproduction stays explicit, external, locked, and
# fails closed without creating a network clone during this contract test.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

for required in \
  '--commit is required' \
  '--workspace-root must refer to a path outside the repository' \
  'clean-clone destination already exists' \
  'gh repo clone "$repository" "$clone_path"' \
  'git checkout --detach "$commit"' \
  'pnpm install --frozen-lockfile' \
  './scripts/release-readiness-check.sh' \
  'clean-clone readiness generated an uncommitted change' \
  'CLEAN_CLONE_REPRODUCTION_PASSED'; do
  grep -Fq -- "$required" scripts/reproduce-clean-clone-readiness.sh || {
    echo "clean-clone procedure must retain: $required" >&2
    exit 1
  }
done

set +e
./scripts/reproduce-clean-clone-readiness.sh >/tmp/nodescope-clean-clone-missing-commit.out 2>&1
missing_commit_status=$?
./scripts/reproduce-clean-clone-readiness.sh --commit fixture --workspace-root "$repository_root" >/tmp/nodescope-clean-clone-inside-repo.out 2>&1
inside_repository_status=$?
set -e
test "$missing_commit_status" -eq 2
test "$inside_repository_status" -eq 1
grep -Fq -- '--commit is required' /tmp/nodescope-clean-clone-missing-commit.out
grep -Fq -- '--workspace-root must refer to a path outside the repository' /tmp/nodescope-clean-clone-inside-repo.out

echo "Clean-clone reproduction contract passed."
