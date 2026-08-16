# Lightweight Console Enhancement — Browser Check

**Environment:** local development preview at `/preview`. **Protected-environment activity:** none. **Observed route:** fleet overview.

The browser preview rendered the desktop sidebar, evidence-quality badges, host cards, alert drill-down controls, and the new **Refresh** button in the fleet header. The refresh control was exposed with the accessible name `Refresh fleet telemetry`; the page announced `Fleet telemetry refresh idle` through its live region before interaction. The action is a client-side query refetch only and does not create, modify, or deploy any protected resource.

The visual check preserved existing evidence boundaries: Framework AMD GPU and XDNA NPU remained marked experimental, Asus NPU remained not supported, and no endpoint, credential, prompt, or response content was displayed.

During follow-up host-directory validation, the first rendering exposed a React hook-order error on the loading-to-data transition. The hook was moved ahead of the early loading return and focused TypeScript/Vitest checks then passed. A second browser check identified an existing development-preview routing limitation: `/preview/hosts` did not enter preview mode because the layout recognized only the exact `/preview` pathname. This is tracked for a bounded correction that will preserve production authentication outside the development-only `/preview` route prefix.

After the route correction, `/preview/hosts` rendered the development fixture, search field, four availability filter buttons, and `Showing 2 of 2 configured hosts.` result count. Entering `nonexistent-host` produced `Showing 0 of 2 configured hosts.`, the explicit empty state, and both clear-filter recovery controls. No host data was written, deleted, estimated, or substituted during this interaction.

Using the empty-state clear control restored `Showing 2 of 2 configured hosts.` and both host cards. Returning to the fleet preview confirmed the visible Refresh control, its `Refresh fleet telemetry` accessible name, and its idle live-region message remain present after the host-directory and preview-route changes.

Activating Refresh completed without a visible error or protected operation. The refreshed preview remained on the same fixture-backed fleet route, returned to the idle live-region message, and showed an updated generated-at time. The action only refetched the browser query; it did not alter metrics, hosts, alerts, schedules, roles, credentials, or deployment state.
