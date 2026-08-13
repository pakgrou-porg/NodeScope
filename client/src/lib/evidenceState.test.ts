import { describe, expect, it } from "vitest";
import { metricQualities, metricQualityAriaLabel, metricQualityLabel } from "./evidenceState";

describe("metric evidence state labels", () => {
  it("renders every supported quality as explicit evidence", () => {
    expect(metricQualities.map(metricQualityLabel)).toEqual([
      "fresh evidence",
      "stale evidence",
      "unavailable evidence",
      "not supported",
      "estimated evidence",
      "experimental evidence",
    ]);
  });

  it("does not turn unavailable evidence into a numeric substitute", () => {
    expect(metricQualityLabel("unavailable")).toBe("unavailable evidence");
    expect(metricQualityAriaLabel("unavailable")).toBe("Metric quality: unavailable evidence");
  });
});
