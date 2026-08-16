#!/usr/bin/env node
import { readFileSync } from "node:fs";

const expectedCheckIds = new Set([
  "source_and_policy",
  "shared_schema_safety",
  "local_resilience",
  "native_builds",
  "browser_console",
]);

const expectedLiveGates = new Set([
  "No live Framework hardware qualification",
  "No live dual-replica failover, certificate revocation, or isolated restore acceptance",
  "No real Supabase magic-link, approved-backend streaming, or tagged release-attestation verification",
]);

function fail(message) {
  console.error(`readiness report validation failed: ${message}`);
  process.exit(1);
}

function requireNonEmptyString(value, path) {
  if (typeof value !== "string" || value.trim() === "") {
    fail(`${path} must be a non-empty string`);
  }
}

if (process.argv.length !== 3) {
  console.error(`usage: ${process.argv[1]} <report.json>`);
  process.exit(2);
}

const reportPath = process.argv[2];
let report;
try {
  report = JSON.parse(readFileSync(reportPath, "utf8"));
} catch (error) {
  fail(`cannot parse JSON from ${reportPath}: ${error.message}`);
}

if (!report || typeof report !== "object" || Array.isArray(report)) {
  fail("top-level value must be an object");
}
if (report.schema_version !== 2) fail("schema_version must equal 2");
if (report.scope !== "local deterministic release readiness") fail("scope is not local deterministic release readiness");
if (report.result !== "passed") fail("result must equal passed");
if (!/^[0-9a-f]{40}$/i.test(report.commit_sha ?? "")) fail("commit_sha must be a 40-character hexadecimal revision");
if (!Number.isSafeInteger(report.commit_timestamp_unix) || report.commit_timestamp_unix <= 0) fail("commit_timestamp_unix must be a positive integer");
if (report.working_tree_clean !== true) fail("working_tree_clean must equal true");

if (!Array.isArray(report.checks) || report.checks.length !== expectedCheckIds.size) {
  fail(`checks must contain exactly ${expectedCheckIds.size} entries`);
}
const observedCheckIds = new Set();
for (const check of report.checks) {
  if (!check || typeof check !== "object" || Array.isArray(check)) fail("each checks entry must be an object");
  requireNonEmptyString(check.id, "checks[].id");
  if (!expectedCheckIds.has(check.id)) fail(`checks contains an unexpected id: ${check.id}`);
  if (observedCheckIds.has(check.id)) fail(`checks contains a duplicate id: ${check.id}`);
  observedCheckIds.add(check.id);
  for (const field of ["command", "expected", "evidence_boundary", "known_limitation", "recovery"]) {
    requireNonEmptyString(check[field], `checks[${check.id}].${field}`);
  }
  if (check.observed !== "passed") fail(`checks[${check.id}].observed must equal passed`);
}
for (const id of expectedCheckIds) {
  if (!observedCheckIds.has(id)) fail(`checks is missing required id: ${id}`);
}

if (!Array.isArray(report.evidence_locations) || report.evidence_locations.length === 0) {
  fail("evidence_locations must be a non-empty array");
}
for (const location of report.evidence_locations) requireNonEmptyString(location, "evidence_locations[]");

if (!Array.isArray(report.live_gates_retained) || report.live_gates_retained.length !== expectedLiveGates.size) {
  fail(`live_gates_retained must contain exactly ${expectedLiveGates.size} entries`);
}
const observedLiveGates = new Set();
for (const gate of report.live_gates_retained) {
  requireNonEmptyString(gate, "live_gates_retained[]");
  if (!expectedLiveGates.has(gate)) fail(`live_gates_retained contains an unexpected gate: ${gate}`);
  if (observedLiveGates.has(gate)) fail(`live_gates_retained contains a duplicate gate: ${gate}`);
  observedLiveGates.add(gate);
}
for (const gate of expectedLiveGates) {
  if (!observedLiveGates.has(gate)) fail(`live_gates_retained is missing required gate: ${gate}`);
}
requireNonEmptyString(report.recovery, "recovery");

console.log(`Release-readiness report parsed and semantically verified: ${reportPath}`);
