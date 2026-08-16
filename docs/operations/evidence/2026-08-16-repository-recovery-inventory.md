# Repository Recovery Inventory

**Baseline revision:** `32e32e225aa584676ffa2dfeb6f3f60308c50cd8`. **Captured:** 2026-08-16. **Scope:** recovery freeze; no protected-environment deployment, shared-Supabase mutation, host deployment, certificate action, credential enrollment, or broad feature expansion was performed.

> This is an inventory of **every changed or untracked path in the recovered working tree**, not a claim that every tracked source module is operationally accepted. The repository began this recovery with one modified, unstaged path and no untracked paths.

## Changed and untracked paths

| Path | Classification | Disposition | Rationale |
| --- | --- | --- | --- |
| `todo.md` | Documentation / task ledger | Intended recovery commit | Records the recovery freeze and acceptance criteria requested for the reviewable baseline. |
| `docs/operations/evidence/2026-08-16-repository-recovery-inventory.md` | Documentation / evidence | Intended recovery commit | This inventory and its exclusions. |
| `docs/operations/release-plan.md` | Documentation / release management | Intended recovery commit | Dependency-aware Release 1 gates and evidence states. |
| `scripts/test-repository-recovery-contract.sh` | Test / tooling | Intended recovery commit | Verifies the recoverability documentation, exclusions, Apache-2.0 declarations, and version-controlled release workflow. |

There were **no staged paths** and **no other untracked paths** at capture time. The intended recovery change set is limited to the four paths in the table above; it contains no product feature implementation or protected-environment procedure.

## Tracked baseline classification

The tracked baseline contains the following path classes, counted by repository path prefix. These are review categories, not evidence states.

| Classification | Tracked file count | Principal locations |
| --- | ---: | --- |
| Source | 254 | `client/`, `server/`, `internal/`, `cmd/`, `drizzle/` |
| Test / tooling | 49 | `scripts/` |
| Infrastructure / deployment | 26 | `deploy/`, `supabase/isolation/`, `supabase/operations/` |
| Migration | 19 | `supabase/migrations/`, `drizzle/migrations/` |
| Contracts | 6 | `api/`, `telemetry/` |
| CI / release | 7 | `.github/` |
| Documentation | 65 | `docs/`, `README.md`, `todo.md` |
| License / repository metadata | 3 | `LICENSE` and root repository metadata |

## Intentionally excluded local material

| Path or pattern | Classification | Reason for exclusion | Commit status |
| --- | --- | --- | --- |
| `.manus-logs/` | Local generated diagnostics | Sandbox development logs; not source or release evidence. | Ignored; never commit. |
| `.project-config.json` | Local configuration | Managed workspace configuration; may identify sandbox-specific state. | Ignored; never commit. |
| `node_modules/` | Generated dependency cache | Recreated from `pnpm-lock.yaml`. | Ignored; never commit. |
| `.env*` | Local configuration / secret | Environment and credential material. | Ignored; never commit. |
| `*.pem`, `*.key`, `*.p12`, `*.pfx`, credential files | Local secret | Certificate keys and credentials are prohibited from source control. | No tracked instances; never commit. |
| `release/`, `dist/`, temporary reports, binary archives | Generated artifact | Produced by the release or build process and represented by checksums, SBOMs, provenance, and workflow artifacts instead. | Excluded by release workflow and ignore policy; never commit. |

The tracked-file name scan found no certificate, private-key, credential, or environment-file path. The release workflow includes an automated secret scan. This inventory does not treat a filename scan as proof of secret absence; the clean-clone reproduction additionally runs the repository’s release-readiness command and CI-equivalent checks.

## Reviewable commit baseline

The recovered history is already partitioned into logical units. The recovery commit adds only baseline evidence and the release-gate plan. Prior implementation commits remain individually reviewable by domain: infrastructure and Portainer (`f2f0e66`), agent enrollment (`b633b68`), auxiliary-agent validation documentation (`e936f66`), operational-role validation (`039f330`), migration-prefix guard (`1f74e46`), and pg_cron preparation evidence (`32e32e2`). The clean-clone report will name the exact recovery commit and test output.

## Recovery condition

The repository must have an empty `git status --short` after the intended recovery documentation and contract are committed. Any later modified or untracked path requires a new classification entry before review, rather than being bundled into this baseline.
