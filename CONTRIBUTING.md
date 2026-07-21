# Contributing to NodeScope

Thank you for helping improve NodeScope. The project aims to provide accurate, privacy-preserving fleet observability across heterogeneous local compute systems.

## Development principles

Contributions must preserve the project’s core invariants: do not synthesize memory values, label metric provenance and quality, render stale or unavailable data explicitly, enforce authorization server-side, and never retain inference prompt or response content.

Use small, focused pull requests. Add or update unit tests for every behavioral change, add regression fixtures for platform/collector changes, and document any new configuration, security, migration, or operational behavior. Avoid changing generated files manually unless the repository instructions explicitly say otherwise.

## Local validation

Before opening a pull request, run the repository’s formatting, static analysis, unit-test, contract-test, and secret-scan commands. Cross-platform changes must pass the applicable AMD64 and ARM64 build workflows. Hardware-dependent collectors require fixture tests and a documented live-hardware validation result.

## Security and secrets

Never submit credentials, private keys, actual telemetry exports, backup files, prompt/response content, or secret-bearing configuration. Report vulnerabilities privately through the process in [SECURITY.md](SECURITY.md), not through a public issue.

## Design changes

Changes to data retention, permission enforcement, proxy privacy, internal PKI, backup fencing, telemetry envelope compatibility, or host platform semantics require an architecture decision record in `docs/decisions/` and associated tests.
