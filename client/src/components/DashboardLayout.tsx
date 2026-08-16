import { useAuth } from "@/_core/hooks/useAuth";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { startLogin } from "@/const";
import { useIsMobile } from "@/hooks/useMobile";
import {
  Activity,
  BellRing,
  ChevronRight,
  DatabaseBackup,
  Gauge,
  LayoutDashboard,
  LogOut,
  PanelLeft,
  ServerCog,
  Settings2,
  ShieldCheck,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useLocation } from "wouter";
import { DashboardLayoutSkeleton } from "./DashboardLayoutSkeleton";

const navigation = [
  { icon: LayoutDashboard, label: "Fleet overview", path: "/" },
  { icon: Gauge, label: "Hosts", path: "/hosts" },
  { icon: BellRing, label: "Alerts", path: "/alerts" },
  { icon: ServerCog, label: "Operations", path: "/operations" },
  { icon: Settings2, label: "Administration", path: "/administration" },
];

const previewUser = { name: "Development preview", email: "fixture-only", role: "Administrator" };

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { loading, user } = useAuth();
  const preview = import.meta.env.DEV && /^\/preview(?:\/|$)/.test(window.location.pathname);

  if (loading && !preview) return <DashboardLayoutSkeleton />;
  if (!user && !preview) {
    return (
      <div className="min-h-screen bg-[#071018] text-[#edf5f7] grid place-items-center p-6">
        <section className="w-full max-w-md rounded-2xl border border-white/10 bg-[#0c1922] p-8 shadow-2xl shadow-black/30">
          <div className="mb-8 flex items-center gap-3">
            <div className="grid h-10 w-10 place-items-center rounded-xl bg-cyan-300 text-[#071018] shadow-lg shadow-cyan-300/20">
              <Activity className="h-5 w-5" />
            </div>
            <div>
              <p className="text-sm font-semibold tracking-[0.16em] text-cyan-200">NODESCOPE</p>
              <p className="text-xs text-slate-400">Fleet observability and operations</p>
            </div>
          </div>
          <h1 className="text-2xl font-semibold tracking-tight">Authorized access required</h1>
          <p className="mt-3 text-sm leading-6 text-slate-400">
            NodeScope protects fleet health, operational controls, and audit data behind authenticated, role-aware access.
          </p>
          <Button onClick={() => startLogin()} className="mt-8 w-full bg-cyan-300 text-[#071018] hover:bg-cyan-200">
            Continue to sign in
          </Button>
        </section>
      </div>
    );
  }

  return (
    <SidebarProvider>
      <NodeScopeSidebar preview={preview} user={user ? { name: user.name ?? "Operator", email: user.email ?? "", role: user.role } : previewUser}>
        {children}
      </NodeScopeSidebar>
    </SidebarProvider>
  );
}

