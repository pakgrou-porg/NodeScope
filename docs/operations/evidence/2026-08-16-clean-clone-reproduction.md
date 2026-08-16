# Clean-Clone Reproduction Record

**Recorded:** 2026-08-16. **Repository:** `pakgrou-porg/NodeScope`. **Commit tested:** `06924b3eff857428bd5d002e1db7a83b40e48507`. **Environment:** fresh sandbox clone at `/tmp/nodescope-clean-06924b3`. **Protected-environment activity:** none.

> The repository was cloned afresh from GitHub, detached at the recorded commit, and validated without carrying the source checkout's working tree, build outputs, local configuration, cache, or credentials into the clone.

## Reproduction sequence

```bash
gh repo clone pakgrou-porg/NodeScope /tmp/nodescope-clean-06924b3
cd /tmp/nodescope-clean-06924b3
git checkout --detach 06924b3eff857428bd5d002e1db7a83b40e48507
pnpm install --frozen-lockfile
export PATH="/home/ubuntu/.local/go1.25.12/bin:/home/ubuntu/go/bin:$PATH"
./scripts/release-readiness-check.sh
test -z "$(git status --short)"
```

| Evidence field | Result |
| --- | --- |
| Exact revision | `06924b3eff857428bd5d002e1db7a83b40e48507` |
| Dependency installation | `pnpm install --frozen-lockfile` completed successfully. |
| Go verification | Go static analysis, package tests, Linux AMD64/ARM64 cross-builds, and Windows baseline cross-compiles completed successfully. |
| TypeScript and Vitest | `tsc --noEmit` completed and Vitest reported 18 passed files, 1 skipped file, 40 passed tests, and 1 skipped test. |
| Contract and generation checks | Telemetry/OpenAPI contracts, proto generation, API generation, and generated-artifact drift checks passed. |
| Production build | Vite browser build and server bundle completed successfully. |
| Release controls | License metadata, signed-tag workflow, CI workflow, SBOM/provenance, release-evidence, installation, migration, isolation, privacy, and recovery contracts passed. |
| Final result | `NodeScope release-readiness checks passed.` followed by `CLEAN_CLONE_REPRODUCTION_PASSED commit=06924b3eff857428bd5d002e1db7a83b40e48507 status=clean`. |

## Known limitation

This validates the committed source in a clean sandbox. It does not exercise protected Supabase mutations, live magic-link authentication, host enrollment, Framework/Asus deployment, certificate issuance, real backend streaming, replica failover, or the 72-hour canary. Those remain separate environment-validation gates in the release plan.

## Recovery path

If future reproduction fails, stop release or deployment work, record the exact commit and failed command, remove the temporary clone, and compare the failure against the pinned toolchain, lockfile, generated artifacts, and release-readiness contracts. Do not work around a failure by committing cache, generated release archives, local configuration, or credentials.
