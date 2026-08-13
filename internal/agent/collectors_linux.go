//go:build linux

package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type LinuxHostCollector struct {
	lastTotal uint64
	lastIdle  uint64
	hasCPU    bool
}

func NewLinuxHostCollector() *LinuxHostCollector {
	return &LinuxHostCollector{}
}

func (collector *LinuxHostCollector) Name() string { return "linux_host" }

func (collector *LinuxHostCollector) Collect(_ context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	samples := make([]telemetry.Sample, 0, 64)
	cpuSamples, err := collector.collectCPU(observedAt)
	if err != nil {
		return nil, err
	}
	samples = append(samples, cpuSamples...)
	memorySamples, err := collectMemory(observedAt)
	if err != nil {
		return nil, err
	}
	samples = append(samples, memorySamples...)
	samples = append(samples, collectTemperatures(observedAt)...)
	samples = append(samples, collectMountCapacity(observedAt)...)
	return samples, nil
}

func (collector *LinuxHostCollector) collectCPU(observedAt time.Time) ([]telemetry.Sample, error) {
	contents, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("read /proc/stat: %w", err)
	}
	fields := strings.Fields(strings.SplitN(string(contents), "\n", 2)[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return nil, fmt.Errorf("parse aggregate CPU stats")
	}
	var values []uint64
	for _, raw := range fields[1:] {
		value, parseErr := strconv.ParseUint(raw, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("parse CPU counter: %w", parseErr)
		}
		values = append(values, value)
	}
	var total uint64
	for _, value := range values {
		total += value
	}
	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}

	if !collector.hasCPU {
		collector.lastTotal, collector.lastIdle, collector.hasCPU = total, idle, true
		return []telemetry.Sample{unavailableSample("cpu-aggregate", "cpu.utilization", "percent", "procfs", "aggregate host CPU utilization requires two observations", observedAt)}, nil
	}
	totalDelta := total - collector.lastTotal
	idleDelta := idle - collector.lastIdle
	collector.lastTotal, collector.lastIdle = total, idle
	if totalDelta == 0 || idleDelta > totalDelta {
		return []telemetry.Sample{unavailableSample("cpu-aggregate", "cpu.utilization", "percent", "procfs", "aggregate host CPU utilization delta unavailable", observedAt)}, nil
	}
	utilization := (1 - float64(idleDelta)/float64(totalDelta)) * 100
	return []telemetry.Sample{numericSample("cpu-aggregate", "cpu.utilization", "percent", utilization, "procfs", "aggregate host CPU utilization", observedAt)}, nil
}

func collectMemory(observedAt time.Time) ([]telemetry.Sample, error) {
	contents, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("read /proc/meminfo: %w", err)
	}
	values := map[string]float64{}
	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		value, parseErr := strconv.ParseFloat(parts[1], 64)
		if parseErr != nil {
			continue
		}
		values[strings.TrimSuffix(parts[0], ":")] = value * 1024
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan /proc/meminfo: %w", err)
	}
	samples := []telemetry.Sample{}
	for _, definition := range []struct {
		name      string
		semantics string
		field     string
	}{
		{"memory.total_bytes", "OS-reported total system memory", "MemTotal"},
		{"memory.available_bytes", "OS MemAvailable; UMA panel source value", "MemAvailable"},
		{"memory.swap_free_bytes", "OS SwapFree; UMA panel source value", "SwapFree"},
		{"memory.hugepages_free_bytes", "OS huge-page free capacity; UMA panel source value", "HugePages_Free"},
	} {
		value, exists := values[definition.field]
		if !exists {
			samples = append(samples, unavailableSample("memory-host", definition.name, "bytes", "procfs", definition.semantics, observedAt))
			continue
		}
		if definition.field == "HugePages_Free" {
			pageSize := values["Hugepagesize"]
			if pageSize > 0 {
				value = value / 1024 * pageSize
			}
		}
		samples = append(samples, numericSample("memory-host", definition.name, "bytes", value, "procfs", definition.semantics, observedAt))
	}
	return samples, nil
}

func collectTemperatures(observedAt time.Time) []telemetry.Sample {
	paths, err := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	if err != nil || len(paths) == 0 {
		return []telemetry.Sample{unavailableSample("thermal-host", "temperature.celsius", "celsius", "sysfs", "no readable Linux thermal zone", observedAt)}
	}
	samples := make([]telemetry.Sample, 0, len(paths))
	for _, path := range paths {
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		milliCelsius, parseErr := strconv.ParseFloat(strings.TrimSpace(string(contents)), 64)
		if parseErr != nil {
			continue
		}
		deviceID := strings.TrimSuffix(filepath.Base(filepath.Dir(path)), "")
		samples = append(samples, numericSample(deviceID, "temperature.celsius", "celsius", milliCelsius/1000, "sysfs", "Linux thermal zone temperature", observedAt))
	}
	if len(samples) == 0 {
		return []telemetry.Sample{unavailableSample("thermal-host", "temperature.celsius", "celsius", "sysfs", "thermal zone values could not be parsed", observedAt)}
	}
	return samples
}

func collectMountCapacity(observedAt time.Time) []telemetry.Sample {
	contents, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return []telemetry.Sample{unavailableSample("storage-host", "storage.available_bytes", "bytes", "procfs", "mount table unavailable", observedAt)}
	}
	excludedTypes := map[string]bool{"proc": true, "sysfs": true, "tmpfs": true, "devtmpfs": true, "cgroup": true, "cgroup2": true, "overlay": true, "squashfs": true}
	seen := map[string]bool{}
	type mount struct{ path, filesystem string }
	mounts := []mount{}
	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || excludedTypes[fields[2]] || seen[fields[1]] {
			continue
		}
		seen[fields[1]] = true
		mounts = append(mounts, mount{path: fields[1], filesystem: fields[2]})
	}
	sort.Slice(mounts, func(i, j int) bool { return mounts[i].path < mounts[j].path })
	samples := make([]telemetry.Sample, 0, len(mounts)*2)
	for _, mount := range mounts {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount.path, &stat); err != nil {
			continue
		}
		available := float64(stat.Bavail) * float64(stat.Bsize)
		total := float64(stat.Blocks) * float64(stat.Bsize)
		deviceID := "mount:" + mount.path
		samples = append(samples,
			numericSample(deviceID, "storage.available_bytes", "bytes", available, "statfs", "filesystem available space for "+mount.path, observedAt),
			numericSample(deviceID, "storage.total_bytes", "bytes", total, "statfs", "filesystem total space for "+mount.path, observedAt),
		)
	}
	return samples
}

func numericSample(deviceID, name, unit string, value float64, source, semantics string, observedAt time.Time) telemetry.Sample {
	return telemetry.Sample{DeviceID: deviceID, Metric: domain.MetricValue{Name: name, Unit: unit, Value: &value, Quality: domain.QualityFresh, Source: source, Semantics: semantics, ObservedAt: observedAt}}
}

func experimentalNumericSample(deviceID, name, unit string, value float64, source, semantics string, observedAt time.Time) telemetry.Sample {
	return telemetry.Sample{DeviceID: deviceID, Metric: domain.MetricValue{Name: name, Unit: unit, Value: &value, Quality: domain.QualityExperimental, Source: source, Semantics: semantics, ObservedAt: observedAt}}
}

func unavailableSample(deviceID, name, unit, source, semantics string, observedAt time.Time) telemetry.Sample {
	return telemetry.Sample{DeviceID: deviceID, Metric: domain.MetricValue{Name: name, Unit: unit, Quality: domain.QualityUnavailable, Source: source, Semantics: semantics, ObservedAt: observedAt}}
}
