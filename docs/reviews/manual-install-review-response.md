# NodeScope manual installation guide review response

**Document status:** Draft hardening response, not an executable production runbook.  
**Classification:** Internal Restricted.  
**Author:** Manus AI.  
**Date:** 2026-07-23.  
**Scope:** Framework and Asus GX10 native-agent installation guidance, enrollment, database-operational access, and supporting implementation boundaries.

> **Conclusion.** The review is substantially correct. The original guide preserved several important NodeScope requirements—no database credentials on agents, explicit UMA versus dedicated-memory semantics, refusal to synthesize VRAM, and no-content retention—but it should not be used as a production runbook until the critical enrollment, credential, database-access, build-provenance, and installation-path defects are corrected.

## Disposition matrix

| Review finding | Disposition | Why | Required improvement |
| --- | --- | --- | --- |
| Re-enrollment generates a new host UUID before an `ON CONFLICT (slug)` upsert. | **Accepted — critical.** | The current guide can attempt to insert an agent against a non-canonical host ID during a repeat enrollment. | Replace ad hoc SQL with one restricted enrollment procedure that returns the canonical host ID. Keep the existing agent UUID stable for ordinary credential rotation, increment a rotation version, and audit the event. |
| Plaintext agent secret appears in `psql -v`, SQL substitution, terminal output, and manual environment-file input. | **Accepted — critical.** | This violates the intended per-host credential boundary even though the database stores only a digest. | Add a local enrollment utility that generates the secret, hashes it locally, sends only the digest to the database, and atomically writes the secret directly to a root-owned credential source. The secret must never appear in argv, SQL text, ordinary environment variables, logs, or terminal output. |
| Enrollment shell sequence can continue after a failed database action. | **Accepted — high.** | A failed write could produce an unusable local credential and misleading operator status. | Replace the ad hoc shell block with a fail-closed enrollment utility. Any temporary wrapper must use `set -euo pipefail`, `umask 077`, secure cleanup traps, atomic writes, and positive database confirmation before publishing the credential. |
| `sslmode=require` and `PGPASSWORD` are used for privileged administration. | **Accepted — high.** | PostgreSQL documents that `verify-full` additionally checks the expected server identity and discourages `PGPASSWORD` because some systems expose process environments.[1] [2] | Use `sslmode=verify-full`, explicit CA policy, and a temporary mode-`0600` `PGPASSFILE` or an interactive credential provider. Never export privileged database passwords as a durable process environment variable. |
| Routine verification and enrollment can assume `nodescope_owner`. | **Accepted — high.** | The migration owner is intentionally more privileged than day-to-day operations need. | Create narrowly scoped `nodescope_enroller`, `nodescope_verifier`, and `nodescope_storage_auditor` login paths. Restrict ordinary workflows to security-definer functions or views; reserve `nodescope_owner` for migration automation only. |
| Build instructions clone the moving default branch and run a mutating formatter. | **Accepted — high.** | The guide does not establish a reproducible source identity. | Require a signed release tag or approved full commit SHA, detached checkout, tag/commit verification, clean-tree check, non-mutating formatting validation, `go mod verify`, full tests, checksum, SBOM, and provenance evidence.[9] [10] [11] |
| Root installs from `/tmp` and a normal checkout. | **Accepted — high.** | The root installer currently trusts caller-controlled inputs and does not verify a checksum or stage atomically. | Harden the installer to reject symlinks and non-regular files, validate an expected SHA-256 before elevation, stage in a root-owned directory, atomically replace the binary, preserve prior binary/checksum, and support rollback. |
| Framework AMD GPU/NPU guidance implies Fedora package support. | **Accepted — high.** | Current AMD Ryzen AI Linux and ROCm materials establish Ubuntu 24.04 support for the relevant stack, not a qualified Fedora matrix.[6] [7] | Treat Framework CPU, RAM, storage, and generic Linux telemetry as supported. Label AMD SMI/XDNA/NPU telemetry on Fedora **experimental** until NodeScope qualifies and publishes an exact Fedora, kernel, firmware, ROCm, AMD-SMI, XRT, and XDNA matrix. |
| Docker-group membership is suggested for inventory. | **Accepted — high.** | Docker states that membership in the `docker` group grants root-level privileges.[4] | Disable Docker inventory by default. Use an approved narrow socket proxy or dedicated privileged helper with a fixed, read-only output schema; document container metadata sensitivity. Rootless Docker is an optional deployment design, not an in-place workaround.[5] |
| Secret is placed in a heredoc environment file. | **Accepted — high.** | Environment files and terminal history are weaker than intended for bearer credentials. | Use systemd credentials: `LoadCredential=`/`LoadCredentialEncrypted=` provides a per-service file rather than process-environment propagation; support a reduced-assurance fallback only after target systemd version review.[12] [13] |
| Only primary replica TLS is always tested; retry cadence lacks backoff and jitter. | **Accepted — high.** | Secondary failure would otherwise surface only during an outage, and synchronized retries can amplify faults. | Require bilateral TLS identity, certificate-expiry, `/healthz`, version, and authenticated non-mutating preflight checks. Implement exponential backoff with jitter and per-endpoint circuit breaking while preserving preferred-endpoint failback. |
| Verification query counts the outer-join row rather than metric rows and does not show clock skew. | **Accepted — medium.** | `count(*)` can report one for a host with no metric rows. | Use `count(m.metric_name)` or an equivalent non-null metric field and add an explicit `agent.clock_offset_seconds` query that shows value, quality, source, observation time, and receipt time. |
| Storage-feasibility evidence relies on agent timestamps and simplistic compressed-byte summaries. | **Accepted — high.** | The report correctly identifies clock, outage, peak, cardinality, and physical-storage blind spots. | Base the observation window on server `received_at`; calculate expected/received completeness, maximum gap, median/p95/p99 compressed size, cardinality, relation/index/TOAST deltas, rollup/deletion cost, and atomic evidence files named from the host slug. Keep the Free-plan threshold as a dated, sourced policy input rather than a permanent constant. |
| Credential lifecycle, identity binding, CA lifecycle, rate controls, upgrade, data-boundary, and systemd review need more detail. | **Accepted — planned.** | These are production acceptance concerns that extend beyond prose. | Add schema fields and tests for rotation/expiry/revocation/last-used audit; derive host identity from agent credential; add CA fingerprint/rotation controls; bound payloads/rates/cardinality; add rollback procedures; run serialization/log leak tests; and require `systemd-analyze verify` plus `systemd-analyze security` evidence.[14] |
| Asus UMA treatment must model `Memory-Usage: Not Supported` as expected. | **Accepted — already aligned in principle; implementation must be explicit.** | NVIDIA documents this as expected on iGPU UMA systems, even when per-process GPU memory is available.[8] | Emit a typed `not_supported` state for dedicated framebuffer metrics, not a generic collector error. Preserve the mandated side-by-side OS `MemAvailable`, `SwapFree`, huge-page, and per-process GPU-memory panel with provenance labels. |
| The rendered PDF has clipping, extraction, and metadata defects. | **Accepted, but deferred.** | The current deliverable is Markdown; no production PDF should be distributed before source/runbook hardening. | Maintain the canonical runbook as versioned Markdown. If a PDF is later requested, use a Unicode-capable generator, code wrapping, bookmarks, headers/footers, revision metadata, and an Internal Restricted label. |
| The report could not verify the repository URL. | **Modified.** | The public repository now exists at `https://github.com/pakgrou-porg/NodeScope`, but it does not yet publish a signed release artifact. | The revised runbook will reference the repository only as a development source. Production instructions will require a signed tag/release, attestation, checksum, and explicit approved revision. |

