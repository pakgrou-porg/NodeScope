import DashboardLayout from "@/components/DashboardLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import { trpc } from "@/lib/trpc";
import { BellRing, CircleGauge, Save, ShieldAlert } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";

type EditableRule = {
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

export default function AlertRulesPage({ preview = false }: { preview?: boolean }) {
  const previewQuery = trpc.nodescope.alerts.rules.preview.useQuery(undefined, { enabled: preview });
  const liveQuery = trpc.nodescope.alerts.rules.list.useQuery(undefined, { enabled: !preview });
  const rules = (preview ? previewQuery.data : liveQuery.data) as EditableRule[] | undefined;
  const [drafts, setDrafts] = useState<Record<string, EditableRule>>({});
  const save = trpc.nodescope.alerts.rules.save.useMutation({
    onSuccess: () => { toast.success("Alert rule saved", { description: "The server recorded the policy update and audit event." }); liveQuery.refetch(); },
    onError: (error) => toast.error("Unable to save alert rule", { description: error.message }),
  });
  const effectiveRules = useMemo(() => rules?.map((rule) => drafts[rule.id] ?? rule) ?? [], [rules, drafts]);
  const patch = (id: string, values: Partial<EditableRule>) => setDrafts((current) => ({ ...current, [id]: { ...(current[id] ?? effectiveRules.find((rule) => rule.id === id)!), ...values } }));
  if (!rules) return <DashboardLayout><div className="p-8 text-slate-400">Loading alert rules…</div></DashboardLayout>;

  return <DashboardLayout><div className="mx-auto max-w-[1320px] p-6 xl:p-8"><header className="flex flex-col gap-4 border-b border-white/8 pb-6 md:flex-row md:items-end md:justify-between"><div><div className="flex items-center gap-2 text-amber-100"><BellRing className="h-4 w-4" /><span className="text-xs font-medium tracking-[0.12em]">PLATFORM-AWARE ALERTING</span></div><h1 className="mt-3 text-3xl font-semibold tracking-[-0.035em] text-slate-50">Alert rules</h1><p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">Defaults are explicit, editable, and evaluated against real sources. Unavailable values are never converted into zero or a passing condition.</p></div>{preview && <Badge className="w-fit border-0 bg-amber-300/10 text-[10px] tracking-[0.12em] text-amber-100">PREVIEW · SAVES DISABLED</Badge>}</header><section className="mt-6 rounded-2xl border border-white/8 bg-[#0b1924] p-5"><div className="flex items-start gap-3"><span className="grid h-9 w-9 place-items-center rounded-xl bg-amber-300/10 text-amber-100"><ShieldAlert className="h-4 w-4" /></span><div><h2 className="text-sm font-semibold text-slate-100">Active policy set</h2><p className="mt-1 text-xs leading-5 text-slate-500">A rule may be scoped fleet-wide or to one host. Evaluation duration prevents isolated short-lived spikes from generating noise.</p></div></div><div className="mt-5 space-y-3">{effectiveRules.map((rule) => <article key={rule.id} className="rounded-xl border border-white/7 bg-white/[0.025] p-4"><div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_110px_120px_130px_110px_auto]"><div><p className="text-xs font-medium text-slate-200">{rule.metric}</p><p className="mt-1 text-[10px] text-slate-600">{rule.scope === "fleet" ? "Fleet-wide" : `Host: ${rule.hostId}`}</p></div><select value={rule.operator} onChange={(event) => patch(rule.id, { operator: event.target.value as EditableRule["operator"] })} disabled={preview} className="h-9 rounded-lg border border-white/10 bg-white/[0.035] px-2 text-xs text-slate-200 disabled:opacity-50"><option value="gt">greater than</option><option value="lt">less than</option></select><Input type="number" value={rule.threshold} onChange={(event) => patch(rule.id, { threshold: Number(event.target.value) })} disabled={preview} className="h-9 border-white/10 bg-white/[0.035] text-xs text-slate-200 disabled:opacity-50" /><Input type="number" value={rule.durationSeconds} onChange={(event) => patch(rule.id, { durationSeconds: Number(event.target.value) })} disabled={preview} className="h-9 border-white/10 bg-white/[0.035] text-xs text-slate-200 disabled:opacity-50" /><select value={rule.severity} onChange={(event) => patch(rule.id, { severity: event.target.value as EditableRule["severity"] })} disabled={preview} className="h-9 rounded-lg border border-white/10 bg-white/[0.035] px-2 text-xs text-slate-200 disabled:opacity-50"><option value="info">info</option><option value="warning">warning</option><option value="critical">critical</option></select><div className="flex items-center justify-end gap-3"><Switch checked={rule.enabled} onCheckedChange={(checked) => patch(rule.id, { enabled: checked })} disabled={preview} /><Button size="sm" disabled={preview || save.isPending} onClick={() => save.mutate(rule)} className="h-9 bg-cyan-300 text-[#071018] hover:bg-cyan-200"><Save className="mr-1.5 h-3.5 w-3.5" />Save</Button></div></div><div className="mt-3 flex items-center gap-2 text-[10px] text-slate-600"><CircleGauge className="h-3.5 w-3.5" /> Threshold {rule.operator === "gt" ? "exceeds" : "falls below"} {rule.threshold} for {rule.durationSeconds}s · {rule.severity}</div></article>)}</div></section></div></DashboardLayout>;
}
