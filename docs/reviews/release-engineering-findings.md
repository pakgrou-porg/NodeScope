# NodeScope Release Engineering Findings

GitHub Documents that binary provenance attestations require `id-token: write`, `contents: read`, and `attestations: write`, and can be generated using `actions/attest@v4` with a `subject-path`. Container image attestations additionally require `packages: write` when publishing to a registry and use image name plus digest.[1]

Docker documents that `docker/build-push-action` version 4 or newer can emit image provenance, and that `sbom: true` emits an SBOM attestation. Public repository image builds receive maximum provenance by default, but NodeScope will set `provenance: mode=max` explicitly. No image build argument may contain a secret because public provenance can expose build-argument values.[2]

| NodeScope control | Implementation consequence |
|---|---|
| Native binaries | Build per Linux architecture, generate checksums and SPDX SBOMs, attest every binary, and upload the release assets. |
| Replica image | Build one `linux/amd64,linux/arm64` manifest to GHCR, push it, emit SBOM plus max-level provenance, and attest the pushed digest. |
| Secrets | Pass no secret as a Docker build argument. Runtime credentials are volume or environment-file inputs only. |
| Verification | Publish `gh attestation verify` commands in the installation guide and require verification before production agent installation. |

## Local Workflow Policy Guards

`scripts/release-readiness-check.sh` validates both release and continuous-integration workflow contracts without needing a tag push or GitHub-hosted runner. The CI contract requires a real `windows-2022` job to execute `go test ./internal/agent`, retains AMD64 and ARM64 Windows cross-build compilation, and prevents a second `pnpm/action-setup` version from drifting away from the repository-pinned `packageManager` declaration. Its companion negative fixture proves that a conflicting pnpm version is rejected.

These checks are a supplement to, not a replacement for, GitHub Actions execution. The protected workflow remains the evidence that the native Windows test actually ran and that the web, API-contract, Go, cross-build, and secret-scan jobs passed.

[1]: https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds "GitHub artifact attestations"
[2]: https://docs.docker.com/build/ci/github-actions/attestations/ "Docker SBOM and provenance attestations"
