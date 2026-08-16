import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("host directory filters", () => {
  it("keeps host search, availability filters, result count, empty-state recovery, and detail navigation explicit", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/HostsPage.tsx"), "utf8");

    expect(source).toContain('aria-label="Search hosts"');
    expect(source.indexOf("const visibleHosts = useMemo")).toBeLessThan(source.indexOf("if (query.isLoading || !query.data)"));
    expect(source).toContain('aria-label="Filter hosts by availability"');
    expect(source).toContain('aria-pressed={availability === filter}');
    expect(source).toContain('Showing {visibleHosts.length} of {query.data.hosts.length} configured hosts.');
    expect(source).toContain("No hosts match the current directory filters.");
    expect(source).toContain("Clear directory filters");
    expect(source).toContain('targetPath(`/hosts/${host.id}`)');
  });
});
