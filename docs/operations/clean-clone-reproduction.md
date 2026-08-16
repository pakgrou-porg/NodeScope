# Reproducing Release Readiness from a Clean Clone

Use the reusable procedure to prove a specific published Git revision from a fresh external clone. It always requires an explicit commit or tag and refuses any workspace inside the source repository. The procedure runs locked JavaScript dependency installation, the full aggregate readiness suite, and a final clean-tree check.

> This procedure validates source, build, generated-contract, fixture, and policy controls. It does not perform a deployment or test a protected database, a LAN host, Supabase Auth, PKI, an approved runtime, or a release service.

```bash
set -euo pipefail
cd /path/to/NodeScope

./scripts/reproduce-clean-clone-readiness.sh \
  --commit "$(git rev-parse HEAD)" \
  --workspace-root "$HOME/nodescope-clean-clones"
```

The command creates one disposable directory named from the requested revision and exits rather than reusing an existing directory. Its final `CLEAN_CLONE_REPRODUCTION_PASSED` line identifies the repository, immutable detached commit, and clone path. Preserve that line and the aggregate command output as local release-review evidence, then remove the clone when retention is no longer required.

| Failure point | Meaning | Recovery |
| --- | --- | --- |
| Destination exists | A prior clone could invalidate fresh-clone proof. | Select a new external workspace or deliberately remove the prior disposable clone. |
| Locked install fails | The declared dependency graph does not reproduce. | Restore the prior accepted lockfile or remediate dependency metadata. |
| Aggregate readiness fails | A source, build, contract, fixture, or policy control failed. | Preserve output, stop promotion, remediate, and rerun from a new clone. |
| Final tree is dirty | A generator or build changed tracked source. | Do not commit generated drift; correct the generator or source and repeat. |

Protected-environment work remains separate. A successful clean clone does not advance any environment-validation or operational-acceptance gate.
