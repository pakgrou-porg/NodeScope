#!/usr/bin/env bash
# Preserve the distinction between deterministic local resilience proof and
# live operational rehearsal requirements.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

harness=scripts/rehearse-resilience-local.sh
for required in \
  'TestSenderReturnsToPreferredReplicaAfterCircuitCooldown' \
  'TestBackupRefusesFinalPublicationWhenLeaseIsLostDuringArchiveCreation' \
  'TestIssueRejectsInvalidOrOutlivedCertificateAuthorities' \
  'certificate_revocation":"live operational gate' \
  'isolated_restore":"live operational gate' \
  '24 hours pending administrator acceptance' \
  '4 hours pending measured live rehearsal'; do
  if ! grep -Fq "$required" "$harness"; then
    echo "local resilience rehearsal must retain $required" >&2
    exit 1
  fi
done

echo "Local resilience rehearsal contract passed."