## Controls retained without change

The following original design commitments remain mandatory. Agent hosts receive no Supabase, database-runtime, or database-migrator credentials. Each host uses a distinct revocable agent credential. Inference prompts, responses, process arguments, environment variables, and raw tool output are never stored. Stale or unavailable data is explicit, never treated as zero, and GPU memory provenance is never blurred. The Asus UMA panel retains OS memory, swap, huge-page, and per-process GPU memory as separately labeled sources; dedicated VRAM is never invented when NVIDIA reports it as unsupported.

## Recommended implementation sequence

The first patch set should add an Internal Restricted document header, pin production build inputs, reclassify Fedora GPU/NPU collection, remove Docker-group enrollment advice, correct both-replica verification, and correct the verification/storage evidence queries. This reduces immediate operator risk without altering the shared Supabase project.

The second patch set should alter the agent configuration and systemd unit to consume a credential file rather than `NODESCOPE_AGENT_CREDENTIAL`, and harden the installer’s root trust boundary. It should include negative tests proving that the bearer token cannot appear in service logs, process environment diagnostics, or the non-secret configuration file.

The third patch set should add the least-privilege enrollment/verifier/auditor functions, credential lifecycle fields, and an enrollment utility. Because these are shared-project database changes, they must pass the dedicated migrator rollback preflight and sibling-schema noninterference gate before application.

The final validation set should use a real, qualified build on Framework and Asus, exercise both replica endpoints, run 72 hours of representative telemetry, and capture storage evidence using server receipt times and physical relation growth. Only then should the guide become an approved production runbook.

## References

[1] [PostgreSQL: SSL Support](https://www.postgresql.org/docs/current/libpq-ssl.html)

[2] [PostgreSQL: Environment Variables](https://www.postgresql.org/docs/current/libpq-envars.html)

[3] [PostgreSQL: Password File](https://www.postgresql.org/docs/current/libpq-pgpass.html)

[4] [Docker: Linux post-installation steps](https://docs.docker.com/engine/install/linux-postinstall/)

[5] [Docker: Rootless mode](https://docs.docker.com/engine/security/rootless/)

[6] [AMD: Ryzen AI Linux Installation Instructions](https://ryzenai.docs.amd.com/en/latest/linux.html)

[7] [AMD: ROCm Radeon and Ryzen Linux support matrix](https://rocm.docs.amd.com/projects/radeon-ryzen/en/latest/docs/compatibility/compatibilityryz/native_linux/native_linux_compatibility.html)

[8] [NVIDIA: DGX Spark known issues](https://docs.nvidia.com/dgx/dgx-spark/known-issues.html)

[9] [GitHub: Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)

[10] [GitHub: Signing tags](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-tags)

[11] [Go Modules Reference](https://go.dev/ref/mod)

[12] [systemd: System and Service Credentials](https://systemd.io/CREDENTIALS/)

[13] [systemd-creds manual](https://www.freedesktop.org/software/systemd/man/systemd-creds.html)

[14] [systemd-analyze manual](https://www.freedesktop.org/software/systemd/man/systemd-analyze.html)
