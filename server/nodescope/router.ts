import { TRPCError } from "@trpc/server";
import { z } from "zod";
import { administratorProcedure, operatorProcedure, publicProcedure, router, viewerProcedure } from "../_core/trpc";
import { buildFleetSnapshot } from "./fixtures";

type AlertRule = {
  id: string;
  metric: string;
  operator: "gt" | "lt";
  threshold: number;
  durationSeconds: number;
  severity: "critical" | "warning" | "info";
  enabled: boolean;
  scope: "fleet" | "host";
  hostId: string | null;
};

type DevelopmentAudit = {
  id: string;
  actorId: string;
  actorType: "user";
  action: string;
  targetType: string;
  targetId: string | null;
  outcome: "intent" | "acknowledged" | "completed" | "failed" | "denied";
  occurredAt: Date;
  metadata: Record<string, unknown>;
};

const developmentState = {
  globalIntervalSeconds: 5,
  hostOverrides: new Map<string, number>([["asus", 10]]),
  acknowledgedAlerts: new Set<string>(),
  approvedRuntimeIds: new Set<string>(),
  audits: [] as DevelopmentAudit[],
  alertRules: [
    { id: "rule-asus-ttft", metric: "inference.ttft_p95_ms", operator: "gt", threshold: 750, durationSeconds: 300, severity: "warning", enabled: true, scope: "host", hostId: "asus" },
    { id: "rule-storage", metric: "storage.usage_percent", operator: "gt", threshold: 85, durationSeconds: 600, severity: "warning", enabled: true, scope: "fleet", hostId: null },
    { id: "rule-temperature", metric: "device.temperature_celsius", operator: "gt", threshold: 82, durationSeconds: 120, severity: "critical", enabled: true, scope: "fleet", hostId: null },
  ] as AlertRule[],
};

function createAudit(
  actorId: string,
  action: string,
  targetType: string,
  targetId: string | null,
  metadata: Record<string, unknown>,
  outcome: DevelopmentAudit["outcome"] = "intent",
): DevelopmentAudit {
  const audit: DevelopmentAudit = {
    id: crypto.randomUUID(),
    actorId,
    actorType: "user",
    action,
    targetType,
    targetId,
    outcome,
    occurredAt: new Date(),
    metadata,
  };
  developmentState.audits.unshift(audit);
  return audit;
}

function requireHost(hostId: string) {
  const host = buildFleetSnapshot().hosts.find(candidate => candidate.id === hostId);
  if (!host) {
    throw new TRPCError({ code: "NOT_FOUND", message: `Unknown host ${hostId}` });
  }
  return host;
}

function validateRuntimeApprovalEndpoint(rawEndpoint: string): { transport: "https" | "loopback_http" } {
	let endpoint: URL;
	try {
		endpoint = new URL(rawEndpoint);
	} catch {
		throw new TRPCError({ code: "BAD_REQUEST", message: "Runtime endpoint must be a valid URL" });
	}
	const path = endpoint.pathname.replace(/\/$/, "");
	if (endpoint.username || endpoint.password || endpoint.search || endpoint.hash || !endpoint.hostname || !["http:", "https:"].includes(endpoint.protocol) || !["", "/v1"].includes(path)) {
		throw new TRPCError({ code: "BAD_REQUEST", message: "Runtime endpoint must be a credential-free HTTP(S) base URL or /v1 endpoint" });
	}
	if (endpoint.protocol === "http:") {
		const hostname = endpoint.hostname.toLowerCase();
		if (hostname !== "localhost" && hostname !== "127.0.0.1" && hostname !== "::1") {
			throw new TRPCError({ code: "BAD_REQUEST", message: "Runtime endpoint must use HTTPS unless it is loopback" });
		}
		return { transport: "loopback_http" };
	}
	return { transport: "https" };
}

