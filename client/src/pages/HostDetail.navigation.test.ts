import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("host-detail navigation and summary", () => {
  it("keeps fleet and directory navigation explicit while exposing provenance-aware host summary counts", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/HostDetail.tsx"), "utf8");

    expect(source).toContain('navigate(`${gateway}/hosts`)');
    expect(source).toContain("Host evidence summary");
    expect(source).toContain("activeAlertCount");
    expect(source).toContain("nonFreshMetricCount");
    expect(source).toContain("preflightAttentionCount");
    expect(source).toContain('aria-label="Host detail sections"');
    expect(source).toContain("Provenance remains visible; no value is inferred");
    expect(source).toContain('`Alerts (${activeAlertCount})`');
  });
});
