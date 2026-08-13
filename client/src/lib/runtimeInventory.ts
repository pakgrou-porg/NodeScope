export type RuntimeInventoryEntry = {
  kind: "vllm" | "llama.cpp" | "lmstudio" | "agentzero" | "other";
  endpoint: string;
  state: "approved" | "discovered" | "unavailable";
  health: "healthy" | "degraded" | "unavailable";
};

export type RuntimeInventoryDisplay = {
  kind: string;
  approvalLabel: string;
  healthLabel: string;
  stateTone: "approved" | "discovered" | "unavailable";
  healthTone: "healthy" | "degraded" | "unavailable";
};

export const runtimeInventoryEmptyState = {
  title: "No runtimes observed",
  detail: "No approved or discovered inference runtimes are currently available for this host.",
};

// Runtime endpoints may be configured locally but are intentionally not part of
// browser presentation data. The console needs approval and health evidence,
// not a network location or an inference-content surface.
export function runtimeInventoryDisplay(entry: RuntimeInventoryEntry): RuntimeInventoryDisplay {
  return {
    kind: entry.kind === "lmstudio" ? "LM Studio" : entry.kind,
    approvalLabel: entry.state === "approved" ? "Approved route" : entry.state === "discovered" ? "Discovered · approval required" : "Runtime unavailable",
    healthLabel: entry.health === "healthy" ? "Healthy" : entry.health === "degraded" ? "Degraded" : "Unavailable",
    stateTone: entry.state,
    healthTone: entry.health,
  };
}
