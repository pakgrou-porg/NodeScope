import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("administration page clarity", () => {
  it("keeps control boundaries, summary-backup scope, audit state, preview safety, and server-side roles explicit", () => {
    const source = readFileSync(resolve(process.cwd(), "client/src/pages/AdministrationPage.tsx"), "utf8");

    expect(source).toContain('aria-label="Administration control boundaries"');
    expect(source).toContain("This map distinguishes configuration views, role-protected actions, and controls that remain deferred");
    expect(source).toContain("Summary backup scope: configuration and summary telemetry only.");
    expect(source).toContain("Full raw telemetry is not included by this control.");
    expect(source).toContain("auditStatus");
    expect(source).toContain('disabled={preview || backup.isPending}');
    expect(source).toContain('backup.mutate({ includeRawTelemetry: false })');
    expect(source).toContain("enforced on the server—never only hidden in the user interface");
  });
});
