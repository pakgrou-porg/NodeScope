#!/usr/bin/env node
import { readFileSync } from "node:fs";

const safeName = /^[A-Za-z0-9._-]+$/;
const archiveName = /^[A-Za-z0-9._-]+\.(?:tar\.gz|zip)$/;
const sbomName = /^[A-Za-z0-9._-]+\.spdx\.json$/;
const releaseTag = /^v[0-9][0-9A-Za-z._-]*$/;
const revision = /^[0-9a-f]{40}$/i;
const digest = /^[0-9a-f]{64}$/i;

function fail(message) {
  console.error(`release evidence manifest validation failed: ${message}`);
  process.exit(1);
}

function nonEmptyString(value, path) {
  if (typeof value !== "string" || value.trim() === "") fail(`${path} must be a non-empty string`);
}

if (process.argv.length !== 3) {
  console.error(`usage: ${process.argv[1]} <release-evidence.json>`);
  process.exit(2);
}

const manifestPath = process.argv[2];
let manifest;
try {
  manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
} catch (error) {
  fail(`cannot parse JSON from ${manifestPath}: ${error.message}`);
}

if (!manifest || typeof manifest !== "object" || Array.isArray(manifest)) fail("top-level value must be an object");
if (manifest.schema_version !== 1) fail("schema_version must equal 1");
if (!releaseTag.test(manifest.release_tag ?? "")) fail("release_tag must be a canonical v-prefixed tag");
if (!revision.test(manifest.commit_sha ?? "")) fail("commit_sha must be a 40-character hexadecimal revision");

if (!Array.isArray(manifest.artifacts) || manifest.artifacts.length === 0) fail("artifacts must be a non-empty array");
const artifactNames = new Set();
for (const artifact of manifest.artifacts) {
  if (!artifact || typeof artifact !== "object" || Array.isArray(artifact)) fail("each artifact must be an object");
  if (!archiveName.test(artifact.name ?? "") || !safeName.test(artifact.name)) fail("artifact name is unsafe or unsupported");
  if (artifactNames.has(artifact.name)) fail(`artifacts contains a duplicate name: ${artifact.name}`);
  artifactNames.add(artifact.name);
  if (!digest.test(artifact.sha256 ?? "")) fail(`artifacts[${artifact.name}].sha256 must be a SHA-256 digest`);
  if (artifact.checksum !== `${artifact.name}.sha256`) fail(`artifacts[${artifact.name}].checksum must name its exact sidecar`);
}

if (!Array.isArray(manifest.sboms) || manifest.sboms.length === 0) fail("sboms must be a non-empty array");
const sbomNames = new Set();
for (const sbom of manifest.sboms) {
  if (!sbomName.test(sbom ?? "") || !safeName.test(sbom)) fail("SBOM name is unsafe or unsupported");
  if (sbomNames.has(sbom)) fail(`sboms contains a duplicate name: ${sbom}`);
  sbomNames.add(sbom);
}

for (const field of ["provenance", "signing", "verification"]) nonEmptyString(manifest[field], field);

console.log(`Release evidence manifest parsed and semantically verified: ${manifestPath}`);
