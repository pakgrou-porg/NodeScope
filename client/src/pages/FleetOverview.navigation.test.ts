import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("fleet overview navigation", () => {
  it("routes every posture metric to an actionable host, alert, or operations view", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/FleetOverview.tsx"), "utf8");

    expect(source).toContain('targetPath("/hosts")');
    expect(source).toContain('targetPath("/alerts")');
    expect(source).toContain('targetPath("/operations")');
    expect(source).toContain("Choose a host");
    expect(source).toContain("Open view");
    expect(source).toContain("Open alert details");
    expect(source).toContain("Inspect host");
    expect(source).toContain("Inference observability");
  });
});
