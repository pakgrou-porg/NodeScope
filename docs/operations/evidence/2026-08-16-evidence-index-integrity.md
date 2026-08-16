# Operational Evidence Index Integrity Check

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; verifier implementation and index updates are committed with this evidence record. |
| Environment | Local NodeScope checkout. No protected database, deployment host, identity provider, runtime backend, or release service was contacted. |
| Command | `./scripts/test-operational-evidence-index-contract.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | The index rejects missing proof fields in a disposable malformed copy, accepts the tracked index, and aggregate readiness includes the new integrity contract. |
| Observed result | The tracked index passed; a fixture with an empty environment field was rejected; the complete aggregate readiness suite passed. |
| Evidence location | [`scripts/verify-operational-evidence-index.sh`](../../../scripts/verify-operational-evidence-index.sh), [`scripts/test-operational-evidence-index-contract.sh`](../../../scripts/test-operational-evidence-index-contract.sh), and [`index.md`](index.md). |
| Known limitation | The verifier enforces index completeness only. It cannot prove the truth of a source claim, replace an administrator's acceptance, or convert local/CI proof into an environment result. |
| Rollback or recovery | Revert the verifier and index update if it blocks a valid documented claim, correct the missing proof field, rerun the contract, and preserve the live-acceptance boundary. |

> The contract checks the eight index fields and a named non-acceptance boundary. It deliberately does not access external services or mutate operational systems.
