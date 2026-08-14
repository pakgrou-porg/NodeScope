# NodeScope Operational Release Ledger

**Baseline commit:** `0ddcd19` on `pakgrou-porg/NodeScope` `main`. **Purpose:** this is the dependency-aware operational ledger for release stabilization. The historical [`todo.md`](../../todo.md) remains a complete implementation history; this ledger is the authoritative readiness view for what has been proven, where the proof resides, what remains, and how a claim is recovered or invalidated.

> **Operational acceptance is intentionally empty.** A control reaches that state only after the listed environment proof, an administrator review, and recovery evidence have all been recorded. Passing local tests or CI is not acceptance for a LAN deployment.

## Evidence-state model

| State | Meaning | Promotion requirement |
| --- | --- | --- |
| **Implemented** | Reviewed code or procedure exists in the repository. | A stable source path and focused regression test or contract check. |
| **Locally validated** | Deterministic local test, build, or contract suite has passed. | Command, expected result, observed result, and commit are recorded. |
| **Environment validated** | The approved external environment or host has exercised the control. | Redacted environment evidence, known limitation, and recovery path are recorded. |
| **Operationally accepted** | An administrator has accepted the environment result for the stated release scope. | All prerequisite states plus named approval and recovery rehearsal. |

## Epic 1 — Source, supply chain, and reproducibility

| Field | Current record |
| --- | --- |
| Source TODO items | `6`, `52`, `53`, `55`–`61`, `117`, `125`, `126` |
| Current state | **Environment validated** for the GitHub clean-clone baseline; **not operationally accepted**. |
| Commit | `0ddcd19` (published baseline); the local evidence content originated in `c1794b8`. |
| Test command and expected result | Fresh clone, `pnpm install --frozen-lockfile`, then `./scripts/release-readiness-check.sh`; command and generated-contract drift must pass and leave no tracked diff. |
| Observed result | A fresh clone at `0ddcd19` installed locked dependencies, passed the aggregate readiness suite, and remained clean after generation. |
| Evidence location | This ledger; [release workflow](../../.github/workflows/release.yml); [aggregate command](../../scripts/release-readiness-check.sh). |
| License result | `LICENSE`, `package.json`, and `README.md` all declare Apache-2.0; the tracked-file MIT search returned no explicit MIT declaration. No source change was required because the mismatch was already resolved in the published baseline. |
| Known limitation | Tagged release execution, release-signing key availability, artifact download verification, and machine-readable release evidence remain unexercised. |
| Rollback or recovery | Revert the implicated baseline commit, restore from the previous signed release tag, and rerun the clean-clone command before re-publishing. |

## Epic 2 — Shared Supabase isolation and migration safety

| Field | Current record |
| --- | --- |
| Source TODO items | `11`, `31`–`38`, `40`, `41`, `114`–`116`, `119` |
| Current state | **Environment validated** for read-only role boundaries, disposable sibling-schema denial, and one rollback-only migrator preflight; **not operationally accepted**. |
| Commit | `3f6118f`, `d5d92b6`, and `0ddcd19`. |
| Test command and expected result | `scripts/verify-sibling-denials.sh` must deny read, DML, DDL, routine execution, and routine replacement to both NodeScope login paths; a selected migration must run inside `BEGIN`/`ROLLBACK` and leave no migration ledger entry or object. |
| Observed result | Both login paths were denied all tested fixture operations; fixture cleanup passed; `0015_terminal_fleet_status.sql` passed a dedicated-migrator rollback preflight and remained unrecorded with no function persisted. |
| Evidence location | [Read-only preflight](evidence/2026-08-13-shared-supabase-readonly-preflight.md), [sibling denial](evidence/2026-08-13-sibling-schema-denial-gate.md), [rollback preflight](evidence/2026-08-13-migrator-rollback-preflight.md). |
| Known limitation | No production migration has been applied; the agent role, real sibling TTRPG-OCR schema, pg_cron jobs, RLS object matrix, and post-apply verification are incomplete. |
| Rollback or recovery | The fixture has been removed. For any future apply, halt on gate failure, restore the last database backup or migration rollback procedure, and rerun sibling-denial plus post-apply isolation verification. |

## Epic 3 — Framework Linux primary canary

| Field | Current record |
| --- | --- |
| Source TODO items | `14`, `15`, `45`–`47`, `67`–`70`, `120` |
| Current state | **Locally validated** for agent protocol, Linux collectors, installation contracts, receipt-time verification, retry/failover, and evidence-quality behavior; **environment validation blocked** on Framework host access. |
| Commit | Local controls are represented by the current `main` baseline and historical commits referenced by [`todo.md`](../../todo.md). |
| Required environment test | Enroll Framework with a revocable credential, collect real telemetry at the approved interval, force retry/idempotency behavior, and produce host qualification plus storage evidence. |
| Expected result | Authenticated samples arrive with explicit provenance and quality; duplicate delivery is safely handled; raw-retention qualification has no unacceptable gap or storage result. |
| Evidence location | [Framework/Asus manual guide](agents/manual-install-framework-asus-v2.md), [activation gates](activation-gates.md), future `docs/operations/evidence/framework-*` records. |
| Known limitation | No live Framework telemetry or 72-hour storage qualification has been recorded; AMD GPU/NPU evidence remains experimental until the version matrix is qualified. |
| Rollback or recovery | Revoke the agent credential, stop the systemd service, remove the staged version using the installer rollback metadata, and retain only redacted evidence. |

## Epic 4 — Dual replica, PKI, backups, and recovery

