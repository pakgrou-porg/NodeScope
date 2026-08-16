# Verified NodeScope releases and automatic agent updates

NodeScope publishes Linux AMD64 and ARM64 native artifacts through GitHub Releases and a multi-architecture replica image through GitHub Container Registry. Release automation generates SHA-256 checksums, SPDX SBOMs, GitHub artifact attestations for native archives, and image provenance/SBOM attestations. Release builds never pass credentials through Docker build arguments; runtime credentials stay in protected host configuration.[1] [2]

Before workflow attestation, NodeScope assembles release evidence only from safe-named archives and SPDX SBOMs under a canonical signed tag and an immutable 40-character source revision. The resulting manifest is parsed and semantically validated for exact artifact sidecars, SHA-256 digests, unique names, required provenance text, and required signing/verification guidance. This local control does not publish a release or establish operational acceptance.

## Operator verification

Before an initial agent install or an update-manifest change, download the exact pinned archive and verify its GitHub provenance:

```bash
gh attestation verify ./nodescope_<version>_linux_<arch>.tar.gz \
  -R pakgrou-porg/NodeScope
sha256sum -c ./nodescope_<version>_linux_<arch>.tar.gz.sha256

# Runs the same checksum, attestation, and immutable release-target checks
# that manual installation evidence requires; it makes no host changes.
./scripts/verify-manual-agent-install.sh \
  ./nodescope_<version>_linux_<arch>.tar.gz \
  ./nodescope_<version>_linux_<arch>.tar.gz.sha256 \
  ./nodescope_<version>_linux_<arch>.tar.gz.spdx.json \
  ./nodescope_<version>_linux_<arch>.tar.gz.spdx.json.sha256 \
  v<version> <40-or-64-character-source-revision>
```

The command must identify the expected public repository and succeed before any artifact is trusted. It binds each checksum sidecar to exactly the supplied artifact or SBOM filename and structurally validates the SPDX JSON before remote attestation verification. An Administrator should record the approved release tag and checksum in the NodeScope audit log. Do not use a moving `latest` URL as an unattended-update input.

For a manual Linux installation, also record the immutable source revision shown in the release evidence. The installer requires both the pinned release tag and source revision in addition to independently checked binary and unit hashes. It writes a root-owned `/var/lib/nodescope-installer/metadata/installed.env` record containing the installed release, revision, artifact hashes, and previous binary, unit, and metadata backup references. Preserve this record for rollback review; do not edit it manually.

## Staged update policy

Framework is the canary. Set an exact tag, archive URL, and checksum URL in the root-only `/etc/nodescope-agent/update.env` after a successful manual verification. The timer checks the approved manifest weekly; it does not discover or install arbitrary new releases. The verified update helper stages the archive, validates the checksum, runs `gh attestation verify`, preserves the previous binary, and atomically activates the candidate. It restarts the agent only after activation succeeds.

After Framework remains healthy through the configured observation period, the same exact release manifest may be approved for Asus. If the agent fails health checks, run:

```bash
sudo systemctl stop nodescope-agent
sudo /usr/local/bin/nodescope-update --rollback
sudo systemctl start nodescope-agent
```

The update workflow retains the failed candidate and restores the prior verified binary. Rollback and canary promotion must be recorded as NodeScope audit events when the control plane is active.

## Installation artifacts

The companion update installer verifies the `nodescope-update` binary, update service, and update timer from exact checksums before installing them. It does not create or expose credentials. The NodeScope agent bearer token remains in a systemd credential file and is unrelated to release verification.

## Deferred activation prerequisites

Automatic updates remain disabled until the internal CA, remote GitHub CLI authentication/verification behavior, exact release tags, and systemd unit validation are complete on Framework and Asus. These are deployment gates; the release build, staging, verification, and rollback code is already present.

## References

[1]: https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds "GitHub artifact attestations"
[2]: https://docs.docker.com/build/ci/github-actions/attestations/ "Docker SBOM and provenance attestations"
