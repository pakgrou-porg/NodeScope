#!/usr/bin/env bash
# Run deterministic local resilience controls. This is preparation for, not a
# substitute for, a live two-replica, PKI-revocation, backup, and restore drill.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

if ! command -v go >/dev/null 2>&1; then
  echo "required dependency is unavailable: go" >&2
  exit 2
fi

go test ./internal/agent -run '^(TestSenderFailsOverOnTransientFailure|TestSenderPreflightFailsOverOnlyAfterTransientFailure|TestSenderCircuitSkipsRepeatedlyFailingPreferredReplica|TestSenderReturnsToPreferredReplicaAfterCircuitCooldown)$' -count=1
go test ./internal/pki -run '^(TestReplicaAndAgentCertificatesHaveSeparatedTLSUsage|TestIssueRejectsInvalidOrOutlivedCertificateAuthorities|TestCertificatePublicationReplacesSymlinkWithoutFollowingIt)$' -count=1
go test ./internal/backup -run '^(TestDefaultBackupExcludesRawTelemetry|TestBackupRefusesPublicationAfterLeaseLoss|TestBackupRefusesFinalPublicationWhenLeaseIsLostDuringArchiveCreation|TestArchiveCreationNeverOverwritesOrFollowsExistingPartialPath|TestArchiveCreationRejectsSymlinkInStagingSource)$' -count=1
go test ./internal/consoleclient -run '^TestHTTPSClientRequiresTLS13AndRejectsRedirects$' -count=1

cat <<'JSON'
{"version":1,"scope":"local deterministic resilience rehearsal","passed":true,"controls":{"replica_failover_and_failback":"locally validated","tls13_transport":"locally validated","certificate_issuance_and_atomic_publication":"locally validated","certificate_revocation":"live operational gate","backup_lease_fencing":"locally validated","backup_archive_safety":"locally validated","isolated_restore":"live operational gate"},"rpo_target":{"configuration_and_summary_telemetry":"24 hours pending administrator acceptance","raw_telemetry":"not an archival recovery commitment"},"rto_target":{"isolated_restore":"4 hours pending measured live rehearsal and administrator acceptance"}}
JSON
