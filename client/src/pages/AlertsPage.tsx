import DashboardLayout from "@/components/DashboardLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { trpc } from "@/lib/trpc";
import { AlertTriangle, Check, Clock3, ShieldAlert } from "lucide-react";
import { toast } from "sonner";
import { useLocation } from "wouter";

export default function AlertsPage({ preview = false }: { preview?: boolean }) {
  const [, navigate] = useLocation();
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
  if (!fleet) return <DashboardLayout><div className="p-8 text-slate-400">Loading alert queue…</div></DashboardLayout>;

  const targetPath = (path: string) => preview ? `/preview${path}` : path;
  return <DashboardLayout><div className="mx-auto max-w-[1320px] p-6 xl:p-8">
    <header className="flex flex-col gap-4 border-b border-white/8 pb-6 md:flex-row md:items-end md:justify-between"><div><div className="flex items-center gap-2 text-amber-100"><ShieldAlert className="h-4 w-4" /><span className="text-xs font-medium tracking-[0.12em]">OPERATOR QUEUE</span></div><h1 className="mt-3 text-3xl font-semibold tracking-[-0.035em] text-slate-50">Alerts</h1><p className="mt-2 text-sm text-slate-400">Conditions remain visible until resolved or explicitly acknowledged. No data-quality state is silently suppressed.</p></div><div className="rounded-xl border border-white/8 bg-white/[0.035] px-4 py-3"><p className="text-[10px] tracking-[0.12em] text-slate-500">ACTIVE</p><p className="mt-1 text-xl font-semibold text-amber-100">{fleet.alerts.filter((alert) => alert.state === "active").length}</p></div></header>
    <div className="mt-6 space-y-3">{fleet.alerts.map((alert) => { const host = fleet.hosts.find((candidate) => candidate.id === alert.hostId); const active = alert.state === "active"; return <article key={alert.id} className={cn("rounded-2xl border p-5", alert.severity === "critical" ? "border-rose-300/20 bg-rose-400/7" : "border-amber-300/15 bg-[#0b1924]")}><div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"><div className="flex gap-3"><span className={cn("grid h-9 w-9 shrink-0 place-items-center rounded-xl", alert.severity === "critical" ? "bg-rose-400/12 text-rose-200" : "bg-amber-300/10 text-amber-100")}><AlertTriangle className="h-4 w-4" /></span><div><div className="flex flex-wrap items-center gap-2"><h2 className="text-sm font-semibold text-slate-100">{alert.title}</h2><Badge className={cn("border-0 text-[10px]", alert.state === "acknowledged" ? "bg-slate-400/10 text-slate-300" : alert.severity === "critical" ? "bg-rose-400/12 text-rose-100" : "bg-amber-300/10 text-amber-100")}>{alert.state}</Badge></div><p className="mt-2 max-w-2xl text-xs leading-5 text-slate-400">{alert.detail}</p><div className="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-[11px] text-slate-500"><span className="inline-flex items-center gap-1.5"><Clock3 className="h-3.5 w-3.5" />{new Date(alert.observedAt).toLocaleString()}</span><button onClick={() => navigate(targetPath(`/hosts/${alert.hostId}`))} className="text-cyan-200 hover:text-cyan-100">{host?.name ?? alert.hostId}</button></div></div></div>{active && <Button disabled={preview || acknowledge.isPending} onClick={() => acknowledge.mutate({ alertId: alert.id })} variant="outline" className="border-white/10 bg-white/[0.035] text-slate-200 hover:bg-white/[0.08] disabled:opacity-50"><Check className="mr-1.5 h-4 w-4" />{preview ? "Preview only" : "Acknowledge"}</Button>}</div></article>; })}</div>
  </div></DashboardLayout>;
}
