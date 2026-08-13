import { renderToStaticMarkup } from "react-dom/server";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { HostAlertSummaryPanel } from "./HostDetail";

describe("HostAlertSummaryPanel", () => {
  it("makes the fresh-only automatic alerting boundary explicit when no alerts exist", () => {
    const html = renderToStaticMarkup(createElement(HostAlertSummaryPanel, { alerts: [] }));

    expect(html).toContain("fresh metric evidence only");
    expect(html).toContain("stale, unavailable, estimated, and experimental evidence cannot open an alert");
    expect(html).toContain("No host alerts recorded");
  });

  it("distinguishes active and acknowledged operator states without weakening the quality guard", () => {
    const html = renderToStaticMarkup(createElement(HostAlertSummaryPanel, { alerts: [
      { id: "active", severity: "warning", state: "active", title: "Latency threshold", detail: "P95 TTFT remains above the preferred threshold.", observedAt: "2026-08-13T22:00:00.000Z" },
      { id: "acknowledged", severity: "info", state: "acknowledged", title: "Reviewed event", detail: "Operator acknowledgement is recorded.", observedAt: "2026-08-13T22:01:00.000Z" },
    ] }));

    expect(html).toContain("1 active alert requires operator attention");
    expect(html).toContain("active");
    expect(html).toContain("acknowledged");
    expect(html).toContain("does not change the evidence-quality guard");
  });
});
