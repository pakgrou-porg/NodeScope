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

	it("rejects a Viewer-equivalent user approving a runtime", async () => {
		const caller = appRouter.createCaller(createContext("user"));

		await expect(caller.nodescope.runtimes.approve({ hostId: "framework", endpoint: "https://runtime.example.lan/v1", runtimeKind: "vllm" })).rejects.toMatchObject({
			code: "FORBIDDEN",
		});
	});

	it("records only an opaque candidate ID and safe transport metadata for approved runtimes", async () => {
		const caller = appRouter.createCaller(createContext("admin"));
		const endpointCanary = "https://runtime-location-canary.example.lan:8443/v1";
		const result = await caller.nodescope.runtimes.approve({ hostId: "framework", endpoint: endpointCanary, runtimeKind: "vllm" });

		expect(result.state).toBe("approved");
		expect(result.candidateId).toMatch(/^runtime:framework:/);
		expect(result.candidateId).not.toContain("runtime-location-canary");
		expect(result.auditEvent.metadata).toEqual({ runtimeKind: "vllm", transport: "https" });
		expect(JSON.stringify(result.auditEvent)).not.toContain("runtime-location-canary");
	});

	it("rejects unsafe runtime endpoint forms before creating an approval record", async () => {
		const caller = appRouter.createCaller(createContext("admin"));
		for (const endpoint of [
			"https://client:credential@example.lan/v1",
			"https://runtime.example.lan/v1?token=canary",
			"https://runtime.example.lan/other-path",
			"http://runtime.example.lan:8000/v1",
		]) {
			await expect(caller.nodescope.runtimes.approve({ hostId: "framework", endpoint, runtimeKind: "vllm" })).rejects.toMatchObject({ code: "BAD_REQUEST" });
		}
	});

	it("permits an explicitly local HTTP runtime endpoint without storing its location", async () => {
		const caller = appRouter.createCaller(createContext("admin"));
		const result = await caller.nodescope.runtimes.approve({ hostId: "framework", endpoint: "http://127.0.0.1:8000/v1", runtimeKind: "llama.cpp" });

		expect(result.auditEvent.metadata).toEqual({ runtimeKind: "llama.cpp", transport: "loopback_http" });
		expect(JSON.stringify(result)).not.toContain("127.0.0.1:8000");
	});
});
