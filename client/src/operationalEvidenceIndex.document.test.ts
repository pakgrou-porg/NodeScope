import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("operational evidence index", () => {
  it("keeps every required proof field and the live acceptance boundary explicit", () => {
    const index = readFileSync(resolve(process.cwd(), "docs/operations/evidence/index.md"), "utf8");

    expect(index).toContain("Commit");
    expect(index).toContain("Validation command or procedure");
    expect(index).toContain("Environment");
    expect(index).toContain("Expected and observed result");
    expect(index).toContain("Evidence location");
    expect(index).toContain("Known limitation");
    expect(index).toContain("Rollback or recovery");
    expect(index).toContain("not operationally accepted");
    expect(index).toContain("Migration `0015_terminal_fleet_status` apply");
    expect(index).toContain("Current cloud sandbox compose preflight");
    expect(index).toContain("Ordered console usability series");
    expect(index).toContain("Local aggregate readiness report");
    expect(index).toContain("verify-operational-evidence-index.sh");
  });
});
