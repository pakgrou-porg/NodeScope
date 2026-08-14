import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("alert triage navigation", () => {
  it("keeps the affected host name as a drill-down control to the host evidence view", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/AlertsPage.tsx"), "utf8");

    expect(source).toContain('navigate(targetPath(`/hosts/${alert.hostId}`))');
    expect(source).toContain("host?.name ?? alert.hostId");
  });
});
