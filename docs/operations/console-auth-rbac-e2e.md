# Console Authentication, RBAC, and Degraded-Replica E2E Procedure

**Purpose:** validate the live NodeScope browser console against approved Supabase Auth and the deployed replica stack. This procedure is intentionally **not runnable without explicit authorization** because it creates invite-only test identities, sends magic links, and changes an approved replica availability condition.

> Local contracts prove the current Viewer and Administrator-equivalent route behavior. This procedure is the separate environment proof for real magic links, sessions, role mapping, browser callbacks, and degraded-replica behavior.

## Required approval and setup

Before starting, an administrator must approve the test email identities, allowed redirect URL, maintenance window, preferred-replica degradation method, and the scope of any Operator or Administrator test action. Do not reuse a personal primary-user account for destructive or availability testing. Create only named invite-only test identities and record their UUIDs in the evidence record, not in source control.

| Prerequisite | Verification | Stop condition |
| --- | --- | --- |
| Magic-link configuration | Supabase Auth invite-only mode, redirect allow-list, sender, and deployed console callback are reviewed. | Callback or sender configuration differs from the approved value. |
| Role mapping | Viewer, Operator, and Administrator assignments are present in the NodeScope role source of truth and reviewed. | Any test identity receives a broader role than approved. |
| Replica protection | The preferred and secondary replicas are healthy before the drill, with controlled restoration and monitoring available. | Either replica is already degraded or the secondary is not ready. |
| Audit boundary | Test labels, expected actions, and rollback owner are agreed before login. | The intended action could affect a non-test host, route, rule, or credential. |

## Browser E2E sequence

| Step | Actor and action | Expected result | Evidence to record |
| --- | --- | --- | --- |
| 1 | Viewer opens the approved magic link and completes callback. | Session is established only at the approved console URL. | Redirect URL, session success, timestamp, browser version, and test identity UUID. |
| 2 | Viewer reads fleet state and one host’s alerts. | Read views load; alerts are scoped to the selected host. | Screenshot and redacted network result codes. |
| 3 | Viewer attempts collection-interval change and runtime approval. | Both actions fail closed with an authorization error and create no operation. | UI result, response status, and audit query proving no action. |
| 4 | Operator performs only the pre-approved reversible action. | Action follows the documented role matrix and creates an audit event. | Action ID, audit ID, before/after state, and reversal evidence. |
| 5 | Administrator performs one pre-approved configuration action and reverses it. | Action and reversal are recorded with the correct actor and target scope. | Action ID, audit ID, reversal ID, and final-state query. |
| 6 | Preferred replica is placed into the approved temporary degraded condition. | Console reports degradation without exposing credentials; secondary path remains available. | Start/end timestamps, health outputs, browser behavior, and server logs redacted of secrets. |
| 7 | Preferred replica is restored. | Console regains preferred health; no unexpected privilege, session, or data change occurs. | Recovery timestamp, health output, and final audit query. |

## Acceptance evidence record

For every step, record the following fields in a new `docs/operations/evidence/` entry: source commit SHA, test command or browser procedure, environment and host/replica, expected result, observed result, evidence path, known limitation, and rollback or recovery result. The record must distinguish **environment validated** from **operationally accepted**; administrator acceptance is a separate decision after reviewing the complete evidence.

## Recovery and stop rules

If magic-link, callback, role, or replica behavior differs from expectation, stop the drill. Disable the affected invite, invalidate the test session, restore the prior callback or replica configuration, reverse the approved action, and retain only redacted diagnostics. Do not continue to another role or replica step until the failed invariant has a documented cause and recovery result.

## Local readiness command

Run `./scripts/rehearse-console-rbac-local.sh` before the live procedure. It validates the local Viewer/Administrator contract and route loading but deliberately does not create users, send email, contact Supabase Auth, start a browser session, or degrade a replica.
