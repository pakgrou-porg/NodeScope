import DashboardLayout from "@/components/DashboardLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { trpc } from "@/lib/trpc";
import { ArrowRight, CircleDot, Search, Server, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useLocation } from "wouter";

const statusStyles = {
  healthy: "bg-emerald-300",
  degraded: "bg-amber-300",
  unavailable: "bg-rose-400",
};

const availabilityFilters = ["all", "healthy", "degraded", "unavailable"] as const;
type AvailabilityFilter = (typeof availabilityFilters)[number];

export default function HostsPage({ preview = false }: { preview?: boolean }) {
  const [, navigate] = useLocation();
  const [search, setSearch] = useState("");
  const [availability, setAvailability] = useState<AvailabilityFilter>("all");
  const previewQuery = trpc.nodescope.fleet.preview.useQuery(undefined, { enabled: preview, refetchInterval: 5_000 });
  const liveQuery = trpc.nodescope.fleet.overview.useQuery(undefined, { enabled: !preview, refetchInterval: 5_000 });
  const query = preview ? previewQuery : liveQuery;
  const targetPath = (path: string) => preview ? `/preview${path}` : path;
  const hosts = query.data?.hosts ?? [];
  const visibleHosts = useMemo(() => {
    const normalizedSearch = search.trim().toLocaleLowerCase();
    return hosts.filter((host) => {
      const matchesAvailability = availability === "all" || host.status === availability;
      const searchableText = [host.name, host.platform, host.architecture, host.role, ...host.tags].join(" ").toLocaleLowerCase();
      return matchesAvailability && (!normalizedSearch || searchableText.includes(normalizedSearch));
    });
  }, [availability, hosts, search]);

  const clearFilters = () => {
    setSearch("");
    setAvailability("all");
  };
  const filtersActive = search.trim().length > 0 || availability !== "all";

  if (query.isLoading || !query.data) {
    return <DashboardLayout><div className="space-y-4 p-8"><Skeleton className="h-14 w-full bg-white/5" /><Skeleton className="h-40 w-full bg-white/5" /><Skeleton className="h-40 w-full bg-white/5" /></div></DashboardLayout>;
  }
  if (query.isError) {
    return <DashboardLayout><div className="p-8 text-rose-200">Host inventory could not be loaded. The console has retained no substitute values.</div></DashboardLayout>;
  }

  return (
    <DashboardLayout>
      <div className="mx-auto max-w-[1320px] p-6 xl:p-8">
        <section className="flex flex-col gap-4 border-b border-white/8 pb-6 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs font-medium tracking-[0.14em] text-cyan-200">HOST DIRECTORY</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-[-0.035em] text-slate-50">Select a compute host</h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">Choose a host to inspect its current evidence, hardware, storage, runtime inventory, alerts, and recent history.</p>
          </div>
          <Badge className="w-fit border-white/10 bg-white/[0.04] text-slate-300">{query.data.hosts.length} configured hosts</Badge>
        </section>

        <section aria-label="Host directory controls" className="mt-6 rounded-2xl border border-white/8 bg-white/[0.025] p-4">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <label className="relative block min-w-0 flex-1 xl:max-w-md">
              <span className="sr-only">Search hosts</span>
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
              <input
                aria-label="Search hosts"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Search by host, platform, architecture, role, or tag"
                className="h-10 w-full rounded-xl border border-white/10 bg-[#08141e] pl-10 pr-10 text-sm text-slate-100 placeholder:text-slate-600 outline-none transition focus:border-cyan-300/50 focus:ring-2 focus:ring-cyan-300/20"
              />
              {search && (
                <button
                  type="button"
                  onClick={() => setSearch("")}
                  className="absolute right-2 top-1/2 grid h-7 w-7 -translate-y-1/2 place-items-center rounded-lg text-slate-500 transition hover:bg-white/[0.07] hover:text-slate-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300"
                  aria-label="Clear host search"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              )}
            </label>
            <div className="flex flex-wrap items-center gap-2" aria-label="Filter hosts by availability">
              {availabilityFilters.map((filter) => (
                <button
                  type="button"
                  key={filter}
                  onClick={() => setAvailability(filter)}
                  aria-pressed={availability === filter}
                  className={cn(
                    "rounded-lg border px-3 py-2 text-xs font-medium capitalize transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300",
                    availability === filter ? "border-cyan-300/30 bg-cyan-300/10 text-cyan-100" : "border-white/8 bg-white/[0.025] text-slate-400 hover:bg-white/[0.06] hover:text-slate-200",
                  )}
                >
                  {filter}
                </button>
              ))}
              {filtersActive && (
                <button type="button" onClick={clearFilters} className="px-2 text-xs font-medium text-cyan-200 transition hover:text-cyan-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300">Clear filters</button>
              )}
            </div>
          </div>
          <p aria-live="polite" className="mt-3 text-xs text-slate-500">Showing {visibleHosts.length} of {query.data.hosts.length} configured hosts.</p>
        </section>

        <section aria-label="Available hosts" className="mt-4 grid gap-4 lg:grid-cols-2">
          {visibleHosts.map((host) => (
            <button
              key={host.id}
              onClick={() => navigate(targetPath(`/hosts/${host.id}`))}
              className="group rounded-2xl border border-white/8 bg-[#0b1924] p-5 text-left shadow-[0_18px_55px_rgba(0,0,0,.18)] transition hover:-translate-y-0.5 hover:border-cyan-200/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex min-w-0 items-start gap-3">
                  <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-cyan-300/10 text-cyan-200"><Server className="h-4 w-4" /></span>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2"><h2 className="font-semibold text-slate-100">{host.name}</h2><Badge className="border-white/10 bg-white/[0.04] text-[10px] text-slate-300">{host.role === "preferred" ? "PRIMARY" : "SECONDARY"}</Badge></div>
                    <p className="mt-1 text-xs text-slate-500">{host.platform} · {host.architecture}</p>
                  </div>
                </div>
                <ArrowRight className="mt-1 h-4 w-4 shrink-0 text-slate-500 transition group-hover:translate-x-0.5 group-hover:text-cyan-200" />
              </div>
              <div className="mt-5 flex flex-wrap items-center gap-x-5 gap-y-2 border-t border-white/8 pt-4 text-xs text-slate-400">
                <span className="inline-flex items-center gap-2"><span className={cn("h-2 w-2 rounded-full", statusStyles[host.status])} />{host.status}</span>
                <span className="inline-flex items-center gap-2"><CircleDot className="h-3.5 w-3.5 text-slate-500" />{host.freshness.state === "fresh" ? `${host.freshness.ageSeconds}s since evidence` : host.freshness.state}</span>
                <span>{host.tags.slice(0, 3).join(" · ") || "No host tags"}</span>
              </div>
            </button>
          ))}
        </section>

        {visibleHosts.length === 0 && (
          <section className="mt-4 rounded-2xl border border-dashed border-white/10 bg-white/[0.025] p-8 text-center">
            <p className="text-sm font-medium text-slate-200">No hosts match the current directory filters.</p>
            <p className="mt-2 text-xs leading-5 text-slate-500">NodeScope has not hidden or substituted host evidence; adjust or clear the filters to return to the full configured host list.</p>
            <Button variant="outline" onClick={clearFilters} className="mt-5 border-white/10 text-slate-300 hover:bg-white/[0.06]">Clear directory filters</Button>
          </section>
        )}

        <div className="mt-8 rounded-2xl border border-white/8 bg-white/[0.025] p-5">
          <p className="text-sm font-medium text-slate-200">Need a different operational view?</p>
          <div className="mt-4 flex flex-wrap gap-3">
            <Button variant="outline" onClick={() => navigate(targetPath("/alerts"))} className="border-white/10 text-slate-300 hover:bg-white/[0.06]">Open alert queue</Button>
            <Button variant="outline" onClick={() => navigate(targetPath("/operations"))} className="border-white/10 text-slate-300 hover:bg-white/[0.06]">Open operations</Button>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
