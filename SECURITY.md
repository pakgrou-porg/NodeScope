# Security Policy

## Reporting a vulnerability

Please do not open a public issue for a suspected vulnerability. Report it privately through the repository’s security advisory workflow or contact the maintainers through the address published in the repository profile. Include a clear reproduction path, affected version or commit, impact assessment, and any suggested mitigation.

Maintainers will acknowledge a well-formed report, assess it privately, and coordinate a remediation and disclosure timeline. Do not include credentials, private telemetry, prompts, responses, or backup contents in a report.

## Supported releases

Before the first stable release, only the latest `main` revision is supported. After Release 1, the project will publish a support table covering current stable, security-fix, and end-of-life versions.

## Security boundaries

NodeScope treats the following as critical boundaries: fleet-wide role authorization, agent and client credentials, Supabase service credentials, internal certificate material, operation audit records, backup leases, and the proxy’s no-content-retention guarantee. Changes affecting those boundaries require tests and maintainers’ review before merge.
