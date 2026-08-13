import DashboardLayout from "@/components/DashboardLayout";
import React from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { metricQualityAriaLabel, metricQualityLabel, type MetricQuality } from "@/lib/evidenceState";
import { runtimeInventoryDisplay, runtimeInventoryEmptyState, type RuntimeInventoryEntry } from "@/lib/runtimeInventory";
import { cn } from "@/lib/utils";
import { trpc } from "@/lib/trpc";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { Activity, ArrowLeft, Container, Cpu, Database, HardDrive, MemoryStick, Network, SearchCheck, ThermometerSun, Zap } from "lucide-react";
import { useLocation } from "wouter";

type Quality = MetricQuality;

const qualityText: Record<Quality, string> = {
  fresh: "text-emerald-200",
  stale: "text-amber-100",
  unavailable: "text-rose-200",
  unsupported: "text-slate-400",
  estimated: "text-violet-200",
  experimental: "text-fuchsia-200",
};

const qualityBadgeStyles: Record<Quality, string> = {
  fresh: "border-emerald-300/15 bg-emerald-400/10 text-emerald-200",
  stale: "border-amber-300/20 bg-amber-300/10 text-amber-100",
  unavailable: "border-rose-300/20 bg-rose-400/10 text-rose-100",
  unsupported: "border-slate-400/15 bg-slate-400/10 text-slate-300",
  estimated: "border-violet-300/20 bg-violet-400/10 text-violet-100",
  experimental: "border-fuchsia-300/20 bg-fuchsia-400/10 text-fuchsia-100",
};

function EvidenceBadge({ quality }: { quality: Quality }) {
  return <span aria-label={metricQualityAriaLabel(quality)} className={cn("inline-flex items-center rounded-full border px-2 py-0.5 text-[9px] font-medium tracking-[0.08em] uppercase", qualityBadgeStyles[quality])}>{metricQualityLabel(quality)}</span>;
}

function MetricRow({ metric }: { metric: { label: string; display: string; quality: Quality; source: string; semantics: string } }) {
  return <div className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-6 gap-y-2 border-b border-white/6 py-3 last:border-0"><div><p className="text-xs text-slate-300">{metric.label}</p><p className="mt-1 text-[10px] leading-4 text-slate-600">{metric.semantics}</p></div><div className="flex min-w-[156px] flex-col items-end text-right"><EvidenceBadge quality={metric.quality} /><p className={cn("mt-2 text-sm font-medium", qualityText[metric.quality])}>{metric.display}</p><p className="mt-1 text-[10px] text-slate-600">Source: {metric.source}</p></div></div>;
}

function SectionHeader({ icon: Icon, title, note }: { icon: typeof Cpu; title: string; note: string }) {
  return <div className="mb-4 flex items-start gap-3"><span className="grid h-9 w-9 place-items-center rounded-xl bg-cyan-300/10 text-cyan-200"><Icon className="h-4 w-4" /></span><div><h2 className="text-sm font-semibold text-slate-100">{title}</h2><p className="mt-1 text-xs text-slate-500">{note}</p></div></div>;
}

