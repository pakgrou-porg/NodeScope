import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const repositoryRoot = resolve(import.meta.dirname, "..");
const packagePath = resolve(repositoryRoot, "package.json");
const licensePath = resolve(repositoryRoot, "LICENSE");

const packageMetadata = JSON.parse(await readFile(packagePath, "utf8"));
if (packageMetadata.license !== "Apache-2.0") {
  console.error(`package.json license must be Apache-2.0, got ${JSON.stringify(packageMetadata.license)}.`);
  process.exit(1);
}

const licenseText = await readFile(licensePath, "utf8");
if (!licenseText.includes("Apache License") || !licenseText.includes("Version 2.0")) {
  console.error("LICENSE must contain the Apache License, Version 2.0 text.");
  process.exit(1);
}

console.log("NodeScope package license metadata matches the Apache-2.0 repository license.");
