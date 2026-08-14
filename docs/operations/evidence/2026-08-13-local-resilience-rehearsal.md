# Local Resilience Rehearsal Record

**Recorded:** 2026-08-13. **Environment:** cloud sandbox, Linux AMD64. **Harness:** `scripts/rehearse-resilience-local.sh`. **Scope:** deterministic local control proof; not a live dual-replica or restore acceptance drill.

> This rehearsal verifies the code paths that must protect a future production drill. It does not claim a real Framework-to-Asus failover, certificate revocation propagation, lease transfer between servers, backup creation with `pg_dump`, or isolated database restore.

## Deterministic results

| Control | Expected result | Observed result | Evidence |
| --- | --- | --- | --- |
| Preferred-replica failure and fallback | A transient preferred-replica failure sends to the secondary. | **Passed.** | `TestSenderFailsOverOnTransientFailure` and `TestSenderPreflightFailsOverOnlyAfterTransientFailure`. |
| Circuit failback | After cooldown, the agent retries the preferred endpoint before the secondary. | **Passed.** | `TestSenderReturnsToPreferredReplicaAfterCircuitCooldown`. |
| TLS and certificate issuance | Agent and replica keys have segregated TLS usages; malformed, expired, non-signing, or outlived issuers fail; publication does not follow symlinks. | **Passed.** | `internal/pki` rehearsal set. |
| Backup lease fencing | A lost lease prevents archive publication both before and after archive creation. | **Passed.** | `internal/backup` rehearsal set. |
| Archive safety | Existing partial artifacts and symlink or non-regular staging entries are rejected. | **Passed.** | `internal/backup` rehearsal set. |
| Console transport | Native console refuses redirects and requires TLS 1.3. | **Passed.** | `TestHTTPSClientRequiresTLS13AndRejectsRedirects`. |

## Objective targets

| Measure | Target | Acceptance rule |
| --- | --- | --- |
| RPO — configuration and summary telemetry | **24 hours maximum**, pending administrator acceptance. | Measure the age of the last successful fenced backup and retained summary data during the live restore drill. |
| RPO — raw telemetry | **No archival-recovery commitment.** | Raw data is retention-controlled and is not a restore acceptance dependency. |
| RTO — isolated restore | **4 hours maximum**, pending measured live rehearsal and administrator acceptance. | Time from the documented restore start to validated NodeScope-only isolated database availability, including integrity and cross-schema checks. |
| Failover detection and routing | **To be measured**, not asserted locally. | Record first failed preferred attempt, first accepted secondary attempt, and first restored preferred attempt during the dual-replica drill. |

## Required live rehearsal sequence

The administrator must deploy both complete replicas with distinct ordered agent endpoints and the internal CA. The drill must then simulate preferred-replica loss, confirm secondary acceptance, restore preferred service, confirm post-cooldown failback, issue and rotate a replacement leaf, revoke the old credential according to the approved revocation mechanism, transfer the backup lease, create a backup, restore it into an isolated target, and verify NodeScope schema contents without sibling-schema access. Record the observed timestamps and compare them to the target table above.

## Known limitations and recovery

Certificate **issuance and atomic publication** are locally tested, but a deployed revocation distribution and enforcement observation is not yet available; this remains a live operational gate. Similarly, backup fencing is locally tested, but no backup target or isolated restore environment exists. On any production drill failure, stop routing agents to the affected replica, revoke or rotate the suspect credential, fence the old backup writer, restore only to an isolated target, and return traffic only after the failed invariant and the corresponding evidence record are reviewed.

## Evidence location

The executable procedure is [`scripts/rehearse-resilience-local.sh`](../../../scripts/rehearse-resilience-local.sh). Its contract is [`scripts/test-rehearse-resilience-local-contract.sh`](../../../scripts/test-rehearse-resilience-local-contract.sh). The broader dependency state remains in the [operational release ledger](../release-epics.md).