export default function HostDetail({ hostId, preview = false }: { hostId: string; preview?: boolean }) {
  const [, navigate] = useLocation();
  const previewQuery = trpc.nodescope.fleet.preview.useQuery(undefined, { enabled: preview, refetchInterval: 5_000 });
  const hostQuery = trpc.nodescope.fleet.host.useQuery({ hostId }, { enabled: !preview, refetchInterval: 5_000 });
  const host = preview ? previewQuery.data?.hosts.find((candidate) => candidate.id === hostId) : hostQuery.data;
  const loading = preview ? previewQuery.isLoading : hostQuery.isLoading;

  if (loading || !host) {
    return <DashboardLayout><div className="p-8 text-slate-400">Loading host telemetry…</div></DashboardLayout>;
  }
  const isGX10 = host.platform.includes("GX10");
  const gateway = preview ? "/preview" : "";

  return (
    <DashboardLayout>
      <div className="mx-auto max-w-[1680px] p-6 xl:p-8">
        <button onClick={() => navigate(gateway || "/")} className="mb-5 inline-flex items-center gap-2 text-xs text-slate-500 transition-colors hover:text-slate-200"><ArrowLeft className="h-3.5 w-3.5" /> Fleet overview</button>
        <header className="flex flex-col gap-5 border-b border-white/8 pb-6 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <div className="flex items-center gap-2"><span className={cn("h-2 w-2 rounded-full", host.status === "healthy" ? "bg-emerald-300" : host.status === "degraded" ? "bg-amber-300" : "bg-rose-400")} /><Badge className="border-white/10 bg-white/[0.04] text-[10px] text-slate-300">{host.role.toUpperCase()}</Badge><span className="text-xs text-slate-500">{host.freshness.state} · {host.freshness.ageSeconds}s ago</span></div>
            <h1 className="mt-3 text-3xl font-semibold tracking-[-0.035em] text-slate-50">{host.name}</h1>
            <p className="mt-2 text-sm text-slate-400">{host.platform} · {host.architecture}</p>
          </div>
          <div className="flex flex-wrap gap-2">{host.tags.map((tag) => <span key={tag} className="rounded-full border border-white/10 bg-white/[0.035] px-3 py-1.5 text-[11px] text-slate-400">{tag}</span>)}</div>
        </header>

        <Tabs defaultValue="hardware" className="mt-6">
          <TabsList className="h-auto w-full justify-start gap-1 overflow-x-auto rounded-xl border border-white/8 bg-[#0b1924] p-1">
            {[["hardware", "Hardware"], ["memory", isGX10 ? "UMA memory" : "Memory"], ["storage", "Storage"], ["processes", "Processes"], ["containers", "Containers"], ["inference", "Inference"], ["preflight", "Preflight"], ["history", "History"]].map(([value, label]) => <TabsTrigger key={value} value={value} className="whitespace-nowrap rounded-lg px-3 py-2 text-xs text-slate-400 data-[state=active]:bg-white/[0.09] data-[state=active]:text-slate-100">{label}</TabsTrigger>)}
          </TabsList>

          <TabsContent value="hardware" className="mt-6"><div className="grid gap-6 xl:grid-cols-[minmax(0,1.05fr)_360px]"><section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={Cpu} title="Compute and devices" note="Current host telemetry with source and semantics preserved." />{host.hardware.map((metric) => <MetricRow key={metric.id} metric={metric} />)}</section><aside className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={ThermometerSun} title="Current posture" note="Selected fleet-scale readings; unavailable values are never replaced with estimates." /><div className="space-y-4">{host.quickMetrics.map((metric) => <div key={metric.id} className="rounded-xl border border-white/7 bg-white/[0.025] p-3"><div className="flex items-start justify-between gap-3"><p className="text-[11px] text-slate-500">{metric.label}</p><EvidenceBadge quality={metric.quality} /></div><p className={cn("mt-2 text-lg font-medium", qualityText[metric.quality])}>{metric.display}</p><p className="mt-1 text-[10px] text-slate-600">Source: {metric.source}</p></div>)}</div></aside></div></TabsContent>

          <TabsContent value="memory" className="mt-6">{isGX10 ? <UMAPanel host={host} /> : <section className="max-w-3xl rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={MemoryStick} title="Memory" note="Host and device memory are reported from their original source." />{host.memory.map((metric) => <MetricRow key={metric.id} metric={metric} />)}</section>}</TabsContent>

          <TabsContent value="storage" className="mt-6"><section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={HardDrive} title="Mounts and volumes" note="Explicit mount state prevents missing capacity from appearing as zero." /><div className="overflow-hidden rounded-xl border border-white/7"><table className="w-full text-left text-xs"><thead className="bg-white/[0.04] text-[10px] tracking-[0.1em] text-slate-500"><tr><th className="px-4 py-3 font-medium">MOUNT</th><th className="px-4 py-3 font-medium">UTILIZATION</th><th className="px-4 py-3 font-medium">STATE</th><th className="px-4 py-3 font-medium">PROVENANCE</th></tr></thead><tbody>{host.storage.map((mount) => <tr key={mount.id} className="border-t border-white/6 text-slate-300"><td className="px-4 py-3 font-medium">{mount.mount}</td><td className="px-4 py-3">{mount.usage === null ? mount.used : `${mount.usage}% · ${mount.used} / ${mount.total}`}</td><td className={cn("px-4 py-3", mount.state === "mounted" ? "text-emerald-200" : mount.state === "learning" ? "text-violet-200" : "text-rose-200")}>{mount.state}</td><td className="px-4 py-3 text-slate-500">{mount.source} · {mount.quality}</td></tr>)}</tbody></table></div></section></TabsContent>

          <TabsContent value="processes" className="mt-6"><InventoryPanel icon={Activity} title="Selected process health" note="Only explicitly approved workloads are tracked." rows={host.processes.map((process) => ({ primary: process.name, secondary: `PID ${process.pid} · ${process.uptime}`, state: process.status, selected: process.selected }))} /></TabsContent>
          <TabsContent value="containers" className="mt-6"><InventoryPanel icon={Container} title="Container inventory" note="All discovered containers are displayed; alerts remain limited to selected containers." rows={host.containers.map((container) => ({ primary: container.name, secondary: `${container.image} · ${container.age}`, state: container.state === "running" ? "healthy" : container.state === "restarting" ? "degraded" : "unavailable", selected: container.selected }))} /></TabsContent>

	          <TabsContent value="inference" className="mt-6"><div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]"><section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={Zap} title="Inference performance" note="Request-level observation without prompt or response retention." /><div className="grid gap-3 sm:grid-cols-2">{[host.inference.requestRate, host.inference.ttft, host.inference.promptThroughput, host.inference.generationThroughput, host.inference.activeRequests].map((metric) => <div key={metric.id} className="rounded-xl border border-white/7 bg-white/[0.025] p-4"><div className="flex items-start justify-between gap-3"><p className="text-[11px] text-slate-500">{metric.label}</p><EvidenceBadge quality={metric.quality} /></div><p className={cn("mt-2 text-xl font-medium", qualityText[metric.quality])}>{metric.display}</p><p className="mt-1 text-[10px] text-slate-600">Source: {metric.source}</p></div>)}</div></section><div className="space-y-6"><RuntimeInventoryPanel runtimes={host.runtimes} /><section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={Database} title="Client usage" note="Attribution-only metadata; no content is persisted." /><div className="space-y-3">{host.inference.clientUsage.map((client) => <div key={client.client} className="rounded-xl border border-white/7 p-3"><div className="flex items-center justify-between"><p className="text-xs font-medium text-slate-300">{client.client}</p><span className="text-[11px] text-cyan-200">{client.ttft}</span></div><p className="mt-2 text-[11px] text-slate-500">{client.requests} requests · {client.promptTokens.toLocaleString()} prompt tok · {client.outputTokens.toLocaleString()} output tok</p></div>)}</div></section></div></div></TabsContent>

          <TabsContent value="preflight" className="mt-6"><section className="max-w-4xl rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={SearchCheck} title="Capability preflight" note="Missing dependencies and unsupported interfaces remain explicit." />{host.preflight.map((capability) => <div key={capability.capability} className="flex gap-3 border-b border-white/6 py-4 last:border-0"><span className={cn("mt-0.5 h-2 w-2 shrink-0 rounded-full", capability.state === "available" ? "bg-emerald-300" : capability.state === "degraded" ? "bg-amber-300" : "bg-rose-400")} /><div><p className="text-xs font-medium text-slate-300">{capability.capability}</p><p className="mt-1 text-xs leading-5 text-slate-500">{capability.detail}</p>{capability.installHint && <p className="mt-2 rounded-lg bg-amber-300/7 px-3 py-2 text-[11px] text-amber-100">Remediation: {capability.installHint}</p>}</div></div>)}</section></TabsContent>

	          <TabsContent value="history" className="mt-6"><HistoryPanel history={host.history} /></TabsContent>
        </Tabs>
      </div>
    </DashboardLayout>
  );
}

