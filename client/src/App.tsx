import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import NotFound from "@/pages/NotFound";
import { Route, Switch } from "wouter";
import ErrorBoundary from "./components/ErrorBoundary";
import { ThemeProvider } from "./contexts/ThemeContext";
import Home from "./pages/Home";
import FleetOverview from "./pages/FleetOverview";
import HostDetail from "./pages/HostDetail";
import AlertsPage from "./pages/AlertsPage";
import OperationsPage from "./pages/OperationsPage";
import AdministrationPage from "./pages/AdministrationPage";
import AlertRulesPage from "./pages/AlertRulesPage";

function Router() {
  // make sure to consider if you need authentication for certain routes
  return (
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
      <Route path={"/"} component={Home} />
      <Route path={"/404"} component={NotFound} />
      {/* Final fallback route */}
      <Route component={NotFound} />
    </Switch>
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