export const nodeScopeRouter = router({
  fleet: router({
    preview: publicProcedure.query(() => {
      if (process.env.NODE_ENV === "production") {
        throw new TRPCError({ code: "FORBIDDEN", message: "Development preview is unavailable in production" });
      }
      return buildFleetSnapshot();
    }),
    overview: viewerProcedure.query(() => {
      const snapshot = buildFleetSnapshot();
      return {
        ...snapshot,
        globalIntervalSeconds: developmentState.globalIntervalSeconds,
        hostOverrides: Object.fromEntries(developmentState.hostOverrides),
        alerts: snapshot.alerts.map(alert => ({
          ...alert,
          state: developmentState.acknowledgedAlerts.has(alert.id) ? "acknowledged" as const : alert.state,
        })),
      };
    }),
    host: viewerProcedure
      .input(z.object({ hostId: z.string().min(1) }))
      .query(({ input }) => requireHost(input.hostId)),
  }),

  alerts: router({
    rules: router({
      preview: publicProcedure.query(() => {
        if (process.env.NODE_ENV === "production") {
          throw new TRPCError({ code: "FORBIDDEN", message: "Development preview is unavailable in production" });
        }
        return developmentState.alertRules;
      }),
      list: viewerProcedure.query(() => developmentState.alertRules),
      save: operatorProcedure
        .input(z.object({
          id: z.string().min(1).max(100),
          metric: z.string().min(1).max(120),
          operator: z.enum(["gt", "lt"]),
          threshold: z.number().finite(),
          durationSeconds: z.number().int().min(1).max(86_400),
          severity: z.enum(["critical", "warning", "info"]),
          enabled: z.boolean(),
          scope: z.enum(["fleet", "host"]),
          hostId: z.string().min(1).nullable(),
        }))
        .mutation(({ ctx, input }) => {
          if (input.scope === "host" && !input.hostId) {
            throw new TRPCError({ code: "BAD_REQUEST", message: "A host-scoped alert rule requires hostId" });
          }
          if (input.scope === "fleet" && input.hostId) {
            throw new TRPCError({ code: "BAD_REQUEST", message: "A fleet-scoped alert rule cannot target a host" });
          }
          if (input.scope === "host" && input.hostId) requireHost(input.hostId);
          const index = developmentState.alertRules.findIndex((rule) => rule.id === input.id);
          const rule: AlertRule = { ...input };
          if (index >= 0) developmentState.alertRules[index] = rule;
          else developmentState.alertRules.unshift(rule);
          const audit = createAudit(ctx.user.openId, "alert_rule.save", "alert_rule", rule.id, { metric: rule.metric, severity: rule.severity, scope: rule.scope }, "completed");
          return { rule, auditEvent: audit };
        }),
    }),
    acknowledge: operatorProcedure
      .input(z.object({ alertId: z.string().min(1), note: z.string().trim().max(500).optional() }))
      .mutation(({ ctx, input }) => {
        const snapshot = buildFleetSnapshot();
        const alert = snapshot.alerts.find(candidate => candidate.id === input.alertId);
        if (!alert) {
          throw new TRPCError({ code: "NOT_FOUND", message: `Unknown alert ${input.alertId}` });
        }
        developmentState.acknowledgedAlerts.add(alert.id);
        const audit = createAudit(ctx.user.openId, "alert.acknowledge", "alert", alert.id, {
          note: input.note ?? null,
          hostId: alert.hostId,
        }, "completed");
        return { alertId: alert.id, state: "acknowledged" as const, auditEvent: audit };
      }),
  }),

  configuration: router({
    setCollectionInterval: operatorProcedure
      .input(z.object({ hostId: z.string().min(1).optional(), intervalSeconds: z.number().int().min(1).max(60) }))
      .mutation(({ ctx, input }) => {
        if (input.hostId) {
          requireHost(input.hostId);
          developmentState.hostOverrides.set(input.hostId, input.intervalSeconds);
        } else {
          developmentState.globalIntervalSeconds = input.intervalSeconds;
        }
        const audit = createAudit(ctx.user.openId, "collection_interval.set", input.hostId ? "host" : "fleet", input.hostId ?? null, {
          intervalSeconds: input.intervalSeconds,
        }, "completed");
        return {
          operationId: crypto.randomUUID(),
          state: "completed" as const,
          auditEvent: audit,
          effectiveIntervalSeconds: input.intervalSeconds,
        };
      }),
    refreshStorageBaseline: operatorProcedure
      .input(z.object({ hostId: z.string().min(1), acknowledgedDiff: z.boolean() }))
      .mutation(({ ctx, input }) => {
        const host = requireHost(input.hostId);
        if (!input.acknowledgedDiff) {
          throw new TRPCError({ code: "BAD_REQUEST", message: "Storage baseline diff acknowledgement is required" });
        }
        const audit = createAudit(ctx.user.openId, "storage_baseline.refresh", "host", host.id, {
          resourceCount: host.storage.length,
        }, "completed");
        return { operationId: crypto.randomUUID(), state: "completed" as const, auditEvent: audit };
      }),
  }),

	runtimes: router({
		approve: administratorProcedure
			.input(z.object({ hostId: z.string().min(1), endpoint: z.string().trim().min(1).max(2048), runtimeKind: z.enum(["vllm", "llama.cpp", "lmstudio", "agentzero", "other"]) }))
			.mutation(({ ctx, input }) => {
				requireHost(input.hostId);
				const endpoint = validateRuntimeApprovalEndpoint(input.endpoint);
				const candidateId = `runtime:${input.hostId}:${crypto.randomUUID()}`;
				developmentState.approvedRuntimeIds.add(candidateId);
				const audit = createAudit(ctx.user.openId, "runtime.approve", "runtime_candidate", candidateId, {
					runtimeKind: input.runtimeKind,
					transport: endpoint.transport,
				}, "completed");
        return { candidateId, state: "approved" as const, auditEvent: audit };
      }),
  }),

  backups: router({
    initiate: administratorProcedure
      .input(z.object({ includeRawTelemetry: z.boolean().default(false) }))
      .mutation(({ ctx, input }) => {
        const audit = createAudit(ctx.user.openId, "backup.initiate", "backup", null, {
          includeRawTelemetry: input.includeRawTelemetry,
        }, "intent");
        return { operationId: crypto.randomUUID(), state: "pending" as const, auditEvent: audit };
      }),
  }),

  audit: router({
    list: viewerProcedure
      .input(z.object({ limit: z.number().int().min(1).max(100).default(25) }))
      .query(({ input }) => developmentState.audits.slice(0, input.limit)),
  }),
});
