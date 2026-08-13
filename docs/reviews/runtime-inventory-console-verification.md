# Runtime Inventory Console Verification

**Verified:** 2026-08-13

The desktop host-detail pages for the Framework and Asus preview hosts rendered successfully after the runtime inventory addition. The existing hardware, provenance, badge, navigation, and desktop grid layouts remained intact at a 1440-pixel viewport.

The runtime inventory presentation is covered by the focused `runtimeInventoryDisplay` regression test. That test passes an endpoint-location canary into the presentation helper and verifies that its browser-facing output contains only runtime kind, approval state, and health state. The host-detail panel consumes that helper and describes endpoint locations as intentionally withheld.