function NodeScopeSidebar({ children, preview, user }: { children: React.ReactNode; preview: boolean; user: { name: string; email: string; role: string } }) {
  const [, navigate] = useLocation();
  const { state, toggleSidebar } = useSidebar();
  const isMobile = useIsMobile();
  const { logout } = useAuth();
  const [location] = useLocation();
  const collapsed = state === "collapsed";
  const targetPath = (path: string) => preview ? `/preview${path === "/" ? "" : path}` : path;
  const normalizedLocation = preview ? location.replace(/^\/preview(?=\/|$)/, "") || "/" : location;

  useEffect(() => {
    document.documentElement.classList.add("dark");
    return () => document.documentElement.classList.remove("dark");
  }, []);

  return (
    <>
      <Sidebar collapsible="icon" className="border-r border-white/8 bg-[#08141e] text-slate-200">
        <SidebarHeader className="h-[76px] border-b border-white/8 px-3">
          <div className="flex h-full items-center gap-3">
            <button
              onClick={toggleSidebar}
              className="grid h-9 w-9 shrink-0 place-items-center rounded-xl border border-white/10 bg-white/[0.035] text-slate-300 transition-colors hover:bg-white/[0.08] focus-visible:ring-2 focus-visible:ring-cyan-300"
              aria-label="Toggle navigation"
            >
              <PanelLeft className="h-4 w-4" />
            </button>
            {!collapsed && (
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-semibold tracking-[0.17em] text-cyan-200">NODESCOPE</span>
                  <span className="h-1.5 w-1.5 rounded-full bg-emerald-300" aria-label="System operational" />
                </div>
                <p className="mt-0.5 truncate text-[11px] text-slate-500">Fleet observability</p>
              </div>
            )}
          </div>
        </SidebarHeader>

        <SidebarContent className="px-2 py-4">
          {!collapsed && <p className="px-3 pb-2 text-[10px] font-medium tracking-[0.16em] text-slate-500">MONITOR</p>}
          <SidebarMenu>
            {navigation.map((item) => {
              const active = item.path === "/" ? normalizedLocation === "/" : normalizedLocation.startsWith(item.path.split("/").slice(0, 2).join("/"));
              return (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton
                    isActive={active}
                    onClick={() => navigate(targetPath(item.path))}
                    tooltip={item.label}
                    className="h-10 rounded-lg text-slate-400 transition-colors hover:bg-white/[0.07] hover:text-slate-100 data-[active=true]:bg-cyan-300/10 data-[active=true]:text-cyan-100"
                  >
                    <item.icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })}
          </SidebarMenu>

          {!collapsed && (
            <div className="mx-2 mt-8 rounded-xl border border-white/8 bg-white/[0.035] p-3">
              <div className="flex items-center gap-2 text-xs font-medium text-slate-300">
                <DatabaseBackup className="h-3.5 w-3.5 text-cyan-300" />
                Backup lease
              </div>
              <p className="mt-2 text-xs leading-5 text-slate-500">Framework holds the current lease. Secondary takeover is monitored continuously.</p>
            </div>
          )}
        </SidebarContent>

        <SidebarFooter className="border-t border-white/8 p-3">
          {preview && !collapsed && <Badge className="mb-3 w-full justify-center border-0 bg-amber-300/10 text-[10px] font-medium tracking-[0.12em] text-amber-200">DEVELOPMENT PREVIEW</Badge>}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="flex w-full items-center gap-3 rounded-xl p-1.5 text-left transition-colors hover:bg-white/[0.06] focus-visible:ring-2 focus-visible:ring-cyan-300">
                <Avatar className="h-8 w-8 border border-white/10 bg-cyan-300/10">
                  <AvatarFallback className="bg-transparent text-xs font-semibold text-cyan-100">{user.name.charAt(0).toUpperCase()}</AvatarFallback>
                </Avatar>
                {!collapsed && (
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-xs font-medium text-slate-200">{user.name}</p>
                    <p className="mt-0.5 truncate text-[11px] text-slate-500">{user.role}</p>
                  </div>
                )}
                {!collapsed && <ChevronRight className="h-3.5 w-3.5 text-slate-600" />}
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56 border-white/10 bg-[#0c1922] text-slate-200">
              <div className="px-2 py-2">
                <p className="text-xs font-medium">{user.name}</p>
                <p className="mt-1 text-[11px] text-slate-500">{user.email || "Role-controlled session"}</p>
              </div>
              <DropdownMenuSeparator className="bg-white/10" />
              {!preview && (
                <DropdownMenuItem onClick={logout} className="cursor-pointer text-rose-300 focus:bg-rose-400/10 focus:text-rose-200">
                  <LogOut className="mr-2 h-4 w-4" /> Sign out
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarFooter>
      </Sidebar>

      <SidebarInset className="nodescope-grid min-h-screen bg-[#071018] text-slate-100">
        {isMobile && (
          <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-white/8 bg-[#08141e]/95 px-4 backdrop-blur-xl">
            <SidebarTrigger className="rounded-lg border border-white/10 bg-white/[0.04] text-slate-200" />
            <span className="text-xs font-semibold tracking-[0.14em] text-cyan-100">NODESCOPE</span>
          </header>
        )}
        <main className="nodescope-signal-line min-h-screen">{children}</main>
      </SidebarInset>
    </>
  );
}
