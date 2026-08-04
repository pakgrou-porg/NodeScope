import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, rmSync, statSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";

const root = process.cwd();
const schemaPath = join(root, "telemetry", "v1", "envelope.proto");
const descriptorPath = join(tmpdir(), `nodescope-envelope-${process.pid}.pb`);
const schema = readFileSync(schemaPath, "utf8");

const requiredFragments = [
  'syntax = "proto3";',
  "package nodescope.telemetry.v1;",
  "option go_package = \"github.com/pakgrou-porg/nodescope/telemetry/v1;telemetryv1\";",
  "uint32 schema_version = 1;",
  "string agent_id = 3;",
  "uint64 sequence = 6;",
  "bytes checksum_sha256 = 13;",
  "METRIC_QUALITY_UNAVAILABLE = 3;",
  "METRIC_QUALITY_UNSUPPORTED = 4;",
];

for (const fragment of requiredFragments) {
  if (!schema.includes(fragment)) {
    throw new Error(`telemetry contract is missing required fragment: ${fragment}`);
  }
}

try {
  execFileSync(
    "protoc",
    [
      "--proto_path=.",
      `--descriptor_set_out=${descriptorPath}`,
      "--include_imports",
      "telemetry/v1/envelope.proto",
    ],
    { cwd: root, stdio: "inherit" },
  );

  if (!existsSync(descriptorPath) || statSync(descriptorPath).size === 0) {
    throw new Error("protoc did not produce a non-empty descriptor set");
  }

  console.log("NodeScope telemetry protobuf contract is valid.");
} finally {
  rmSync(descriptorPath, { force: true });
}