function UMAPanel({ host }: { host: any }) {
  const osMetrics = host.memory.filter((metric: any) => ["uma-os", "uma-swap", "uma-hugepages"].includes(metric.id));
  const processMetrics = host.memory.filter((metric: any) => ["uma-process", "uma-runtime", "vram"].includes(metric.id));
  return <section className="rounded-2xl border border-cyan-300/14 bg-[#0b1924] p-5"><div className="mb-5 flex flex-col gap-3 border-b border-white/8 pb-5 lg:flex-row lg:items-start lg:justify-between"><div><div className="flex items-center gap-2"><MemoryStick className="h-4 w-4 text-cyan-200" /><h2 className="text-sm font-semibold text-slate-100">Unified-memory (UMA) panel</h2></div><p className="mt-2 max-w-2xl text-xs leading-5 text-slate-500">OS-reclaimable memory, swap, huge pages, runtime allocation, and per-process GPU memory are intentionally displayed side-by-side. These sources are not interchangeable, and dedicated VRAM is never synthesized.</p></div><Badge className="w-fit border-cyan-300/15 bg-cyan-300/8 text-[10px] text-cyan-100">GX10 PROVENANCE VIEW</Badge></div><div className="grid gap-5 lg:grid-cols-2"><div className="rounded-xl border border-white/7 bg-white/[0.025] p-4"><p className="mb-2 text-[10px] font-medium tracking-[0.13em] text-slate-500">OS MEMORY AND RESERVATIONS</p>{osMetrics.map((metric: any) => <MetricRow key={metric.id} metric={metric} />)}</div><div className="rounded-xl border border-white/7 bg-white/[0.025] p-4"><p className="mb-2 text-[10px] font-medium tracking-[0.13em] text-slate-500">RUNTIME AND PER-PROCESS GPU VIEW</p>{processMetrics.map((metric: any) => <MetricRow key={metric.id} metric={metric} />)}</div></div></section>;
}

