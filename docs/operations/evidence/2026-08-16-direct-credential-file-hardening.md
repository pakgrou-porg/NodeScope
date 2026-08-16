# Direct Credential-File Hardening

| Evidence field | Record |
| --- | --- |
| Evidence state | **Locally validated and CI validated.** No credential, protected host, database, certificate, container, or deployment operation was performed. |
| Commit | [`74447f90f7d64112cfeef1bfba6a8b13c3824c3e`](https://github.com/pakgrou-porg/NodeScope/commit/74447f90f7d64112cfeef1bfba6a8b13c3824c3e) |
| Environment | Local NodeScope sandbox checkout and GitHub-hosted continuous integration. |
| Validation commands and procedure | `go test ./internal/agent`; `./scripts/release-readiness-check.sh`; GitHub Actions [Continuous Integration run `31953303803`](https://github.com/pakgrou-porg/NodeScope/actions/runs/31953303803). |
| Expected result | Agent credential loading must reject a directory, symlink, or other non-regular filesystem object before permission checks or secret reads. A direct regular credential file must retain the existing POSIX permission guard; Windows compatibility must remain intact. |
| Observed result | Focused agent tests passed for direct regular-file acceptance, directory rejection, POSIX symlink rejection, and existing permission behavior. Aggregate readiness passed. Continuous Integration completed successfully across secret scanning, web console validation, API and telemetry contracts, Linux Go tests/vet/builds for AMD64 and ARM64, Windows cross-builds, and Windows runtime tests. |
| Evidence location | This record; [`config.go`](../../../internal/agent/config.go); [`config_test.go`](../../../internal/agent/config_test.go); and the linked GitHub workflow run. |
| Known limitation | The control reduces accidental or configured symlink/non-regular credential paths but is not a replacement for OS-level credential-file ownership, permission, service-account, or protected-environment enrollment controls. No real credential or host enrollment was exercised. |
| Rollback or recovery | Revert commit `74447f9` with a reviewed follow-up commit; do not rewrite shared history. Restore the prior accepted revision through the normal rollback process and rerun focused tests, aggregate readiness, and CI before reusing the credential-loading path. |

> Native agents load credentials only from direct regular files. Secret material remains excluded from source control, logs, evidence, and preflight reports.
