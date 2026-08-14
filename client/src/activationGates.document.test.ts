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

  it("keeps the live read-only Supabase preflight record explicit about verified scope and remaining controls", () => {
    const evidence = readFileSync(resolve(process.cwd(), "docs/operations/evidence/2026-08-13-shared-supabase-readonly-preflight.md"), "utf8");

    expect(evidence).toContain("Database changes made:** none");
    expect(evidence).toContain("nodescope_runtime_login");
    expect(evidence).toContain("nodescope_migrate_login");
    expect(evidence).toContain("BEGIN READ ONLY");
    expect(evidence).toContain("Sibling-schema denial");
    expect(evidence).toContain("No migration");
  });

  it("keeps the live sibling-schema denial evidence explicit about both identities and cleanup", () => {
    const evidence = readFileSync(resolve(process.cwd(), "docs/operations/evidence/2026-08-13-sibling-schema-denial-gate.md"), "utf8");

    expect(evidence).toContain("nodescope_runtime_login");
    expect(evidence).toContain("nodescope_migrate_login");
    expect(evidence).toContain("function replacement");
    expect(evidence).toContain("fixture_cleanup=PASSED");
    expect(evidence).toContain("Production DDL requires a distinct explicit approval");
  });
});
