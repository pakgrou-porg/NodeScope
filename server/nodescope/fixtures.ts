export type Quality = "fresh" | "stale" | "unavailable" | "unsupported" | "estimated";

export type Metric = {
  id: string;
  label: string;
  value: number | null;
  display: string;
  unit: string;
  quality: Quality;
  source: string;
  semantics: string;
  observedAt: Date;
  trend?: "up" | "down" | "steady";
};

export type HostSnapshot = {
  id: string;
  name: string;
  hostname: string;
  platform: string;
  architecture: string;
  role: "preferred" | "secondary";
  freshness: { state: "fresh" | "stale" | "unavailable"; ageSeconds: number; observedAt: Date };
  status: "healthy" | "degraded" | "unavailable";
  tags: string[];
  quickMetrics: Metric[];
  hardware: Metric[];
  memory: Metric[];
  storage: Array<{
    id: string;
    mount: string;
    usage: number | null;
    used: string;
    total: string;
    state: "mounted" | "read-only" | "missing" | "learning";
    quality: Quality;
    source: string;
  }>;
  processes: Array<{ name: string; status: "healthy" | "degraded" | "unavailable"; pid: string; uptime: string; selected: boolean }>;
  containers: Array<{ name: string; image: string; state: "running" | "exited" | "restarting"; selected: boolean; age: string }>;
  runtimes: Array<{ kind: "vllm" | "llama.cpp" | "lmstudio" | "agentzero" | "other"; endpoint: string; state: "approved" | "discovered" | "unavailable"; health: "healthy" | "degraded" | "unavailable" }>;
  inference: {
    requestRate: Metric;
    ttft: Metric;
    promptThroughput: Metric;
    generationThroughput: Metric;
    activeRequests: Metric;
    clientUsage: Array<{ client: string; requests: number; promptTokens: number; outputTokens: number; ttft: string }>;
  };
  preflight: Array<{ capability: string; state: "available" | "missing" | "degraded"; detail: string; installHint?: string }>;
  history: Array<{ timestamp: string; cpu: number; memory: number; temperature: number; throughput: number }>;
};

export type FleetSnapshot = {
  generatedAt: Date;
  developmentPreview: boolean;
  activeAlertCount: number;
  hosts: HostSnapshot[];
  alerts: Array<{
    id: string;
    severity: "critical" | "warning" | "info";
    state: "active" | "acknowledged" | "resolved";
    hostId: string;
    title: string;
    detail: string;
    observedAt: Date;
  }>;
  replicas: Array<{
    id: string;
    hostId: string;
    role: "preferred" | "secondary";
    status: "healthy" | "degraded" | "unavailable";
    version: string;
    certificateDaysRemaining: number;
    backupFreshness: "fresh" | "stale" | "unavailable";
    sharedBackupMount: "mounted" | "missing" | "read-only";
    observedAt: Date;
  }>;
  globalIntervalSeconds: number;
  hostOverrides: Record<string, number>;
};

const metric = (
  id: string,
  label: string,
  value: number | null,
  display: string,
  unit: string,
  quality: Quality,
  source: string,
  semantics: string,
  observedAt: Date,
  trend?: Metric["trend"],
): Metric => ({ id, label, value, display, unit, quality, source, semantics, observedAt, trend });

function makeHistory(seed: number, now: Date) {
  return Array.from({ length: 24 }, (_, index) => {
    const phase = index / 3 + seed;
    return {
      timestamp: new Date(now.getTime() - (23 - index) * 5 * 60_000).toISOString(),
      cpu: Math.round(34 + Math.sin(phase) * 13 + (index % 4) * 2),
      memory: Math.round(48 + Math.cos(phase * 0.75) * 8),
      temperature: Math.round(51 + Math.sin(phase * 0.5) * 5),
      throughput: Math.round(Math.max(0, 21 + Math.cos(phase * 0.8) * 10)),
    };
  });
}

