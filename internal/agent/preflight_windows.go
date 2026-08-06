//go:build windows

package agent

import (
	"fmt"
	"runtime"
	"time"
)

// InspectPreflight describes only the Windows functionality supported by the
// baseline agent. It intentionally does not probe vendor tools or invoke WMI,
// PowerShell, Docker Desktop, or registry APIs until those paths have an
// explicit hardware and security qualification matrix.
func InspectPreflight(now func() time.Time) PreflightReport {
	if now == nil {
		now = time.Now
	}
	logicalCPUCount := runtime.NumCPU()
	cpuState := CapabilityUnavailable
	cpuDetail := "Windows runtime did not report a logical CPU count."
	if logicalCPUCount > 0 {
		cpuState = CapabilityAvailable
		cpuDetail = fmt.Sprintf("Windows runtime reports %d logical CPU(s).", logicalCPUCount)
	}
	unavailable := func(id, detail string) Capability {
		return Capability{
			ID:               id,
			State:            CapabilityUnavailable,
			Detail:           detail,
			Verification:     "Run nodescope-agent.exe -preflight and retain this unavailable state until NodeScope publishes a qualified Windows collector.",
			DocumentationURL: "https://github.com/pakgrou-porg/NodeScope",
		}
	}
	return PreflightReport{
		GeneratedAt:     now().UTC(),
		OperatingSystem: "windows",
		Architecture:    runtime.GOARCH,
		Capabilities: []Capability{
			{
				ID:           "windows_agent_baseline",
				State:        CapabilityAvailable,
				Detail:       "NodeScope Windows baseline supports agent transport and logical CPU-count evidence only.",
				Verification: "nodescope-agent.exe -preflight",
			},
			{ID: "logical_cpu_count", State: cpuState, Detail: cpuDetail, Verification: "nodescope-agent.exe -preflight"},
			unavailable("cpu_utilization", "CPU utilization is not collected by the unqualified Windows baseline."),
			unavailable("memory", "Windows memory telemetry is not collected by the unqualified Windows baseline."),
			unavailable("storage", "Windows mount and free-space telemetry is not collected by the unqualified Windows baseline."),
			unavailable("temperature", "Windows thermal telemetry is not collected by the unqualified Windows baseline."),
			unavailable("gpu_and_vram", "GPU utilization and dedicated VRAM are not inferred from Windows, NVIDIA, or LM Studio tooling by this baseline."),
			unavailable("npu", "Windows NPU telemetry is not collected by the unqualified Windows baseline."),
			unavailable("selected_processes", "Selected process telemetry is not collected by the unqualified Windows baseline."),
			unavailable("container_inventory", "Docker Desktop and container inventory are not collected by the unqualified Windows baseline."),
		},
	}
}
