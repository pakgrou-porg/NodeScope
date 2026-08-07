import { describe, expect, it } from "vitest";

const requiredEnvironmentKeys = [
  "NODESCOPE_SUPABASE_URL",
  "NODESCOPE_SUPABASE_ANON_KEY",
  "NODESCOPE_SUPABASE_SERVICE_ROLE_KEY",
  "NODESCOPE_SUPABASE_DB_URL",
  "NODESCOPE_RUNTIME_DB_PASSWORD",
  "NODESCOPE_MIGRATOR_DB_PASSWORD",
] as const;

const describeLiveSupabase = process.env.NODESCOPE_RUN_LIVE_SUPABASE_CHECK === "true" ? describe : describe.skip;

// This probe intentionally stays out of ordinary unit and CI runs. Run it only
// from a protected deployment environment that injects the NodeScope-specific
// secrets, with NODESCOPE_RUN_LIVE_SUPABASE_CHECK=true.
describeLiveSupabase("NodeScope live Supabase secret configuration", () => {
  it("has the required credentials and reaches the dedicated Auth health endpoint", async () => {
    for (const key of requiredEnvironmentKeys) {
      expect(process.env[key], `${key} must be configured`).toBeTruthy();
    }

    const databaseURL = process.env.NODESCOPE_SUPABASE_DB_URL!;
    expect(/^(postgresql|postgres):\/\//.test(databaseURL), "NODESCOPE_SUPABASE_DB_URL must be a PostgreSQL connection string").toBe(true);
    expect(process.env.NODESCOPE_RUNTIME_DB_PASSWORD!.length).toBeGreaterThanOrEqual(24);
    expect(process.env.NODESCOPE_MIGRATOR_DB_PASSWORD!.length).toBeGreaterThanOrEqual(24);
    expect(process.env.NODESCOPE_RUNTIME_DB_PASSWORD).not.toBe(process.env.NODESCOPE_MIGRATOR_DB_PASSWORD);

    const projectURL = new URL(process.env.NODESCOPE_SUPABASE_URL!);
    const healthURL = new URL("/auth/v1/health", projectURL);
    const response = await fetch(healthURL, {
      headers: {
        apikey: process.env.NODESCOPE_SUPABASE_ANON_KEY!,
      },
      signal: AbortSignal.timeout(10_000),
    });

    expect(response.ok, `Supabase Auth health returned HTTP ${response.status}`).toBe(true);
  });
});
