import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const repositoryRoot = resolve(import.meta.dirname, "..");
const packagePath = resolve(repositoryRoot, "package.json");
const licensePath = resolve(repositoryRoot, "LICENSE");
const readmePath = resolve(repositoryRoot, "README.md");
const dockerfilePath = resolve(repositoryRoot, "deploy/compose/Dockerfile");
const releaseWorkflowPath = resolve(repositoryRoot, ".github/workflows/release.yml");

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

const readme = await readFile(readmePath, "utf8");
if (!readme.includes("Apache License 2.0")) {
  console.error("README.md must declare the Apache License 2.0.");
  process.exit(1);
}

const dockerfile = await readFile(dockerfilePath, "utf8");
if (!dockerfile.includes('org.opencontainers.image.licenses="Apache-2.0"')) {
  console.error("Container image metadata must declare Apache-2.0.");
  process.exit(1);
}

const releaseWorkflow = await readFile(releaseWorkflowPath, "utf8");
if (!releaseWorkflow.includes("PROJECT_LICENSE: Apache-2.0") || !releaseWorkflow.includes("org.opencontainers.image.licenses=${{ env.PROJECT_LICENSE }}")) {
  console.error("Release workflow must propagate Apache-2.0 into published image metadata.");
  process.exit(1);
}

console.log("NodeScope license metadata consistently declares Apache-2.0 across package, README, container, and release workflow.");
