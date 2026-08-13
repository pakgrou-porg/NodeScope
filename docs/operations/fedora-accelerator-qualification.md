# Fedora Accelerator Qualification Boundary

NodeScope treats AMD GPU and AMD XDNA NPU telemetry on the Framework Fedora host as **experimental** until an administrator validates and records an exact compatibility matrix. The supported Release 1 baseline remains generic Linux CPU, RAM, storage, mount, temperature, and selected-process evidence. Experimental accelerator readings are still useful diagnostic evidence, but they must not be used as production alert thresholds, capacity decisions, or a substitute for an approved hardware matrix.

## What the Agent Emits

| Evidence family | Provenance source | Interpretation |
|---|---|---|
| AMD DRM availability, utilization, and kernel-exported dedicated-memory fields | `sysfs-experimental` | Fresh kernel values when available, but explicitly experimental on Fedora. Dedicated VRAM is never inferred from system or UMA memory. |
| AMD XDNA availability and unavailable utilization | `xrt-smi-experimental` | Availability only after a bounded probe; utilization remains unavailable until a stable structured metric contract is qualified. |
| Missing accelerator tooling or unreadable device state | Same experimental provenance | Explicit unavailable evidence; the agent does not invent a value or install host packages. |

## Operator Contract

The preflight report intentionally does **not** prescribe package-install commands for `amd-smi` or `xrt-smi`. A host operator must use the vendor-supported path appropriate to the actual system and must not infer that a package name, driver, or firmware combination is supported merely because a preflight capability is absent.

Before promoting either source family to production use, record and test a matrix that includes the Fedora release, kernel, firmware, ROCm, AMD SMI, XRT, XDNA driver, and NodeScope agent revision. The validation must confirm the measured fields, their units, their meaning on UMA systems, unavailable behavior, and alert-safe thresholds. Until then, consumers should display the source and experimental semantics alongside every accelerator reading.
