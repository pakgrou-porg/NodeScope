# Local Machine-Readable Release-Readiness Report

The deterministic aggregate readiness suite proves that the committed source tree meets the repository's **local** contracts. It does not prove a Framework or Asus deployment, real Supabase Auth, production PKI, live replica recovery, backend streaming privacy, or signed-tag release publication. The report generator preserves that distinction in a JSON artifact intended for review and archival outside the Git working tree.

> Generate the report only from a clean checkout. The generator refuses a dirty tree and refuses an output path within the repository, so generated evidence, logs, local configuration, and credentials cannot be accidentally committed.

## Local procedure

```bash
set -euo pipefail
cd /path/to/NodeScope
git status --short

report_dir="$HOME/nodescope-evidence"
mkdir -p "$report_dir"
report="$report_dir/release-readiness-$(git rev-parse --short HEAD).json"

./scripts/write-release-readiness-report.sh --output "$report"
./scripts/verify-release-readiness-report.sh "$report"
```

The report includes the immutable commit SHA and commit timestamp, a clean-tree assertion, and five check groups. Each group records its command, expected result, observed result, source-only evidence boundary, remaining limitation, and recovery direction. The verifier parses the JSON and checks the exact unique group set, required non-empty proof fields, evidence locations, retained live-gate set, and recovery text; it does not make a deployment decision or connect to any protected environment.

| Check group | Locally established | Still deferred |
| --- | --- | --- |
| Source and policy | Repository recovery, license, workflow, installation, and release contracts | Signed-tag release and external attestation verification |
| Shared-schema safety | Disposable fixture and isolation contracts | Real sibling schema and production telemetry path |
| Local resilience | Deterministic lease, PKI, TLS, backup, and failover rehearsal | Replica deployment, revocation, and isolated restore |
| Native builds | Go tests plus Linux and Windows cross-builds | Framework/Asus/MSI hardware qualification |
| Browser console | TypeScript, Vitest, contracts, and production bundle | Real magic-link/RBAC/degraded-replica browser E2E |

## Review and recovery boundary

Store the report alongside a release-review ticket or controlled evidence archive; it is not a release artifact and must not be represented as environment acceptance. If the report fails, do not promote a release. Retain the command output, restore or revert to the prior accepted signed tag as appropriate, remediate the failed local contract, and regenerate the report from a clean checkout.
