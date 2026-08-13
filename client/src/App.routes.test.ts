import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("browser console route loading", () => {
  it("keeps dashboard views behind route-level lazy imports with an accessible loading fallback", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/App.tsx"), "utf8");

    for (const page of ["FleetOverview", "HostDetail", "AlertsPage", "OperationsPage", "AdministrationPage", "AlertRulesPage"]) {
      expect(source).toContain(`const ${page} = lazy(() => import("./pages/${page}"))`);
    }
    expect(source).toContain("<Suspense fallback={<ConsoleRouteLoading />}>");
    expect(source).toContain('aria-busy="true"');
    expect(source).toContain("Loading NodeScope console…");
  });
});
