//go:build windows

package agent

import (
	"context"
	"runtime"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// WindowsBaselineCollector is intentionally narrow. It emits the portable
// runtime logical CPU count and explicit unavailable samples for telemetry that
// has not been qualified on supported Windows hardware and drivers.
type WindowsBaselineCollector struct {
	logicalCPUCount func() int
}

func NewWindowsBaselineCollector() *WindowsBaselineCollector {
	return &WindowsBaselineCollector{logicalCPUCount: runtime.NumCPU}
}

func (collector *WindowsBaselineCollector) Name() string { return "windows_baseline" }

func (collector *WindowsBaselineCollector) Collect(_ context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	logicalCPUCount := collector.logicalCPUCount()
	samples := make([]telemetry.Sample, 0, 13)
	if logicalCPUCount > 0 {
		value := float64(logicalCPUCount)
		samples = append(samples, telemetry.Sample{
			DeviceID: "cpu-host",
			Metric: domain.MetricValue{
				Name:       "cpu.logical_count",
				Unit:       "count",
				Value:      &value,
				Quality:    domain.QualityFresh,
				Source:     "go-runtime",
				Semantics:  "logical CPU count reported by the Go runtime on Windows",
				ObservedAt: observedAt,
			},
		})
	} else {
		samples = append(samples, windowsUnavailableSample("cpu-host", "cpu.logical_count", "count", "Go runtime did not report a logical CPU count.", observedAt))
	}

	for _, definition := range []struct {
		deviceID  string
		name      string
		unit      string
		semantics string
	}{
		{"cpu-host", "cpu.utilization", "percent", "CPU utilization is not collected by the unqualified Windows baseline."},
		{"memory-host", "memory.total_bytes", "bytes", "Windows total memory is not collected by the unqualified Windows baseline."},
		{"memory-host", "memory.available_bytes", "bytes", "Windows available memory is not collected by the unqualified Windows baseline."},
		{"storage-host", "storage.available_bytes", "bytes", "Windows mount capacity is not collected by the unqualified Windows baseline."},
		{"storage-host", "storage.total_bytes", "bytes", "Windows mount capacity is not collected by the unqualified Windows baseline."},
		{"thermal-host", "temperature.celsius", "celsius", "Windows thermal telemetry is not collected by the unqualified Windows baseline."},
		{"gpu-host", "gpu.utilization", "percent", "Windows GPU utilization is not collected by the unqualified Windows baseline."},
		{"gpu-host", "gpu.memory_total_bytes", "bytes", "Dedicated VRAM is not inferred by the unqualified Windows baseline."},
		{"gpu-host", "gpu.memory_used_bytes", "bytes", "Dedicated VRAM use is not inferred by the unqualified Windows baseline."},
		{"npu-host", "npu.utilization", "percent", "Windows NPU telemetry is not collected by the unqualified Windows baseline."},
		{"process-host", "process.selected_availability", "state", "Selected process telemetry is not collected by the unqualified Windows baseline."},
		{"container-host", "container.inventory", "state", "Docker Desktop and container inventory are not collected by the unqualified Windows baseline."},
	} {
		samples = append(samples, windowsUnavailableSample(definition.deviceID, definition.name, definition.unit, definition.semantics, observedAt))
	}
	return samples, nil
}

func windowsUnavailableSample(deviceID, name, unit, semantics string, observedAt time.Time) telemetry.Sample {
	return telemetry.Sample{DeviceID: deviceID, Metric: domain.MetricValue{
		Name:       name,
		Unit:       unit,
		Quality:    domain.QualityUnavailable,
		Source:     "windows-baseline",
		Semantics:  semantics,
		ObservedAt: observedAt,
	}}
}
