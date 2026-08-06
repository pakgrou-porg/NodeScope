//go:build windows

package agent

import (
	"context"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func TestWindowsBaselineCollectorReportsOnlyQualifiedCPUCount(t *testing.T) {
	observedAt := time.Date(2026, 8, 5, 14, 0, 0, 0, time.UTC)
	collector := &WindowsBaselineCollector{logicalCPUCount: func() int { return 16 }}
	samples, err := collector.Collect(context.Background(), observedAt)
	if err != nil {
		t.Fatalf("collect Windows baseline: %v", err)
	}
	byName := map[string]domain.MetricValue{}
	for _, sample := range samples {
		if err := sample.Metric.Validate(); err != nil {
			t.Fatalf("invalid metric %#v: %v", sample.Metric, err)
		}
		byName[sample.Metric.Name] = sample.Metric
	}
	logicalCPU, exists := byName["cpu.logical_count"]
	if !exists || logicalCPU.Quality != domain.QualityFresh || logicalCPU.Value == nil || *logicalCPU.Value != 16 {
		t.Fatalf("logical CPU count must be the only positive Windows baseline evidence: %#v", logicalCPU)
	}
	for _, name := range []string{
		"cpu.utilization",
		"memory.total_bytes",
		"memory.available_bytes",
		"storage.available_bytes",
		"storage.total_bytes",
		"temperature.celsius",
		"gpu.utilization",
		"gpu.memory_total_bytes",
		"gpu.memory_used_bytes",
		"npu.utilization",
		"process.selected_availability",
		"container.inventory",
	} {
		metric, exists := byName[name]
		if !exists {
			t.Fatalf("missing explicit Windows baseline metric %q", name)
		}
		if metric.Quality != domain.QualityUnavailable || metric.Value != nil {
			t.Fatalf("%s must remain explicitly unavailable, got %#v", name, metric)
		}
	}
}

func TestWindowsPreflightDoesNotExposeLinuxCapabilityProbes(t *testing.T) {
	report := InspectPreflight(func() time.Time { return time.Date(2026, 8, 5, 14, 0, 0, 0, time.UTC) })
	if report.OperatingSystem != "windows" || report.Architecture == "" {
		t.Fatalf("unexpected Windows preflight report %#v", report)
	}
	capabilities := map[string]Capability{}
	for _, capability := range report.Capabilities {
		capabilities[capability.ID] = capability
	}
	if capabilities["windows_agent_baseline"].State != CapabilityAvailable {
		t.Fatalf("missing Windows baseline capability %#v", capabilities["windows_agent_baseline"])
	}
	for _, id := range []string{"cpu_utilization", "memory", "storage", "temperature", "gpu_and_vram", "npu", "selected_processes", "container_inventory"} {
		if capabilities[id].State != CapabilityUnavailable {
			t.Fatalf("%s must remain unavailable in the Windows baseline, got %#v", id, capabilities[id])
		}
	}
	for _, forbidden := range []string{"amd_smi", "xrt_smi", "nvidia_smi", "procfs", "docker_socket"} {
		if _, exists := capabilities[forbidden]; exists {
			t.Fatalf("Windows preflight must not expose Linux probe %q", forbidden)
		}
	}
}
