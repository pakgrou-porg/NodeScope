import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("console authentication and RBAC E2E procedure", () => {
  it("keeps live identity, role, degraded-replica, evidence, and recovery boundaries explicit", () => {
    const procedure = readFileSync(resolve(process.cwd(), "docs/operations/console-auth-rbac-e2e.md"), "utf8");

    expect(procedure).toContain("not runnable without explicit authorization");
    expect(procedure).toContain("invite-only test identities");
    expect(procedure).toContain("Viewer, Operator, and Administrator");
    expect(procedure).toContain("Preferred replica is placed into the approved temporary degraded condition");
    expect(procedure).toContain("source commit SHA");
    expect(procedure).toContain("Disable the affected invite");
    expect(procedure).toContain("rehearse-console-rbac-local.sh");
  });
});
