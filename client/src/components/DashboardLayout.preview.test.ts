import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("dashboard development preview routing", () => {
  it("accepts the exact preview prefix while retaining the development-only boundary", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/components/DashboardLayout.tsx"), "utf8");

    expect(source).toContain('import.meta.env.DEV && /^\\/preview(?:\\/|$)/.test(window.location.pathname)');
    expect(source).not.toContain('window.location.pathname === "/preview"');
  });
});
