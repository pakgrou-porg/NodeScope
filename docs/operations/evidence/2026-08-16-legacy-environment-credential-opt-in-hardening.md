# Legacy Environment-Credential Opt-In Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live host, credential, certificate, database, container, runtime, or deployment operation was performed. |
| Commit | [`acebc56e0dcf88db03e86e116c70ed2288b6fb57`](https://github.com/pakgrou-porg/NodeScope/commit/acebc56e0dcf88db03e86e116c70ed2288b6fb57) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31963100853`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31963100853). |
| Expected result | `NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL` must parse as a boolean before any credential source is chosen. A malformed opt-in must fail configuration loading explicitly, and a true opt-in must remain development-mode-only. |
| Observed result | Focused configuration testing passed for a malformed legacy opt-in with the expected explicit error. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | This control governs static configuration parsing only. It does not enable legacy credentials, enroll an agent, contact an ingestion replica, or test any protected host. Secret-file credentials remain the required production path, and live activation is separately authorized. |
| Rollback or recovery | Revert commit `acebc56` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> The exceptional environment-credential path now has an explicit, validated opt-in. A malformed value cannot quietly influence which secret source the agent considers.
