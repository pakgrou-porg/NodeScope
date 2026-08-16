# Local Aggregate Readiness Report

| Evidence field | Record |
| --- | --- |
| Commit | `54415cefad5fdfa82fc81f367d1386fcf1107385` |
| Environment | Clean local NodeScope checkout; no protected database, deployment host, identity provider, or approved inference backend was contacted. |
| Command | `./scripts/write-release-readiness-report.sh --output /home/ubuntu/nodescope-evidence/release-readiness-54415ce.json` followed by `./scripts/verify-release-readiness-report.sh /home/ubuntu/nodescope-evidence/release-readiness-54415ce.json` |
| Expected result | The aggregate suite passes; an externally stored JSON report records a clean tree, five deterministic check groups, per-group evidence boundaries, retained live gates, and recovery guidance. |
| Observed result | The aggregate suite passed; `schema_version: 2`, `working_tree_clean: true`, five check groups, retained live gates, and recovery guidance were verified. |
| Evidence location | External report: `/home/ubuntu/nodescope-evidence/release-readiness-54415ce.json`; generator: [`scripts/write-release-readiness-report.sh`](../../../scripts/write-release-readiness-report.sh); verifier: [`scripts/verify-release-readiness-report.sh`](../../../scripts/verify-release-readiness-report.sh). |
| Known limitation | This is local source, build, fixture, and contract evidence only. It does not establish Framework/Asus hardware qualification, live replica recovery, PKI revocation, isolated restore, real Supabase magic-link E2E, real backend streaming, a tagged release, or GitHub release attestation verification. |
| Recovery | Do not promote a release from this report. Preserve failed output, restore the previous accepted signed tag as appropriate, remediate the failed local contract, and regenerate the report from a clean tree. |

> The generator refused the dirty implementation tree before this run, and it only wrote outside the repository. This prevents generated evidence or local material from being accidentally committed.
