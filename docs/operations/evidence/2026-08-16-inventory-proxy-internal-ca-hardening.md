# Inventory Proxy Internal-CA Requirement Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No live inventory proxy, Docker host, certificate authority, credential, runtime, database, or deployment operation was performed. |
| Commit | [`fc6286c410c690e332c4f8cc0f142073e5666701`](https://github.com/pakgrou-porg/NodeScope/commit/fc6286c410c690e332c4f8cc0f142073e5666701) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31963697874`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31963697874). |
| Expected result | Enabling the mTLS-protected inventory proxy must require a configured internal CA path in addition to the approved HTTPS proxy URL and paired client certificate/key paths. |
| Observed result | Focused configuration testing passed for both rejection of an enabled proxy without `NODESCOPE_CA_CERT_PATH` and acceptance of the complete CA-plus-client-certificate configuration. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control validates static configuration only. It does not issue or rotate certificates, connect to a proxy, enumerate Docker containers, perform mTLS handshakes, or qualify Framework or Asus. Those remain separately authorized environment gates. |
| Rollback or recovery | Revert commit `fc6286c` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the configuration. |

> The inventory proxy now has a complete static trust configuration: an approved HTTPS target, an internal CA trust anchor, and paired client credentials are all required before collection can be enabled.
