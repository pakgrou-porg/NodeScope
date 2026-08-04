# Manual installation guide: authoritative validation findings

This note records the sources used to validate the report on the Framework and Asus manual-installation guide.

| Area | Verified finding | Runbook consequence |
| --- | --- | --- |
| PostgreSQL server identity | `sslmode=verify-full` verifies the trust chain and server hostname/SAN; `require` does not provide that hostname check. | The guide must use `verify-full` and an explicit root certificate for privileged database administration. |
| PostgreSQL password handling | PostgreSQL discourages `PGPASSWORD` because some systems expose process environments. Its password-file mechanism supports `PGPASSFILE` and requires mode `0600` on Unix. | Replace environment-exported passwords with a temporary, mode-`0600` passfile or an interactive prompt. |
| Docker socket authority | Docker states that membership in the `docker` group grants root-level privileges. | Docker inventory must default to disabled; the guide must not recommend group membership as a routine agent prerequisite. |
| Rootless Docker | Rootless mode runs the daemon and containers without root to reduce daemon/runtime exposure, subject to its documented prerequisites. | Treat rootless Docker as a deployment-specific option; do not imply it is an in-place safe conversion of an existing rootful installation. |
| Framework AMD support | AMD’s current Ryzen AI Linux guide packages NPU drivers for Ubuntu 24.04, and the ROCm 7.2.1 Ryzen/AI Max support matrix lists Ubuntu 24.04.4 as supported. | Framework GPU/NPU collection on Fedora must be labelled experimental until NodeScope qualifies an exact Fedora/kernel/firmware/ROCm/XRT/AMD-SMI matrix. CPU, RAM, storage, and standard Linux telemetry remain supported. |
| Asus DGX Spark UMA | NVIDIA documents that `nvidia-smi` may show `Memory-Usage: Not Supported` on iGPU platforms even while per-process GPU memory is listed, because there is no dedicated framebuffer memory. NVIDIA specifically recommends OS `MemAvailable`, `SwapFree`, and huge-page information for UMA-aware assessment. | The Asus collector and console must emit a typed `not_supported`/unavailable dedicated-framebuffer state while retaining separately labeled OS UMA and per-process GPU values. |

## References

[1] [PostgreSQL: SSL Support](https://www.postgresql.org/docs/current/libpq-ssl.html)

[2] [PostgreSQL: Environment Variables](https://www.postgresql.org/docs/current/libpq-envars.html)

[3] [PostgreSQL: Password File](https://www.postgresql.org/docs/current/libpq-pgpass.html)

[4] [Docker: Linux post-installation steps](https://docs.docker.com/engine/install/linux-postinstall/)

[5] [Docker: Rootless mode](https://docs.docker.com/engine/security/rootless/)

[6] [AMD: Ryzen AI Linux Installation Instructions](https://ryzenai.docs.amd.com/en/latest/linux.html)

[7] [AMD: ROCm Radeon and Ryzen Linux support matrix](https://rocm.docs.amd.com/projects/radeon-ryzen/en/latest/docs/compatibility/compatibilityryz/native_linux/native_linux_compatibility.html)

[8] [NVIDIA: DGX Spark known issues](https://docs.nvidia.com/dgx/dgx-spark/known-issues.html)

| Build provenance | GitHub artifact attestations record the workflow, repository, commit SHA, and triggering event; they can include SBOMs. Signed tags can be verified locally with `git tag -v`. | The runbook must pin a signed tag or full approved commit SHA, verify a clean tree and Go modules without mutating source, then record release checksums/provenance. |

[9] [GitHub: Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)

[10] [GitHub: Signing tags](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-tags)

[11] [Go Modules Reference](https://go.dev/ref/mod)

| Service credential handling | systemd credentials are acquired at activation, made available to the service as files, not propagated like environment variables, and can optionally be encrypted/authenticated with TPM2 or host-bound keys. `LoadCredential=` provides a file-backed credential and `%d` resolves to the per-service credential directory. | Replace the agent bearer token in `EnvironmentFile=` with a credential file delivered through `LoadCredential=`; support `LoadCredentialEncrypted=` where the target systemd version and operator policy allow it. |
| Unit review | `systemd-analyze security UNIT` is an official command for analyzing unit security/sandboxing posture. | The installation acceptance procedure must run `systemd-analyze verify` and `systemd-analyze security nodescope-agent.service`, recording any justified exemptions. |

[12] [systemd: System and Service Credentials](https://systemd.io/CREDENTIALS/)

[13] [systemd-creds manual](https://www.freedesktop.org/software/systemd/man/systemd-creds.html)

[14] [systemd-analyze manual](https://www.freedesktop.org/software/systemd/man/systemd-analyze.html)
