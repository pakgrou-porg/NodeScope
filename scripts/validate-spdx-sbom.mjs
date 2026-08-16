#!/usr/bin/env node
import { readFileSync } from "node:fs";

function fail(message) {
  console.error(`SPDX validation failed: ${message}`);
  process.exit(1);
}

if (process.argv.length !== 3) {
  console.error(`usage: ${process.argv[1]} <sbom.spdx.json>`);
  process.exit(2);
}

const sbomPath = process.argv[2];
let sbom;
try {
  sbom = JSON.parse(readFileSync(sbomPath, "utf8"));
} catch (error) {
  fail(`cannot parse JSON from ${sbomPath}: ${error.message}`);
}

if (!sbom || typeof sbom !== "object" || Array.isArray(sbom)) fail("top-level value must be an object");
if (typeof sbom.spdxVersion !== "string" || !/^SPDX-2\.[0-9]+$/.test(sbom.spdxVersion)) {
  fail("spdxVersion must be an SPDX-2.x string");
}
if (!Array.isArray(sbom.packages) || sbom.packages.length === 0) fail("packages must be a non-empty array");
for (const [index, pkg] of sbom.packages.entries()) {
  if (!pkg || typeof pkg !== "object" || Array.isArray(pkg)) fail(`packages[${index}] must be an object`);
  if (typeof pkg.name !== "string" || pkg.name.trim() === "") fail(`packages[${index}].name must be a non-empty string`);
  if (typeof pkg.SPDXID !== "string" || !/^SPDXRef-[A-Za-z0-9._-]+$/.test(pkg.SPDXID)) {
    fail(`packages[${index}].SPDXID must be an SPDXRef identifier`);
  }
}

console.log(`SPDX SBOM parsed and structurally verified: ${sbomPath}`);