export function buildFleetSnapshot(): FleetSnapshot {
  const now = new Date();
  const frameworkObserved = new Date(now.getTime() - 3_000);
  const asusObserved = new Date(now.getTime() - 7_000);

  const framework: HostSnapshot = {
    id: "framework",
    name: "Framework",
    hostname: "framework",
    platform: "Fedora 43 · Ryzen AI Max+ 395",
    architecture: "linux/amd64",
    role: "preferred",
    freshness: { state: "fresh", ageSeconds: 3, observedAt: frameworkObserved },
    status: "healthy",
    tags: ["Primary replica", "AMD GPU", "XDNA NPU", "Docker"],
    quickMetrics: [
      metric("cpu", "CPU", 47, "47%", "percent", "fresh", "procfs", "aggregate host CPU utilization", frameworkObserved, "up"),
      metric("ram", "RAM", 52, "51.6 GB / 96 GB", "bytes", "fresh", "/proc/meminfo", "host OS memory utilization", frameworkObserved, "steady"),
      metric("gpu", "Radeon 8060S", 61, "61%", "percent", "fresh", "AMD SMI", "GPU utilization", frameworkObserved, "up"),
      metric("npu", "XDNA NPU", 1, "Ready", "state", "fresh", "xrt-smi", "NPU readiness", frameworkObserved, "steady"),
      metric("temp", "Package temp", 68, "68°C", "celsius", "fresh", "lm-sensors", "CPU package temperature", frameworkObserved, "up"),
      metric("storage", "Primary storage", 42, "1.62 TB free", "bytes", "fresh", "statfs", "root filesystem free capacity", frameworkObserved, "steady"),
    ],
    hardware: [
      metric("cpu-util", "CPU utilization", 47, "47%", "percent", "fresh", "procfs", "aggregate host CPU utilization", frameworkObserved),
      metric("cpu-temp", "CPU package temperature", 68, "68°C", "celsius", "fresh", "lm-sensors", "physical package sensor", frameworkObserved),
      metric("gpu-util", "Radeon GPU utilization", 61, "61%", "percent", "fresh", "AMD SMI", "GPU engine utilization", frameworkObserved),
      metric("gpu-temp", "Radeon edge temperature", 64, "64°C", "celsius", "fresh", "AMD SMI", "GPU edge sensor", frameworkObserved),
      metric("npu-ready", "XDNA readiness", 1, "Ready", "state", "fresh", "xrt-smi", "NPU runtime readiness", frameworkObserved),
      metric("npu-throughput", "NPU throughput", null, "Unavailable", "ops/s", "unavailable", "xrt-smi", "not exposed by current runtime", frameworkObserved),
    ],
    memory: [
      metric("ram-avail", "MemAvailable", 46.4, "46.4 GB", "gigabytes", "fresh", "/proc/meminfo", "OS-reclaimable host memory", frameworkObserved),
      metric("swap-free", "SwapFree", 15.9, "15.9 GB", "gigabytes", "fresh", "/proc/meminfo", "free swap capacity", frameworkObserved),
      metric("gpu-memory", "GPU memory used", 18.2, "18.2 GB", "gigabytes", "fresh", "AMD SMI", "GPU-reported memory usage", frameworkObserved),
    ],
    storage: [
      { id: "root", mount: "/", usage: 58, used: "2.23 TB", total: "3.85 TB", state: "mounted", quality: "fresh", source: "statfs" },
      { id: "models", mount: "/mnt/models", usage: 71, used: "5.67 TB", total: "8 TB", state: "mounted", quality: "fresh", source: "statfs" },
      { id: "docker", mount: "docker:node_scope_data", usage: null, used: "Learning", total: "—", state: "learning", quality: "estimated", source: "Docker API" },
    ],
    processes: [
      { name: "AgentZero", status: "healthy", pid: "8412", uptime: "3d 12h", selected: true },
      { name: "kodex-ocr", status: "healthy", pid: "9107", uptime: "18h 32m", selected: true },
      { name: "nodescope-agent", status: "healthy", pid: "1219", uptime: "5d 4h", selected: true },
    ],
    containers: [
      { name: "supabase-db", image: "supabase/postgres", state: "running", selected: true, age: "12d" },
      { name: "agent-zero", image: "agent0ai/agent-zero:2.5", state: "running", selected: true, age: "3d" },
      { name: "portainer", image: "portainer/portainer-ce", state: "running", selected: false, age: "42d" },
      { name: "kodex", image: "local/kodex-ocr", state: "running", selected: true, age: "18h" },
    ],
    runtimes: [
      { kind: "vllm", endpoint: "http://framework:8000/v1", state: "approved", health: "healthy" },
      { kind: "llama.cpp", endpoint: "http://framework:8080", state: "discovered", health: "healthy" },
      { kind: "agentzero", endpoint: "http://framework:50001", state: "approved", health: "healthy" },
    ],
    inference: {
      requestRate: metric("requests", "Requests", 4.8, "4.8 req/min", "req/min", "fresh", "NodeScope proxy", "accepted proxy requests", frameworkObserved, "up"),
      ttft: metric("ttft", "P95 TTFT", 482, "482 ms", "milliseconds", "fresh", "NodeScope proxy", "end-to-end first token latency", frameworkObserved, "down"),
      promptThroughput: metric("prompt-rate", "Prompt processing", 1180, "1,180 tok/s", "tok/s", "fresh", "vLLM metrics", "engine prompt token throughput", frameworkObserved, "up"),
      generationThroughput: metric("generation-rate", "Generation", 86, "86 tok/s", "tok/s", "fresh", "NodeScope proxy", "output token throughput", frameworkObserved, "steady"),
      activeRequests: metric("active", "Active requests", 2, "2", "requests", "fresh", "vLLM metrics", "active engine requests", frameworkObserved),
      clientUsage: [
        { client: "AgentZero", requests: 38, promptTokens: 148_220, outputTokens: 18_433, ttft: "468 ms" },
        { client: "Kodex", requests: 12, promptTokens: 42_098, outputTokens: 5_307, ttft: "512 ms" },
      ],
    },
    preflight: [
      { capability: "AMD SMI", state: "available", detail: "GPU utilization, temperature, process and memory collectors enabled." },
      { capability: "XDNA runtime", state: "available", detail: "xrt-smi readiness and active-context collector enabled." },
      { capability: "Docker socket", state: "available", detail: "Container inventory and selected-container health enabled." },
      { capability: "vLLM metrics", state: "available", detail: "Native metrics endpoint discovered at approved runtime." },
    ],
    history: makeHistory(1, now),
  };

  const asus: HostSnapshot = {
    id: "asus",
    name: "Asus",
    hostname: "asus",
    platform: "DGX OS · ASUS Ascent GX10",
    architecture: "linux/arm64",
    role: "secondary",
    freshness: { state: "fresh", ageSeconds: 7, observedAt: asusObserved },
    status: "degraded",
    tags: ["Secondary replica", "GB10 UMA", "ConnectX-7", "Docker"],
    quickMetrics: [
      metric("cpu", "CPU", 33, "33%", "percent", "fresh", "procfs", "aggregate host CPU utilization", asusObserved, "down"),
      metric("uma", "Unified memory", 54, "69.2 GB / 128 GB", "bytes", "fresh", "/proc/meminfo", "OS memory utilization under UMA", asusObserved, "steady"),
      metric("gpu", "Blackwell GPU", 72, "72%", "percent", "fresh", "nvidia-smi", "GPU engine utilization", asusObserved, "up"),
      metric("npu", "NPU", null, "Unsupported", "state", "unsupported", "platform inventory", "No NPU collector exposed by this platform", asusObserved),
      metric("temp", "SoC temperature", 71, "71°C", "celsius", "fresh", "nvidia-smi", "platform thermal sensor", asusObserved, "up"),
      metric("storage", "Primary storage", 36, "4.1 TB free", "bytes", "fresh", "statfs", "root filesystem free capacity", asusObserved, "steady"),
    ],
    hardware: [
      metric("cpu-util", "CPU utilization", 33, "33%", "percent", "fresh", "procfs", "aggregate Arm CPU utilization", asusObserved),
      metric("soc-temp", "SoC temperature", 71, "71°C", "celsius", "fresh", "nvidia-smi", "platform thermal sensor", asusObserved),
      metric("gpu-util", "Blackwell GPU utilization", 72, "72%", "percent", "fresh", "nvidia-smi", "GPU engine utilization", asusObserved),
      metric("connectx", "ConnectX-7", 1, "Detected", "state", "fresh", "lspci", "inventory-only device detection", asusObserved),
      metric("dcgm", "DCGM metrics", null, "Unsupported", "state", "unsupported", "capability probe", "No hard DCGM dependency on GX10", asusObserved),
    ],
    memory: [
      metric("uma-os", "OS MemAvailable", 52.7, "52.7 GB", "gigabytes", "fresh", "/proc/meminfo", "OS-reclaimable host memory under unified memory", asusObserved),
      metric("uma-swap", "SwapFree", 31.8, "31.8 GB", "gigabytes", "fresh", "/proc/meminfo", "free swap capacity under unified memory", asusObserved),
      metric("uma-hugepages", "Huge-page state", 6, "6.0 GB reserved", "gigabytes", "fresh", "/proc/meminfo", "huge-page reservation", asusObserved),
      metric("uma-runtime", "CUDA allocatable", 46.2, "46.2 GB", "gigabytes", "estimated", "cudaMemGetInfo", "runtime-visible allocation estimate; not total reclaimable UMA", asusObserved),
      metric("uma-process", "vLLM process GPU memory", 31.4, "31.4 GB", "gigabytes", "fresh", "nvidia-smi process query", "per-process GPU memory where exposed", asusObserved),
      metric("vram", "Dedicated VRAM free", null, "Unavailable", "gigabytes", "unsupported", "nvidia-smi", "GX10 has unified memory; this field is not synthesized", asusObserved),
    ],
    storage: [
      { id: "root", mount: "/", usage: 64, used: "1.28 TB", total: "2 TB", state: "mounted", quality: "fresh", source: "statfs" },
      { id: "models", mount: "/mnt/models", usage: 49, used: "3.92 TB", total: "8 TB", state: "mounted", quality: "fresh", source: "statfs" },
      { id: "backup", mount: "/mnt/nodescope-backups", usage: null, used: "Unavailable", total: "—", state: "missing", quality: "unavailable", source: "mountinfo" },
    ],
    processes: [
      { name: "nodescope-agent", status: "healthy", pid: "987", uptime: "5d 4h", selected: true },
      { name: "nvidia-persistenced", status: "healthy", pid: "612", uptime: "5d 4h", selected: false },
      { name: "vllm", status: "degraded", pid: "7251", uptime: "1d 7h", selected: true },
    ],
    containers: [
      { name: "vllm-qwen", image: "vllm/vllm-openai", state: "running", selected: true, age: "1d" },
      { name: "portainer", image: "portainer/portainer-ce", state: "running", selected: false, age: "42d" },
      { name: "nodescope-server", image: "ghcr.io/pakgrou-porg/nodescope", state: "running", selected: true, age: "5d" },
    ],
    runtimes: [
      { kind: "vllm", endpoint: "http://asus:8000/v1", state: "approved", health: "degraded" },
      { kind: "llama.cpp", endpoint: "http://asus:8080", state: "discovered", health: "unavailable" },
    ],
    inference: {
      requestRate: metric("requests", "Requests", 3.1, "3.1 req/min", "req/min", "fresh", "NodeScope proxy", "accepted proxy requests", asusObserved, "steady"),
      ttft: metric("ttft", "P95 TTFT", 861, "861 ms", "milliseconds", "fresh", "NodeScope proxy", "end-to-end first token latency", asusObserved, "up"),
      promptThroughput: metric("prompt-rate", "Prompt processing", 892, "892 tok/s", "tok/s", "fresh", "vLLM metrics", "engine prompt token throughput", asusObserved, "down"),
      generationThroughput: metric("generation-rate", "Generation", 64, "64 tok/s", "tok/s", "fresh", "NodeScope proxy", "output token throughput", asusObserved, "down"),
      activeRequests: metric("active", "Active requests", 1, "1", "requests", "fresh", "vLLM metrics", "active engine requests", asusObserved),
      clientUsage: [
        { client: "AgentZero", requests: 17, promptTokens: 63_009, outputTokens: 8_084, ttft: "844 ms" },
        { client: "CLI benchmark", requests: 5, promptTokens: 9_604, outputTokens: 2_109, ttft: "921 ms" },
      ],
    },
    preflight: [
      { capability: "nvidia-smi", state: "available", detail: "GPU utilization and process-memory collector enabled; dedicated framebuffer fields remain capability-gated." },
      { capability: "DGX UMA semantics", state: "available", detail: "OS, runtime, huge-page, and process views are kept distinct." },
      { capability: "ConnectX-7 inventory", state: "available", detail: "PCI and link-state inventory enabled; RDMA performance is deferred." },
      { capability: "Shared backup mount", state: "missing", detail: "Secondary backup takeover cannot run until the shared target is mounted.", installHint: "Mount the designated backup target at /mnt/nodescope-backups." },
    ],
    history: makeHistory(2.5, now),
  };

  return {
    generatedAt: now,
    developmentPreview: true,
    activeAlertCount: 2,
    hosts: [framework, asus],
    alerts: [
      { id: "alert-backup-mount", severity: "warning", state: "active", hostId: "asus", title: "Shared backup mount unavailable", detail: "Asus cannot participate in backup takeover until /mnt/nodescope-backups is mounted and writable.", observedAt: asusObserved },
      { id: "alert-vllm-ttft", severity: "warning", state: "active", hostId: "asus", title: "vLLM P95 TTFT above preferred threshold", detail: "P95 first-token latency is 861 ms for the approved vLLM route.", observedAt: asusObserved },
    ],
    replicas: [
      { id: "replica-framework", hostId: "framework", role: "preferred", status: "healthy", version: "0.1.0-dev", certificateDaysRemaining: 86, backupFreshness: "fresh", sharedBackupMount: "mounted", observedAt: frameworkObserved },
      { id: "replica-asus", hostId: "asus", role: "secondary", status: "degraded", version: "0.1.0-dev", certificateDaysRemaining: 86, backupFreshness: "stale", sharedBackupMount: "missing", observedAt: asusObserved },
    ],
    globalIntervalSeconds: 5,
    hostOverrides: { asus: 10 },
  };
}
