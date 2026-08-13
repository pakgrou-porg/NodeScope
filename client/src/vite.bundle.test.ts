import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("browser bundle chunking", () => {
  it("keeps deterministic vendor chunk boundaries for console dependencies", () => {
    const source = readFileSync(resolve(process.cwd(), "vite.config.ts"), "utf8");

    expect(source).toContain("function browserVendorChunk");
    expect(source).toContain('return "vendor-react"');
    expect(source).toContain('return "vendor-charts"');
    expect(source).toContain('return "vendor-ui"');
    expect(source).toContain('return "vendor-data"');
    expect(source).toContain("manualChunks: browserVendorChunk");
  });
});
