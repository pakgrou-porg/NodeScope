import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("alert queue filters", () => {
  it("keeps state and severity filtering, result counts, empty-state recovery, host drill-down, and acknowledgement explicit", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/AlertsPage.tsx"), "utf8");

    expect(source).toContain('aria-label="Filter alerts by state"');
    expect(source).toContain('aria-label="Filter alerts by severity"');
    expect(source).toContain('Showing {filteredAlerts.length} of {alerts.length} alerts.');
    expect(source).toContain("No alerts match the current queue filters.");
    expect(source).toContain("Clear alert filters");
    expect(source).toContain('targetPath(`/hosts/${alert.hostId}`)');
    expect(source).toContain("acknowledge.mutate({ alertId: alert.id })");
    expect(source).toContain('preview ? "Preview only" : "Acknowledge"');
  });
});
