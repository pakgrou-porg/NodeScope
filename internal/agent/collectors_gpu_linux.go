//go:build linux

package agent

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// LinuxDRMCollector uses kernel-exported DRM/sysfs values. It is deliberately
// conservative: absent dedicated-memory files produce explicit unavailable
// readings rather than a synthesized VRAM calculation from system RAM.
type LinuxDRMCollector struct{}

func NewLinuxDRMCollector() *LinuxDRMCollector    { return &LinuxDRMCollector{} }
func (collector *LinuxDRMCollector) Name() string { return "linux_drm" }

func (collector *LinuxDRMCollector) Collect(_ context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	cards, err := filepath.Glob("/sys/class/drm/card[0-9]*")
	if err != nil {
		return nil, err
	}
	samples := []telemetry.Sample{}
	for _, card := range cards {
		vendor, err := readTrimmed(filepath.Join(card, "device/vendor"))
		if err != nil {
			continue
		}
		deviceID := filepath.Base(card)
		if vendor == "0x1002" {
			samples = append(samples, collectAMDDRM(deviceID, card, observedAt)...)
		}
	}
	if len(samples) == 0 {
		return []telemetry.Sample{unavailableSample("gpu-host", "gpu.availability", "state", "sysfs", "no supported AMD DRM telemetry device discovered", observedAt)}, nil
	}
	return samples, nil
}

func collectAMDDRM(deviceID, card string, observedAt time.Time) []telemetry.Sample {
	base := filepath.Join(card, "device")
	samples := []telemetry.Sample{numericSample(deviceID, "gpu.availability", "state", 1, "sysfs", "AMD DRM device is present", observedAt)}
	if busy, ok := readFloat(filepath.Join(base, "gpu_busy_percent")); ok {
		samples = append(samples, numericSample(deviceID, "gpu.utilization", "percent", busy, "sysfs", "AMD DRM gpu_busy_percent", observedAt))
	} else {
		samples = append(samples, unavailableSample(deviceID, "gpu.utilization", "percent", "sysfs", "AMD DRM gpu_busy_percent is unavailable", observedAt))
	}

	vramTotal, totalOK := readFloat(filepath.Join(base, "mem_info_vram_total"))
	vramUsed, usedOK := readFloat(filepath.Join(base, "mem_info_vram_used"))
	if totalOK {
		samples = append(samples, numericSample(deviceID, "gpu.dedicated_vram.total_bytes", "bytes", vramTotal, "sysfs", "kernel-reported dedicated VRAM; not UMA", observedAt))
	} else {
		samples = append(samples, unavailableSample(deviceID, "gpu.dedicated_vram.total_bytes", "bytes", "sysfs", "dedicated VRAM is not exposed by this DRM device; no value is inferred from system RAM", observedAt))
	}
	if usedOK {
		samples = append(samples, numericSample(deviceID, "gpu.dedicated_vram.used_bytes", "bytes", vramUsed, "sysfs", "kernel-reported dedicated VRAM use; not UMA", observedAt))
	} else {
		samples = append(samples, unavailableSample(deviceID, "gpu.dedicated_vram.used_bytes", "bytes", "sysfs", "dedicated VRAM use is not exposed by this DRM device; no value is inferred", observedAt))
	}
	if !totalOK && !usedOK {
		samples = append(samples, unavailableSample(deviceID, "gpu.uma.per_process_memory_bytes", "bytes", "sysfs", "per-process UMA GPU memory requires a supported runtime collector", observedAt))
	}
	return samples
}

func readTrimmed(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(contents)), nil
}

func readFloat(path string) (float64, bool) {
	value, err := readTrimmed(path)
	if err != nil {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}
