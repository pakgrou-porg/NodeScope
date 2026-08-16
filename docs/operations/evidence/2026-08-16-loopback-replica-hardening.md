# Development-Gated Loopback Ingestion Replica Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live replica configuration was changed. |
| Commit | [`724e28864a6f2d1b40576000eb7d1779e6476c67`](https://github.com/pakgrou-porg/NodeScope/commit/724e28864a6f2d1b40576000eb7d1779e6476c67) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. No Framework, Asus, Portainer stack, credential, certificate, database, or inference backend was contacted. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31952525146`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31952525146). |
| Expected result | Primary and secondary ingestion endpoints using IPv4, IPv6, or named loopback hosts must fail by default. A loopback replica is accepted only when both `NODESCOPE_ALLOW_LOOPBACK_REPLICA_ENDPOINTS=true` and `NODESCOPE_DEVELOPMENT_MODE=true` are present; malformed opt-in values must fail. Existing authenticated LAN HTTPS replica endpoints must remain valid. |
| Observed result | Focused native-agent tests passed for IPv4, IPv6, and `localhost` rejection, dual development opt-in, and malformed opt-in rejection. The aggregate readiness suite passed. Continuous Integration completed successfully across secret scanning, web console validation, contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates configuration only. It does not prove Framework or Asus connectivity, mTLS enrollment, dual-replica failover, certificate rotation, or authenticated telemetry ingestion. Those remain explicit owner-authorized environment gates. |
| Rollback or recovery | Revert commit `724e288` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun the focused tests, aggregate readiness suite, and CI before permitting any development-only loopback configuration. |

> A development-only loopback opt-in is not an operational exception. Production agents continue to require configured non-loopback HTTPS ingestion replicas.