| Field | Current record |
| --- | --- |
| Source TODO items | `12`, `27`, `87`–`90`, `121` |
| Current state | **Locally validated** for fail-closed lease, PKI, transport, and archive contracts; **environment validation pending**. |
| Required environment test | Deploy Framework primary and Asus secondary replicas; rehearse primary loss, failback, certificate rotation/revocation, backup-lease handoff, and an isolated restore. |
| Expected result | Agents fail over in configured order and return to preferred; revoked/expired certificates fail closed; only the lease holder publishes a backup; recovered data meets declared RPO/RTO. |
| Evidence location | [Activation gates](activation-gates.md) and future replica, PKI, backup, and restore evidence records. |
| Known limitation | No production PKI, replica, backup target, restore sandbox, or objective RPO/RTO measurement exists. |
| Rollback or recovery | Stop the unhealthy replica, revoke or rotate the affected leaf, fence the old backup writer, restore to the isolated target, and promote only after integrity verification. |

## Epic 5 — Console identity, RBAC, and degraded operation

| Field | Current record |
| --- | --- |
| Source TODO items | `20`, `21`, `25`, `122` |
| Current state | **Implemented** for role-aware REST/MCP contracts; **locally validated** for contract tests; **environment validation pending** for real Supabase magic links and browser E2E. |
| Required environment test | Invite a test user, verify Viewer/Operator/Administrator boundaries in a real browser, exercise both replica callbacks, and repeat under degraded or failed preferred replica conditions. |
| Expected result | Authentication and role scopes match the documented matrix; unauthorized actions fail closed without secrets; session continuity and error paths are observable. |
| Evidence location | [AgentZero integration guide](../../integrations/agentzero/README.md), browser and router contract tests, future E2E evidence records. |
| Known limitation | The preview console uses development fixtures; real magic-link configuration and browser E2E have not run. |
| Rollback or recovery | Disable affected invite/API credential, invalidate session material, restore the previous callback configuration, and preserve the audit record. |

## Epic 6 — Metadata-only inference privacy and approved runtimes

| Field | Current record |
| --- | --- |
| Source TODO items | `23`, `24`, `72`–`86`, `94`, `124` |
| Current state | **Locally validated** for adversarial content-retention, redirect-containment, opaque backend, and failover contracts; **environment validation pending** for real streaming backends. |
| Required environment test | Route a canary stream through an approved vLLM, llama.cpp, or LM Studio backend; inspect proxy, audit, trace, support-export, and usage outputs for metadata-only handling. |
| Expected result | Stream behavior and retryable fallback work while prompt, completion, endpoint credential, and backend location data remain absent from all observable outputs. |
| Evidence location | [Proxy handler](../../internal/proxy/handler.go), [observability adapter](../../internal/proxy/observability.go), and future real-stream evidence. |
| Known limitation | Actual backend streaming and approved-runtime discovery are not yet exercised on a live host. |
| Rollback or recovery | Disable the route, revoke the client key, remove the runtime approval, and rotate any exposed backend credential before resuming traffic. |

## Epic 7 — Windows support boundary

| Field | Current record |
| --- | --- |
| Source TODO items | `54`–`56`, `123` |
| Current state | **Implemented** and **locally validated** for an explicit unsupported-evidence baseline and Windows AMD64/ARM64 CI builds; **Windows remains unsupported operationally**. |
| Required environment test | Produce a signed Windows installer/update/rollback path, gather a capability report, and qualify MSI RTX 5080 plus LM Studio telemetry under the approved evidence model. |
| Expected result | The installer verifies signed artifacts, reports unsupported signals honestly, supports rollback, and produces reproducible RTX/LM Studio qualification evidence. |
| Evidence location | [Release workflow](../../.github/workflows/release.yml), Windows agent CI jobs, and future MSI qualification evidence. |
| Known limitation | No signed installer/update/rollback rehearsal or MSI RTX 5080/LM Studio evidence exists. |
| Rollback or recovery | Keep Windows disabled in fleet enrollment; uninstall via signed rollback metadata and revoke the associated agent credential if a test fails. |

## Epic 8 — Operational acceptance and aggregate evidence

| Field | Current record |
| --- | --- |
| Source TODO items | `28`–`30`, `49`–`51`, `125`, `126` |
| Current state | **Implemented** for local aggregation; **not operationally accepted**. |
| Required test | Produce deterministic aggregate output containing commit SHA, command, result, environment, evidence path, limitation, and recovery path for every completed operational claim; attach signed artifacts, checksums, SBOMs, provenance, and release evidence to an approved signed tag. |
| Expected result | A reviewer can reproduce each claim and determine whether the release is blocked, accepted, or recoverable without interpreting source history manually. |
| Evidence location | [Aggregate readiness command](../../scripts/release-readiness-check.sh), [release workflow](../../.github/workflows/release.yml), this ledger, and future machine-readable report. |
| Known limitation | The aggregate command prints human-readable results only; no signed release run or accepted operational evidence report exists. |
| Rollback or recovery | Do not publish or promote a tag on any failed gate; revoke a release, restore the prior accepted tag, and regenerate the report after remediation. |

## Dependency order

1. **Epic 1** maintains the reproducible source and artifact baseline.
2. **Epic 2** must pass before any protected-database apply.
3. **Epic 3** qualifies Framework before Asus or other fleet expansion.
4. **Epic 4** requires two deployed replicas and follows Framework canary stability.
5. **Epic 5** and **Epic 6** require live server paths; both must pass before operational acceptance.
6. **Epic 7** stays explicitly unsupported until its own independent qualification completes.
7. **Epic 8** is the final acceptance gate and cannot be promoted while any predecessor is below environment validation.
