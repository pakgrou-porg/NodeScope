import { renderToStaticMarkup } from "react-dom/server";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { HistoryPanel } from "./HostDetail";

describe("HistoryPanel", () => {
  it("renders explicit missing-history evidence rather than an ambiguous blank chart", () => {
    const html = renderToStaticMarkup(createElement(HistoryPanel, { history: [] }));

    expect(html).toContain("No retained telemetry samples");
    expect(html).toContain("does not synthesize chart points");
    expect(html).toContain("collection configuration");
    expect(html).not.toContain("retained samples</span>");
  });

  it("renders a labeled chart with a retained-sample count when history exists", () => {
    const html = renderToStaticMarkup(createElement(HistoryPanel, { history: [{ timestamp: "2026-08-13T22:00:00.000Z", throughput: 132, cpu: 41 }] }));

    expect(html).toContain("Generation throughput");
    expect(html).toContain("CPU utilization");
    expect(html).toContain("1 retained samples");
    expect(html).not.toContain("No retained telemetry samples");
  });
});
