import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("approved-backend streaming privacy procedure", () => {
  it("keeps authorization, metadata-only checks, fallback, and recovery explicit", () => {
    const procedure = readFileSync(resolve(process.cwd(), "docs/operations/inference-streaming-privacy-e2e.md"), "utf8");

    expect(procedure).toContain("not runnable without explicit authorization");
    expect(procedure).toContain("Metadata-only observability");
    expect(procedure).toContain("Fallback");
    expect(procedure).toContain("Redirect containment");
    expect(procedure).toContain("revoke the canary client credential");
    expect(procedure).toContain("rehearse-inference-privacy-local.sh");
  });
});
