# Host Detail Enhancement — Browser Check

**Environment:** local development preview at `/preview/hosts/framework`. **Protected-environment activity:** none. **Host or alert mutation:** none.

The Framework host-detail page rendered both **Fleet overview** and **All hosts** navigation controls, the active-alert, non-fresh-evidence, and preflight-attention summaries, and the labeled host-detail tab row. Experimental Radeon/XDNA readings and the unavailable NPU-throughput reading remained explicitly labeled with their original quality and source; the summary did not replace, estimate, or suppress them.

Selecting **All hosts** returned to `/preview/hosts`, where the full configured host directory remained visible. Navigation was browser-only and did not change host configuration, metric evidence, alert state, credentials, schedules, or deployment state.
