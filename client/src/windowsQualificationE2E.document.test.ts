import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("Windows MSI qualification procedure", () => {
  it("keeps unsupported status, artifact, hardware, rollback, and recovery gates explicit", () => {
    const procedure = readFileSync(resolve(process.cwd(), "docs/operations/windows-msi-qualification-e2e.md"), "utf8");

    expect(procedure).toContain("operationally unsupported");
    expect(procedure).toContain("signed Windows artifact, installer, update/rollback rehearsal");
    expect(procedure).toContain("RTX 5080");
    expect(procedure).toContain("LM Studio");
    expect(procedure).toContain("Keep Windows enrollment disabled");
    expect(procedure).toContain("rehearse-windows-baseline-local.sh");
  });
});
