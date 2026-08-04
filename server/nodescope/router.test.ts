import { describe, expect, it } from "vitest";
import { appRouter } from "../routers";
import type { TrpcContext } from "../_core/context";

type TemplateRole = "admin" | "user";

function createContext(role: TemplateRole): TrpcContext {
  return {
    user: {
      id: role === "admin" ? 1 : 2,
      openId: `${role}-fixture`,
      name: role,
      email: `${role}@example.test`,
      loginMethod: "fixture",
      role,
      createdAt: new Date(),
      updatedAt: new Date(),
      lastSignedIn: new Date(),
    },
    req: { protocol: "https", headers: {} } as TrpcContext["req"],
    res: { clearCookie: () => undefined } as TrpcContext["res"],
  };
}

describe("NodeScope tRPC authorization", () => {
  it("allows a Viewer-equivalent user to read fleet state", async () => {
    const caller = appRouter.createCaller(createContext("user"));
    const overview = await caller.nodescope.fleet.overview();

    expect(overview.hosts).toHaveLength(2);
    expect(overview.hosts.map(host => host.id)).toEqual(["framework", "asus"]);
  });

  it("rejects a Viewer-equivalent user changing collection intervals", async () => {
    const caller = appRouter.createCaller(createContext("user"));

    await expect(caller.nodescope.configuration.setCollectionInterval({ intervalSeconds: 10 })).rejects.toMatchObject({
      code: "FORBIDDEN",
    });
  });

  it("allows an Administrator-equivalent user to change collection intervals and creates audit output", async () => {
    const caller = appRouter.createCaller(createContext("admin"));
    const result = await caller.nodescope.configuration.setCollectionInterval({ hostId: "framework", intervalSeconds: 8 });

    expect(result.state).toBe("completed");
    expect(result.effectiveIntervalSeconds).toBe(8);
    expect(result.auditEvent.action).toBe("collection_interval.set");
  });

  it("rejects an unknown host before recording a configuration change", async () => {
    const caller = appRouter.createCaller(createContext("admin"));

    await expect(caller.nodescope.configuration.setCollectionInterval({ hostId: "missing-host", intervalSeconds: 8 })).rejects.toMatchObject({
      code: "NOT_FOUND",
    });
  });
});
