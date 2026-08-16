import DashboardLayout from "@/components/DashboardLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { metricQualityAriaLabel, metricQualityLabel, type MetricQuality } from "@/lib/evidenceState";
import { cn } from "@/lib/utils";
import { trpc } from "@/lib/trpc";
import {
  Activity,
  AlertTriangle,
  ArrowUpRight,
  Bot,
  CheckCircle2,
  ChevronRight,
  Clock3,
  DatabaseBackup,
  Gauge,
  HardDrive,
  MemoryStick,
  Network,
  Server,
  ShieldAlert,
  Sparkles,
  RefreshCw,
  ThermometerSun,
  TriangleAlert,
  Zap,
} from "lucide-react";
import { useState } from "react";
import { useLocation } from "wouter";

type Quality = MetricQuality;

const qualityStyles: Record<Quality, string> = {
  fresh: "border-emerald-300/15 bg-emerald-400/10 text-emerald-200",
  stale: "border-amber-300/20 bg-amber-300/10 text-amber-100",
  unavailable: "border-rose-300/20 bg-rose-400/10 text-rose-100",
  unsupported: "border-slate-400/15 bg-slate-400/10 text-slate-300",
  estimated: "border-violet-300/20 bg-violet-400/10 text-violet-100",
  experimental: "border-fuchsia-300/20 bg-fuchsia-400/10 text-fuchsia-100",
};

const statusStyles = {
  healthy: "bg-emerald-300",
  degraded: "bg-amber-300",
  unavailable: "bg-rose-400",
};

function QualityBadge({ quality }: { quality: Quality }) {
  return <span aria-label={metricQualityAriaLabel(quality)} className={cn("inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-medium tracking-[0.08em] uppercase", qualityStyles[quality])}>{metricQualityLabel(quality)}</span>;
}

function Freshness({ state, ageSeconds }: { state: "fresh" | "stale" | "unavailable"; ageSeconds: number }) {
  const label = state === "fresh" ? `${ageSeconds}s ago` : state === "stale" ? `stale · ${ageSeconds}s` : "unavailable";
  return (
    <span className={cn("inline-flex items-center gap-1.5 text-xs", state === "fresh" ? "text-emerald-200" : state === "stale" ? "text-amber-100" : "text-rose-200")}>
      <span className={cn("h-1.5 w-1.5 rounded-full", state === "fresh" ? "bg-emerald-300 shadow-[0_0_12px_rgba(110,231,183,.9)]" : state === "stale" ? "bg-amber-300" : "bg-rose-400")} />
      {label}
    </span>
  );
}

function OverviewSkeleton() {
  return (
    <div className="space-y-6 p-8">
      <Skeleton className="h-20 w-full bg-white/5" />
      <div className="grid grid-cols-4 gap-4"><Skeleton className="h-32 bg-white/5" /><Skeleton className="h-32 bg-white/5" /><Skeleton className="h-32 bg-white/5" /><Skeleton className="h-32 bg-white/5" /></div>
      <Skeleton className="h-[420px] bg-white/5" />
    </div>
  );
}