export function HistoryPanel({ history }: { history: Array<{ timestamp: string; throughput: number; cpu: number }> }) {
  return <section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={Network} title="Five-minute history" note="Source samples are retained only under the configured raw-retention and rollup policy." />{history.length === 0 ? <div role="status" className="grid min-h-[340px] place-items-center rounded-xl border border-dashed border-white/10 bg-white/[0.02] px-6 text-center"><div className="max-w-md"><p className="text-sm font-medium text-slate-200">No retained telemetry samples</p><p className="mt-2 text-xs leading-5 text-slate-500">This host has no history in the selected window. NodeScope does not synthesize chart points from current values, unavailable evidence, or estimates.</p><p className="mt-3 text-[11px] text-slate-600">Check host freshness, collection configuration, and the raw-retention window.</p></div></div> : <><div className="mb-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-[10px] text-slate-500"><span className="inline-flex items-center gap-2"><span className="h-1.5 w-1.5 rounded-full bg-cyan-200" />Generation throughput</span><span className="inline-flex items-center gap-2"><span className="h-1.5 w-1.5 rounded-full bg-indigo-300" />CPU utilization</span><span>{history.length} retained samples</span></div><div className="h-[304px]"><ResponsiveContainer width="100%" height="100%"><AreaChart data={history} margin={{ top: 10, right: 8, bottom: 0, left: -18 }}><defs><linearGradient id="history-throughput" x1="0" x2="0" y1="0" y2="1"><stop offset="0%" stopColor="#67e8f9" stopOpacity={0.35} /><stop offset="100%" stopColor="#67e8f9" stopOpacity={0} /></linearGradient></defs><CartesianGrid stroke="rgba(255,255,255,.06)" vertical={false} /><XAxis dataKey="timestamp" tickFormatter={(value) => new Date(value).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })} tick={{ fill: "#64748b", fontSize: 10 }} axisLine={false} tickLine={false} /><YAxis tick={{ fill: "#64748b", fontSize: 10 }} axisLine={false} tickLine={false} /><Tooltip contentStyle={{ background: "#0b1924", border: "1px solid rgba(255,255,255,.12)", borderRadius: 12, fontSize: 12 }} labelFormatter={(value) => new Date(value).toLocaleString()} /><Area type="monotone" dataKey="throughput" stroke="#67e8f9" strokeWidth={2} fill="url(#history-throughput)" name="Generation throughput" /><Area type="monotone" dataKey="cpu" stroke="#818cf8" strokeWidth={1.5} fill="transparent" name="CPU utilization" /></AreaChart></ResponsiveContainer></div></>}</section>;
}

