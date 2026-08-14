import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("operational release ledger", () => {
  it("keeps evidence-state definitions, dependency order, and recovery paths explicit", () => {
    const ledger = readFileSync(resolve(process.cwd(), "docs/operations/release-epics.md"), "utf8");

    expect(ledger).toContain("**Implemented**");
    expect(ledger).toContain("**Locally validated**");
    expect(ledger).toContain("**Environment validated**");
    expect(ledger).toContain("**Operationally accepted**");
    expect(ledger).toContain("Operational acceptance is intentionally empty");
    expect(ledger).toContain("Framework Linux primary canary");
    expect(ledger).toContain("Windows remains unsupported operationally");
    expect(ledger).toContain("Metadata-only inference privacy");
    expect(ledger).toContain("## Dependency order");
    expect(ledger).toContain("Rollback or recovery");
    expect(ledger).toContain("Local resilience rehearsal");
    expect(ledger).toContain("certificate revocation and isolated restore remain live gates");
    expect(ledger).toContain("Machine-readable preparation");
    expect(ledger).toContain("release manifest assembler");
  });
});
