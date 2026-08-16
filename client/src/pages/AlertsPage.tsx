import DashboardLayout from "@/components/DashboardLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { trpc } from "@/lib/trpc";
import { AlertTriangle, Check, Clock3, ShieldAlert } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { useLocation } from "wouter";

const stateFilters = ["all", "active", "acknowledged"] as const;
const severityFilters = ["all", "critical", "warning"] as const;
type AlertStateFilter = (typeof stateFilters)[number];
type AlertSeverityFilter = (typeof severityFilters)[number];

export default function AlertsPage({ preview = false }: { preview?: boolean }) {
  const [, navigate] = useLocation();
  const [stateFilter, setStateFilter] = useState<AlertStateFilter>("all");
  const [severityFilter, setSeverityFilter] = useState<AlertSeverityFilter>("all");
  const previewQuery = trpc.nodescope.fleet.preview.useQuery(undefined, { enabled: preview, refetchInterval: 5_000 });
  const liveQuery = trpc.nodescope.fleet.overview.useQuery(undefined, { enabled: !preview, refetchInterval: 5_000 });
  const acknowledge = trpc.nodescope.alerts.acknowledge.useMutation({
    onSuccess: () => {
      toast.success("Alert acknowledgement recorded", { description: "The action has been written to the audit trail." });
      liveQuery.refetch();
    },
    onError: (error) => toast.error("Unable to acknowledge alert", { description: error.message }),
  });
  const fleet = preview ? previewQuery.data : liveQuery.data;
  const alerts = fleet?.alerts ?? [];
  const filteredAlerts = useMemo(
    () => alerts.filter((alert) => (stateFilter === "all" || alert.state === stateFilter) && (severityFilter === "all" || alert.severity === severityFilter)),
    [alerts, severityFilter, stateFilter],
  );
  const clearFilters = () => {
    setStateFilter("all");
    setSeverityFilter("all");
  };
  const filtersActive = stateFilter !== "all" || severityFilter !== "all";

  if (!fleet) return <DashboardLayout><div className="p-8 text-slate-400">Loading alert queue…</div></DashboardLayout>;

  const targetPath = (path: string) => preview ? `/preview${path}` : path;
  const activeCount = alerts.filter((alert) => alert.state === "active").length;
  const acknowledgedCount = alerts.filter((alert) => alert.state === "acknowledged").length;

  return (
    <DashboardLayout>
      <div className="mx-auto max-w-[1320px] p-6 xl:p-8">
        <header className="flex flex-col gap-4 border-b border-white/8 pb-6 md:flex-row md:items-end md:justify-between">
          <div>
            <div className="flex items-center gap-2 text-amber-100"><ShieldAlert className="h-4 w-4" /><span className="text-xs font-medium tracking-[0.12em]">OPERATOR QUEUE</span></div>
            <h1 className="mt-3 text-3xl font-semibold tracking-[-0.035em] text-slate-50">Alerts</h1>
            <p className="mt-2 text-sm text-slate-400">Conditions remain visible until resolved or explicitly acknowledged. No data-quality state is silently suppressed.</p>
          </div>
          <div className="flex gap-3">
            <div className="rounded-xl border border-white/8 bg-white/[0.035] px-4 py-3"><p className="text-[10px] tracking-[0.12em] text-slate-500">ACTIVE</p><p className="mt-1 text-xl font-semibold text-amber-100">{activeCount}</p></div>
            <div className="rounded-xl border border-white/8 bg-white/[0.035] px-4 py-3"><p className="text-[10px] tracking-[0.12em] text-slate-500">ACKNOWLEDGED</p><p className="mt-1 text-xl font-semibold text-slate-200">{acknowledgedCount}</p></div>
          </div>
        </header>

        <section aria-label="Alert queue controls" className="mt-6 rounded-2xl border border-white/8 bg-white/[0.025] p-4">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <div className="flex flex-wrap items-center gap-2" aria-label="Filter alerts by state">
              <span className="mr-1 text-[10px] font-medium tracking-[0.12em] text-slate-500">STATE</span>
              {stateFilters.map((filter) => (
                <button
                  type="button"
                  key={filter}
                  onClick={() => setStateFilter(filter)}
                  aria-pressed={stateFilter === filter}
                  className={cn("rounded-lg border px-3 py-2 text-xs font-medium capitalize transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300", stateFilter === filter ? "border-cyan-300/30 bg-cyan-300/10 text-cyan-100" : "border-white/8 bg-white/[0.025] text-slate-400 hover:bg-white/[0.06] hover:text-slate-200")}
                >
                  {filter}
                </button>
              ))}
            </div>
            <div className="flex flex-wrap items-center gap-2" aria-label="Filter alerts by severity">
              <span className="mr-1 text-[10px] font-medium tracking-[0.12em] text-slate-500">SEVERITY</span>
              {severityFilters.map((filter) => (
                <button
                  type="button"
                  key={filter}
                  onClick={() => setSeverityFilter(filter)}
                  aria-pressed={severityFilter === filter}
                  className={cn("rounded-lg border px-3 py-2 text-xs font-medium capitalize transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300", severityFilter === filter ? "border-cyan-300/30 bg-cyan-300/10 text-cyan-100" : "border-white/8 bg-white/[0.025] text-slate-400 hover:bg-white/[0.06] hover:text-slate-200")}
                >
                  {filter}
                </button>
              ))}
              {filtersActive && <button type="button" onClick={clearFilters} className="px-2 text-xs font-medium text-cyan-200 transition hover:text-cyan-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300">Clear filters</button>}
            </div>
          </div>
          <p aria-live="polite" className="mt-3 text-xs text-slate-500">Showing {filteredAlerts.length} of {alerts.length} alerts.</p>
        </section>

        {filteredAlerts.length === 0 ? (
          <section className="mt-4 rounded-2xl border border-dashed border-white/10 bg-white/[0.025] p-8 text-center">
            <p className="text-sm font-medium text-slate-200">No alerts match the current queue filters.</p>
            <p className="mt-2 text-xs leading-5 text-slate-500">NodeScope has not removed or changed alert evidence; clear the filters to return to the full queue.</p>
            <Button variant="outline" onClick={clearFilters} className="mt-5 border-white/10 text-slate-300 hover:bg-white/[0.06]">Clear alert filters</Button>
          </section>
        ) : (
          <div className="mt-4 space-y-3">
            {filteredAlerts.map((alert) => {
              const host = fleet.hosts.find((candidate) => candidate.id === alert.hostId);
              const active = alert.state === "active";
              return (
                <article key={alert.id} className={cn("rounded-2xl border p-5", alert.severity === "critical" ? "border-rose-300/20 bg-rose-400/7" : "border-amber-300/15 bg-[#0b1924]")}>
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                    <div className="flex gap-3">
                      <span className={cn("grid h-9 w-9 shrink-0 place-items-center rounded-xl", alert.severity === "critical" ? "bg-rose-400/12 text-rose-200" : "bg-amber-300/10 text-amber-100")}><AlertTriangle className="h-4 w-4" /></span>
                      <div>
                        <div className="flex flex-wrap items-center gap-2"><h2 className="text-sm font-semibold text-slate-100">{alert.title}</h2><Badge className={cn("border-0 text-[10px]", alert.state === "acknowledged" ? "bg-slate-400/10 text-slate-300" : alert.severity === "critical" ? "bg-rose-400/12 text-rose-100" : "bg-amber-300/10 text-amber-100")}>{alert.state}</Badge></div>
                        <p className="mt-2 max-w-2xl text-xs leading-5 text-slate-400">{alert.detail}</p>
                        <div className="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-[11px] text-slate-500"><span className="inline-flex items-center gap-1.5"><Clock3 className="h-3.5 w-3.5" />{new Date(alert.observedAt).toLocaleString()}</span><button onClick={() => navigate(targetPath(`/hosts/${alert.hostId}`))} className="text-cyan-200 hover:text-cyan-100">{host?.name ?? alert.hostId}</button></div>
                      </div>
                    </div>
                    {active && <Button disabled={preview || acknowledge.isPending} onClick={() => acknowledge.mutate({ alertId: alert.id })} variant="outline" className="border-white/10 bg-white/[0.035] text-slate-200 hover:bg-white/[0.08] disabled:opacity-50"><Check className="mr-1.5 h-4 w-4" />{preview ? "Preview only" : "Acknowledge"}</Button>}
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
