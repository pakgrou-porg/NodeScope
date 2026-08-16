# Readiness-Report Schema Validation

| Evidence field | Record |
| --- | --- |
| Commit | Pending publication; parser, regression coverage, and this evidence record are committed together. |
| Environment | Local NodeScope checkout. The required Go toolchain and protobuf compiler were restored after sandbox reset; no protected database, deployment host, identity provider, inference backend, or release service was contacted. |
| Command | `./scripts/test-release-readiness-report-contract.sh` and `./scripts/release-readiness-check.sh` |
| Expected result | The verifier parses valid JSON, rejects malformed JSON and duplicate check identifiers, validates the exact required schema and live-gate set, and the aggregate suite passes. |
| Observed result | The valid fixture passed; malformed JSON and duplicate IDs were rejected; the full aggregate readiness suite passed. |
| Evidence location | [`verify-release-readiness-report.mjs`](../../../scripts/verify-release-readiness-report.mjs), [`verify-release-readiness-report.sh`](../../../scripts/verify-release-readiness-report.sh), and [`test-release-readiness-report-contract.sh`](../../../scripts/test-release-readiness-report-contract.sh). |
| Known limitation | Semantic report validation proves only the structure and required local boundaries of a report. It cannot prove a report's source claims, contact a protected system, or confer environment validation or operational acceptance. |
| Rollback or recovery | Revert the parser if a verified backward-compatible report schema requires a controlled update; otherwise correct the report generator, preserve failure output, and rerun the local contract and aggregate suite. |

> The first aggregate-suite attempt stopped because the reset sandbox lacked Go and `protoc`. Both prerequisites were restored from verified/local package sources; the subsequent full suite passed without altering project source or protected infrastructure.