export function RuntimeInventoryPanel({ runtimes }: { runtimes: RuntimeInventoryEntry[] }) {
	if (runtimes.length === 0) {
		return <section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={Network} title="Runtime inventory" note="Approval and availability evidence only; endpoint locations are intentionally withheld." /><div role="status" className="rounded-xl border border-dashed border-white/10 bg-white/[0.02] px-4 py-5"><p className="text-xs font-medium text-slate-300">{runtimeInventoryEmptyState.title}</p><p className="mt-2 text-[11px] leading-5 text-slate-500">{runtimeInventoryEmptyState.detail}</p></div></section>;
	}
	return <section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={Network} title="Runtime inventory" note="Approval and availability evidence only; endpoint locations are intentionally withheld." /><div className="overflow-hidden rounded-xl border border-white/7">{runtimes.map((runtime) => {
		const display = runtimeInventoryDisplay(runtime);
		return <div key={`${runtime.kind}-${runtime.endpoint}`} className="flex items-center justify-between gap-4 border-b border-white/6 px-4 py-3 last:border-0"><div><p className="text-xs font-medium text-slate-200">{display.kind}</p><p className={cn("mt-1 text-[10px]", display.stateTone === "approved" ? "text-cyan-200" : display.stateTone === "discovered" ? "text-violet-200" : "text-rose-200")}>{display.approvalLabel}</p></div><div className="flex items-center gap-2"><span className={cn("h-1.5 w-1.5 rounded-full", display.healthTone === "healthy" ? "bg-emerald-300" : display.healthTone === "degraded" ? "bg-amber-300" : "bg-rose-400")} /><span className="text-[10px] text-slate-400">{display.healthLabel}</span></div></div>;
	})}</div></section>;
}

function InventoryPanel({ icon, title, note, rows }: { icon: typeof Activity; title: string; note: string; rows: Array<{ primary: string; secondary: string; state: string; selected: boolean }> }) {
  return <section className="max-w-4xl rounded-2xl border border-white/8 bg-[#0b1924] p-5"><SectionHeader icon={icon} title={title} note={note} /><div className="overflow-hidden rounded-xl border border-white/7">{rows.map((row) => <div key={row.primary} className="flex items-center justify-between border-b border-white/6 px-4 py-3 last:border-0"><div><p className="text-xs font-medium text-slate-300">{row.primary}</p><p className="mt-1 text-[10px] text-slate-600">{row.secondary}</p></div><div className="flex items-center gap-3"><span className={cn("text-[10px]", row.selected ? "text-cyan-200" : "text-slate-600")}>{row.selected ? "selected" : "inventory"}</span><span className={cn("h-1.5 w-1.5 rounded-full", row.state === "healthy" ? "bg-emerald-300" : row.state === "degraded" ? "bg-amber-300" : "bg-rose-400")} /></div></div>)}</div></section>;
}