export default function FleetOverview({ preview = false }: { preview?: boolean }) {
  const [, navigate] = useLocation();
  const [refreshing, setRefreshing] = useState(false);
  const previewQuery = trpc.nodescope.fleet.preview.useQuery(undefined, { enabled: preview, refetchInterval: 5_000 });
  const liveQuery = trpc.nodescope.fleet.overview.useQuery(undefined, { enabled: !preview, refetchInterval: 5_000 });
  const query = preview ? previewQuery : liveQuery;

  if (query.isLoading || !query.data) {
    return <DashboardLayout><OverviewSkeleton /></DashboardLayout>;
  }
  if (query.isError) {
    return <DashboardLayout><div className="p-8 text-rose-200">Fleet telemetry could not be loaded. The console has retained no substitute values.</div></DashboardLayout>;
  }

  const fleet = query.data;
  const targetPath = (path: string) => preview ? `/preview${path === "/" ? "" : path}` : path;
  const healthyHosts = fleet.hosts.filter((host) => host.status === "healthy").length;
  const totalThroughput = fleet.hosts.reduce((sum, host) => sum + (host.inference.generationThroughput.value ?? 0), 0);
  const maxTemperature = Math.max(...fleet.hosts.flatMap((host) => host.quickMetrics.filter((metric) => metric.id === "temp").map((metric) => metric.value ?? 0)));
  const refreshFleet = async () => {
    setRefreshing(true);
    try {
      await query.refetch();
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="mx-auto max-w-[1680px] p-6 xl:p-8">
        <section className="flex flex-col gap-5 border-b border-white/8 pb-6 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <div className="mb-3 flex items-center gap-2">
              {fleet.developmentPreview && <Badge className="border-0 bg-amber-300/10 text-[10px] tracking-[0.13em] text-amber-100">DEVELOPMENT PREVIEW</Badge>}
              <span className="text-xs text-slate-500">Updated {new Date(fleet.generatedAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })}</span>
            </div>
            <h1 className="text-3xl font-semibold tracking-[-0.035em] text-slate-50">Fleet posture</h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">A precise view of availability, capacity, model serving, and operational risk across your local compute fleet.</p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <div className="rounded-xl border border-white/8 bg-white/[0.035] px-3 py-2 text-right">
              <p className="text-[10px] font-medium tracking-[0.12em] text-slate-500">COLLECTION</p>
              <p className="mt-0.5 text-xs font-medium text-slate-200">{fleet.globalIntervalSeconds}s global interval</p>
            </div>
            <Button
              variant="outline"
              onClick={refreshFleet}
              disabled={refreshing}
              className="h-10 border-white/10 text-slate-300 hover:bg-white/[0.06]"
              aria-label="Refresh fleet telemetry"
            >
              <RefreshCw className={cn("mr-2 h-3.5 w-3.5", refreshing && "animate-spin")} />
              {refreshing ? "Refreshing" : "Refresh"}
            </Button>
            <Button onClick={() => navigate(targetPath("/operations"))} className="h-10 bg-cyan-300 px-4 text-[#071018] hover:bg-cyan-200">
              Review operations <ChevronRight className="ml-1 h-4 w-4" />
            </Button>
          </div>
          <span className="sr-only" aria-live="polite">{refreshing ? "Refreshing fleet telemetry" : "Fleet telemetry refresh idle"}</span>
        </section>

        <section className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <PostureMetric onClick={() => navigate(targetPath("/hosts"))} icon={Server} label="Host availability" value={`${healthyHosts}/${fleet.hosts.length}`} note="Select a host to inspect freshness and evidence" tone="emerald" />
          <PostureMetric onClick={() => navigate(targetPath("/alerts"))} icon={AlertTriangle} label="Active alerts" value={String(fleet.alerts.filter((alert) => alert.state === "active").length)} note="Open the alert queue" tone={fleet.activeAlertCount ? "amber" : "emerald"} />
          <PostureMetric onClick={() => navigate(targetPath("/operations"))} icon={Zap} label="Generation throughput" value={`${Math.round(totalThroughput)} tok/s`} note="Review runtime and proxy operations" tone="cyan" />
          <PostureMetric onClick={() => navigate(targetPath("/hosts"))} icon={ThermometerSun} label="Peak temperature" value={`${maxTemperature}°C`} note="Select a host to inspect device evidence" tone={maxTemperature > 80 ? "amber" : "cyan"} />
        </section>

        <div className="mt-7 grid gap-6 2xl:grid-cols-[minmax(0,1.65fr)_390px]">
          <section>
            <div className="mb-3 flex items-center justify-between px-1">
              <div>
                <p className="text-sm font-semibold text-slate-100">Compute hosts</p>
                <p className="mt-1 text-xs text-slate-500">Freshness, utilization, specialized devices, and selected services.</p>
              </div>
              <button onClick={() => navigate(targetPath("/hosts"))} className="text-xs font-medium text-cyan-200 transition-colors hover:text-cyan-100">Choose a host</button>
            </div>
            <div className="grid gap-4 xl:grid-cols-2">
              {fleet.hosts.map((host) => (
                <article key={host.id} className="group overflow-hidden rounded-2xl border border-white/8 bg-[#0b1924] shadow-[0_18px_55px_rgba(0,0,0,.18)] transition-transform duration-200 hover:-translate-y-0.5 hover:border-cyan-200/20">
                  <div className="flex items-start justify-between border-b border-white/8 px-5 pb-4 pt-5">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className={cn("h-2 w-2 rounded-full", statusStyles[host.status])} />
                        <h2 className="font-semibold tracking-tight text-slate-100">{host.name}</h2>
                        <Badge className="border-white/10 bg-white/[0.04] text-[10px] font-medium text-slate-300">{host.role === "preferred" ? "PRIMARY" : "SECONDARY"}</Badge>
                      </div>
                      <p className="mt-1 text-xs text-slate-500">{host.platform} · {host.architecture}</p>
                    </div>
                    <Freshness {...host.freshness} />
                  </div>
                  <div className="grid grid-cols-2 gap-px bg-white/7">
                    {host.quickMetrics.slice(0, 6).map((metric) => (
                      <div key={metric.id} className="min-h-[92px] bg-[#0b1924] p-4">
                        <div className="flex items-start justify-between gap-2">
                          <p className="text-[11px] text-slate-500">{metric.label}</p>
                          <QualityBadge quality={metric.quality} />
                        </div>
                        <p className={cn("mt-3 text-base font-medium tracking-tight", metric.quality === "fresh" ? "text-slate-100" : "text-slate-300")}>{metric.display}</p>
                        <p className="mt-1 truncate text-[10px] text-slate-600">{metric.source}</p>
                      </div>
                    ))}
                  </div>
                  <div className="flex items-center justify-between px-5 py-4">
                    <div className="flex -space-x-1.5">
                      {host.tags.slice(0, 3).map((tag) => <span key={tag} className="rounded-full border border-[#0b1924] bg-white/[0.08] px-2 py-1 text-[10px] text-slate-400">{tag}</span>)}
                    </div>
                    <button onClick={() => navigate(targetPath(`/hosts/${host.id}`))} className="inline-flex items-center gap-1 text-xs font-medium text-cyan-200 transition-colors hover:text-cyan-100">Inspect <ArrowUpRight className="h-3.5 w-3.5" /></button>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <aside className="space-y-6">
            <section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5">
              <div className="flex items-center justify-between">
                <div><p className="text-sm font-semibold text-slate-100">Needs attention</p><p className="mt-1 text-xs text-slate-500">Live operational conditions</p></div>
                <span className="grid h-8 w-8 place-items-center rounded-xl bg-amber-300/10 text-amber-200"><ShieldAlert className="h-4 w-4" /></span>
              </div>
              <div className="mt-5 space-y-3">
				{fleet.alerts.map((alert) => (
				  <button key={alert.id} onClick={() => navigate(targetPath("/alerts"))} className={cn("w-full rounded-xl border p-3.5 text-left transition hover:border-cyan-200/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300", alert.severity === "critical" ? "border-rose-300/20 bg-rose-400/8" : "border-amber-300/15 bg-amber-300/6")}>
					<div className="flex items-start gap-2">
					  <TriangleAlert className={cn("mt-0.5 h-3.5 w-3.5 shrink-0", alert.severity === "critical" ? "text-rose-300" : "text-amber-200")} />
					  <div><p className="text-xs font-medium text-slate-200">{alert.title}</p><p className="mt-1 text-[11px] leading-4 text-slate-500">{alert.detail}</p><span className="mt-2 inline-block text-[10px] font-medium text-cyan-200">Open alert details</span></div>
					</div>
				  </button>
				))}
              </div>
              <button onClick={() => navigate(targetPath("/alerts"))} className="mt-4 w-full rounded-lg border border-white/10 py-2 text-xs font-medium text-slate-300 transition-colors hover:bg-white/[0.06]">Open alert queue</button>
            </section>

            <section className="rounded-2xl border border-white/8 bg-[#0b1924] p-5">
              <div className="flex items-center gap-2"><Network className="h-4 w-4 text-cyan-300" /><p className="text-sm font-semibold text-slate-100">Replica integrity</p></div>
              <div className="mt-4 divide-y divide-white/7">
				{fleet.replicas.map((replica) => {
				  const host = fleet.hosts.find((candidate) => candidate.id === replica.hostId);
				  return <button key={replica.id} onClick={() => navigate(targetPath(`/hosts/${replica.hostId}`))} className="w-full py-3 text-left transition first:pt-0 last:pb-0 hover:bg-white/[0.03] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300">
					<div className="flex items-center justify-between"><div className="flex items-center gap-2"><span className={cn("h-1.5 w-1.5 rounded-full", statusStyles[replica.status])} /><span className="text-xs font-medium text-slate-300">{host?.name} · {replica.role}</span></div><span className="text-[10px] text-slate-500">{replica.version}</span></div>
					<div className="mt-2 flex flex-wrap gap-2"><span className="text-[10px] text-slate-500">Cert {replica.certificateDaysRemaining}d</span><span className="text-[10px] text-slate-500">Backup {replica.backupFreshness}</span><span className={cn("text-[10px]", replica.sharedBackupMount === "mounted" ? "text-emerald-200" : "text-amber-200")}>Mount {replica.sharedBackupMount}</span><span className="text-[10px] text-cyan-200">Inspect host</span></div>
				  </button>;
				})}
              </div>
            </section>

            <section className="rounded-2xl border border-cyan-300/10 bg-gradient-to-br from-cyan-300/[0.08] to-transparent p-5">
			  <button onClick={() => navigate(targetPath("/operations"))} className="w-full text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300"><div className="flex gap-3"><span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-cyan-300/12 text-cyan-200"><Bot className="h-4 w-4" /></span><div><p className="text-sm font-medium text-slate-100">Inference observability</p><p className="mt-1 text-xs leading-5 text-slate-500">Client usage is attributed without retaining any prompt or response content.</p><span className="mt-2 inline-block text-[11px] font-medium text-cyan-200">Open operations</span></div></div></button>
              <div className="mt-4 flex items-center gap-2 text-[11px] text-cyan-100"><CheckCircle2 className="h-3.5 w-3.5" /> No content retention boundary active</div>
            </section>
          </aside>
        </div>
      </div>
    </DashboardLayout>
  );
}

function PostureMetric({ icon: Icon, label, value, note, tone, onClick }: { icon: typeof Server; label: string; value: string; note: string; tone: "emerald" | "amber" | "cyan"; onClick: () => void }) {
  const colors = {
    emerald: "bg-emerald-300/10 text-emerald-200",
    amber: "bg-amber-300/10 text-amber-100",
    cyan: "bg-cyan-300/10 text-cyan-100",
  };
  return <button onClick={onClick} className="group rounded-2xl border border-white/8 bg-[#0b1924] p-4 text-left transition hover:-translate-y-0.5 hover:border-cyan-200/25 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300"><div className="flex items-start justify-between"><p className="text-xs text-slate-500">{label}</p><span className={cn("grid h-8 w-8 place-items-center rounded-xl", colors[tone])}><Icon className="h-4 w-4" /></span></div><p className="mt-5 text-2xl font-semibold tracking-[-0.035em] text-slate-100">{value}</p><p className="mt-1.5 text-[11px] leading-4 text-slate-500">{note}</p><span className="mt-3 inline-flex items-center gap-1 text-[11px] font-medium text-cyan-200 opacity-80 transition group-hover:opacity-100">Open view <ArrowUpRight className="h-3 w-3" /></span></button>;
}
