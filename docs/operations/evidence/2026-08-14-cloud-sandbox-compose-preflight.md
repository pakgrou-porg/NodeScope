# Current Cloud Sandbox Compose Preflight Record

**Environment:** current NodeScope cloud sandbox. **Scope:** validation-only execution of `scripts/preflight-cloud-replica-compose.sh`. **Deployment:** none.

| Check | Expected result | Observed result |
| --- | --- | --- |
| Docker Engine discovery | The preflight must fail before inspecting staging paths or Compose configuration when Docker is unavailable. | **Passed fail-closed.** `docker` was not installed in the current sandbox. |
| Preflight exit | The process must return nonzero with a concise prerequisite diagnostic and must not start, build, pull, or remove containers. | **Passed.** Exit code `1` with `docker is required for replica preflight`. |
| Sensitive staging | The test must not create runtime secret, certificate, key, environment, or container artifacts. | **Passed.** No staging inputs were supplied or created. |

## Interpretation

The current cloud sandbox is suitable for the deterministic local contract and disposable control-plane canary already recorded, but it is **not a Docker-capable cloud host for the compose server-stack canary**. This result proves only that the preflight rejects an unsuitable environment before mutation. It does not validate a cloud replica, image build, TLS certificate, runtime secret, network endpoint, or health check.

## Next requirement

Use a separately approved Docker-capable cloud host or Cloud Computer with Compose v2. Stage the approved checkout, protected `replica.env`, runtime secret, and issued server certificate, then rerun the preflight before requesting authorization to start one replica.

## Recovery

No recovery is needed because no service or file was created. Do not install Docker into this sandbox as a substitute for an approved deployment host; doing so would not establish the persistent host, network, and operational ownership required for the replica canary.
