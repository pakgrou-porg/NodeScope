import { describe, expect, it } from "vitest";
import { runtimeInventoryDisplay, runtimeInventoryEmptyState } from "./runtimeInventory";

describe("runtimeInventoryDisplay", () => {
  it("renders approval and health evidence without exposing the configured endpoint", () => {
    const display = runtimeInventoryDisplay({
      kind: "vllm",
      endpoint: "https://runtime-secret-location.example.lan:8000/v1?should-not-render=true",
      state: "approved",
      health: "healthy",
    });

    expect(display).toEqual({
      kind: "vllm",
      approvalLabel: "Approved route",
      healthLabel: "Healthy",
      stateTone: "approved",
      healthTone: "healthy",
    });
    expect(JSON.stringify(display)).not.toContain("runtime-secret-location");
  });

  it("provides a clear empty state without any endpoint field", () => {
    expect(runtimeInventoryEmptyState).toEqual({
      title: "No runtimes observed",
      detail: "No approved or discovered inference runtimes are currently available for this host.",
    });
    expect(JSON.stringify(runtimeInventoryEmptyState)).not.toContain("endpoint");
  });
});
