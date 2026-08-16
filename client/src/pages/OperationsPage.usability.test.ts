import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("operations page usability", () => {
  it("keeps action requirements, preview safety, runtime review states, and endpoint withholding explicit", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/OperationsPage.tsx"), "utf8");

    expect(source).toContain('aria-label="Operation requirements"');
    expect(source).toContain("These controls are role-protected, preview-safe, and server-audited.");
    expect(source).toContain("runtimeCandidates.length");
    expect(source).toContain("No discovered runtimes awaiting review");
    expect(source).toContain("Discovered endpoint location intentionally withheld");
    expect(source).not.toContain("{runtime.endpoint}</p>");
    expect(source).toContain("disabled={preview || approveRuntime.isPending}");
    expect(source).toContain("acknowledgedDiff: baselineAcknowledged");
  });
});
