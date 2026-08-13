import { renderToStaticMarkup } from "react-dom/server";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { RuntimeInventoryPanel } from "./HostDetail";

describe("RuntimeInventoryPanel", () => {
  it("renders a meaningful empty state without a browser-facing endpoint field", () => {
    const html = renderToStaticMarkup(createElement(RuntimeInventoryPanel, { runtimes: [] }));

    expect(html).toContain("No runtimes observed");
    expect(html).toContain("No approved or discovered inference runtimes are currently available for this host.");
    expect(html).toContain("endpoint locations are intentionally withheld");
    expect(html).not.toContain("endpoint=");
  });

  it("renders approval and health state without serializing configured endpoint locations", () => {
    const endpointCanary = "https://runtime-location-canary.example.lan:8000/v1";
    const html = renderToStaticMarkup(createElement(RuntimeInventoryPanel, { runtimes: [{ kind: "vllm", endpoint: endpointCanary, state: "approved", health: "healthy" }] }));

    expect(html).toContain("vllm");
    expect(html).toContain("Approved route");
    expect(html).toContain("Healthy");
    expect(html).not.toContain(endpointCanary);
    expect(html).not.toContain("runtime-location-canary");
  });
});
