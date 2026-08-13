import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { lazy, Suspense } from "react";
import { Route, Switch } from "wouter";
import ErrorBoundary from "./components/ErrorBoundary";
import { ThemeProvider } from "./contexts/ThemeContext";

const FleetOverview = lazy(() => import("./pages/FleetOverview"));
const HostDetail = lazy(() => import("./pages/HostDetail"));
const AlertsPage = lazy(() => import("./pages/AlertsPage"));
const OperationsPage = lazy(() => import("./pages/OperationsPage"));
const AdministrationPage = lazy(() => import("./pages/AdministrationPage"));
const AlertRulesPage = lazy(() => import("./pages/AlertRulesPage"));
const NotFound = lazy(() => import("@/pages/NotFound"));

function ConsoleRouteLoading() {
  return <main aria-busy="true" aria-live="polite" className="grid min-h-screen place-items-center bg-[#07131d] px-6 text-center"><div><p className="text-sm font-medium text-slate-200">Loading NodeScope console…</p><p className="mt-2 text-xs text-slate-500">Preparing the requested fleet view.</p></div></main>;
}

function Router() {
  // make sure to consider if you need authentication for certain routes
  return (
    <Suspense fallback={<ConsoleRouteLoading />}>
      <Switch>
        <Route path={"/preview/administration/alert-rules"} component={() => <AlertRulesPage preview />} />
        <Route path={"/preview/administration"} component={() => <AdministrationPage preview />} />
        <Route path={"/preview/operations"} component={() => <OperationsPage preview />} />
        <Route path={"/preview/alerts"} component={() => <AlertsPage preview />} />
        <Route path={"/preview/hosts/:hostId"} component={({ params }) => <HostDetail hostId={params.hostId} preview />} />
        <Route path={"/preview"} component={() => <FleetOverview preview />} />
        <Route path={"/hosts/:hostId"} component={({ params }) => <HostDetail hostId={params.hostId} />} />
        <Route path={"/alerts"} component={() => <AlertsPage />} />
        <Route path={"/operations"} component={() => <OperationsPage />} />
        <Route path={"/administration/alert-rules"} component={() => <AlertRulesPage />} />
        <Route path={"/administration"} component={() => <AdministrationPage />} />
        <Route path={"/"} component={() => <FleetOverview />} />
        <Route path={"/404"} component={NotFound} />
        {/* Final fallback route */}
        <Route component={NotFound} />
      </Switch>
    </Suspense>
  );
}

// NOTE: About Theme
// - First choose a default theme according to your design style (dark or light bg), than change color palette in index.css
//   to keep consistent foreground/background color across components
// - If you want to make theme switchable, pass `switchable` ThemeProvider and use `useTheme` hook

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider
        defaultTheme="dark"
        // switchable
      >
        <TooltipProvider>
          <Toaster />
          <Router />
        </TooltipProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
