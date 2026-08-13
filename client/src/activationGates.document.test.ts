import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("activation-gate register", () => {
  it("keeps the mandatory live deployment controls and fail-closed boundaries documented", () => {
    const document = readFileSync(resolve(process.cwd(), "docs/operations/activation-gates.md"), "utf8");

    for (const requiredGate of [
      "Shared Supabase isolation",
      "Internal PKI and credentials",
      "Primary and secondary replicas",
      "Storage and retention qualification",
      "Platform qualification",
      "Backup and recovery",
      "Release activation",
    ]) {
      expect(document).toContain(requiredGate);
    }
    expect(document).toContain("72-hour receipt-time benchmark");
    expect(document).toContain("Do not allow HTTP fallback");
    expect(document).toContain("Do not work around a stop condition");
  });
});
