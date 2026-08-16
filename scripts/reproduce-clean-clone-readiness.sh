#!/usr/bin/env bash
# Reproduce deterministic release readiness from an explicitly selected GitHub
# revision. Clones and outputs are deliberately external to this source tree.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repository="pakgrou-porg/NodeScope"
commit=""
workspace_root="${TMPDIR:-/tmp}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repository)
      repository="${2:-}"
      shift 2
      ;;
    --commit)
      commit="${2:-}"
      shift 2
      ;;
    --workspace-root)
      workspace_root="${2:-}"
      shift 2
      ;;
    *)
      echo "usage: $0 --commit <sha-or-tag> [--repository <owner/repository>] [--workspace-root <external-path>]" >&2
      exit 2
      ;;
  esac
done

if [[ -z "$commit" ]]; then
  echo "--commit is required" >&2
  exit 2
fi
if [[ -z "$workspace_root" ]]; then
  echo "--workspace-root must not be empty" >&2
  exit 2
fi

mkdir -p "$workspace_root"
workspace_abs="$(cd "$workspace_root" && pwd)"
case "$workspace_abs" in
  "$repository_root"|"$repository_root"/*)
    echo "--workspace-root must refer to a path outside the repository" >&2
    exit 1
    ;;
esac

clone_name="nodescope-clean-${commit//[^a-zA-Z0-9._-]/_}"
clone_path="$workspace_abs/$clone_name"
if [[ -e "$clone_path" ]]; then
  echo "clean-clone destination already exists: $clone_path" >&2
  echo "choose a new external workspace or remove the prior disposable clone deliberately" >&2
  exit 1
fi

command -v gh >/dev/null || { echo "gh is required to clone the selected public repository revision" >&2; exit 1; }
command -v pnpm >/dev/null || { echo "pnpm is required for locked browser dependency installation" >&2; exit 1; }

gh repo clone "$repository" "$clone_path"
cd "$clone_path"
git checkout --detach "$commit"
actual_commit="$(git rev-parse HEAD)"

if [[ -x /home/ubuntu/.local/go1.25.12/bin/go ]]; then
  export PATH="/home/ubuntu/.local/go1.25.12/bin:/home/ubuntu/go/bin:$PATH"
fi
command -v go >/dev/null || { echo "Go is required for the aggregate readiness suite" >&2; exit 1; }

pnpm install --frozen-lockfile
./scripts/release-readiness-check.sh

if [[ -n "$(git status --porcelain)" ]]; then
  echo "clean-clone readiness generated an uncommitted change" >&2
  exit 1
fi

printf 'CLEAN_CLONE_REPRODUCTION_PASSED repository=%s commit=%s status=clean clone=%s\n' "$repository" "$actual_commit" "$clone_path"
