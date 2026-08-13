import { renderToStaticMarkup } from "react-dom/server";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { MetricRow } from "./HostDetail";

const offsetSemantics = "server receipt time minus agent observation time; stale when absolute offset exceeds 60 seconds";

describe("agent clock-offset metric presentation", () => {
  it("shows fresh receipt-time offset evidence with provenance", () => {
    const html = renderToStaticMarkup(createElement(MetricRow, { metric: { label: "Agent clock offset", display: "+1.3 s", quality: "fresh", source: "nodescope-server", semantics: offsetSemantics } }));

    expect(html).toContain("Agent clock offset");
    expect(html).toContain("+1.3 s");
    expect(html).toContain("fresh evidence");
    expect(html).toContain("Source: nodescope-server");
    expect(html).toContain(offsetSemantics);
  });

  it("keeps excessive offsets explicitly stale rather than presenting them as fresh timing evidence", () => {
    const html = renderToStaticMarkup(createElement(MetricRow, { metric: { label: "Agent clock offset", display: "−74.2 s", quality: "stale", source: "nodescope-server", semantics: offsetSemantics } }));

    expect(html).toContain("−74.2 s");
    expect(html).toContain("stale evidence");
    expect(html).not.toContain("fresh evidence");
  });
});
