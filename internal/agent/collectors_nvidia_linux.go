//go:build linux

package agent

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// NvidiaCollector uses nvidia-smi only for values it explicitly reports. It
// intentionally treats unavailable framebuffer memory as unavailable rather
// than deriving VRAM from unified system memory.
type NvidiaCollector struct{}

func NewNvidiaCollector() *NvidiaCollector      { return &NvidiaCollector{} }
func (collector *NvidiaCollector) Name() string { return "nvidia_smi" }

func (collector *NvidiaCollector) Collect(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	binary, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return []telemetry.Sample{unavailableSample("gpu-nvidia", "gpu.availability", "state", "nvidia-smi", "nvidia-smi is unavailable; no NVIDIA GPU reading is inferred", observedAt)}, nil
	}
	command := exec.CommandContext(ctx, binary, "--query-gpu=index,name,utilization.gpu,temperature.gpu,memory.total,memory.used", "--format=csv,noheader,nounits")
	output, err := command.Output()
	if err != nil {
		return []telemetry.Sample{unavailableSample("gpu-nvidia", "gpu.availability", "state", "nvidia-smi", "nvidia-smi did not return a readable GPU report", observedAt)}, nil
	}
	samples := []telemetry.Sample{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := splitCSVLine(line)
		if len(fields) < 6 {
			continue
		}
		deviceID := "nvidia-gpu-" + fields[0]
		samples = append(samples, numericSample(deviceID, "gpu.availability", "state", 1, "nvidia-smi", "NVIDIA GPU is reported by nvidia-smi", observedAt))
		if utilization, ok := parseReportedFloat(fields[2]); ok {
			samples = append(samples, numericSample(deviceID, "gpu.utilization", "percent", utilization, "nvidia-smi", "NVIDIA-reported GPU utilization", observedAt))
		} else {
			samples = append(samples, unavailableSample(deviceID, "gpu.utilization", "percent", "nvidia-smi", "NVIDIA GPU utilization was not reported", observedAt))
		}
		if temperature, ok := parseReportedFloat(fields[3]); ok {
			samples = append(samples, numericSample(deviceID, "temperature.celsius", "celsius", temperature, "nvidia-smi", "NVIDIA-reported GPU temperature", observedAt))
		} else {
			samples = append(samples, unavailableSample(deviceID, "temperature.celsius", "celsius", "nvidia-smi", "NVIDIA GPU temperature was not reported", observedAt))
		}
		if totalMiB, ok := parseReportedFloat(fields[4]); ok {
			samples = append(samples, numericSample(deviceID, "gpu.dedicated_vram.total_bytes", "bytes", totalMiB*1024*1024, "nvidia-smi", "nvidia-smi reported framebuffer memory; not inferred from host memory", observedAt))
		} else {
			samples = append(samples, unavailableSample(deviceID, "gpu.dedicated_vram.total_bytes", "bytes", "nvidia-smi", "framebuffer memory is unavailable on this platform; no VRAM value is synthesized", observedAt))
		}
		if usedMiB, ok := parseReportedFloat(fields[5]); ok {
			samples = append(samples, numericSample(deviceID, "gpu.dedicated_vram.used_bytes", "bytes", usedMiB*1024*1024, "nvidia-smi", "nvidia-smi reported framebuffer memory use; not inferred", observedAt))
		} else {
			samples = append(samples, unavailableSample(deviceID, "gpu.dedicated_vram.used_bytes", "bytes", "nvidia-smi", "framebuffer memory use is unavailable on this platform; no VRAM value is synthesized", observedAt))
		}
		samples = append(samples, collector.processMemory(ctx, binary, deviceID, observedAt)...)
	}
	if len(samples) == 0 {
		return []telemetry.Sample{unavailableSample("gpu-nvidia", "gpu.availability", "state", "nvidia-smi", "nvidia-smi returned no GPU inventory", observedAt)}, nil
	}
	return samples, nil
}

func (collector *NvidiaCollector) processMemory(ctx context.Context, binary, deviceID string, observedAt time.Time) []telemetry.Sample {
	command := exec.CommandContext(ctx, binary, "--query-compute-apps=pid,used_memory", "--format=csv,noheader,nounits")
	output, err := command.Output()
	if err != nil {
		return []telemetry.Sample{unavailableSample(deviceID, "gpu.uma.per_process_memory_bytes", "bytes", "nvidia-smi", "per-process GPU memory was not reported", observedAt)}
	}
	samples := []telemetry.Sample{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := splitCSVLine(line)
		if len(fields) != 2 {
			continue
		}
		if memoryMiB, ok := parseReportedFloat(fields[1]); ok {
			samples = append(samples, numericSample(deviceID+":pid-"+fields[0], "gpu.uma.per_process_memory_bytes", "bytes", memoryMiB*1024*1024, "nvidia-smi", "nvidia-smi per-process GPU memory; UMA/FB semantics are vendor-reported", observedAt))
		}
	}
	if len(samples) == 0 {
		return []telemetry.Sample{unavailableSample(deviceID, "gpu.uma.per_process_memory_bytes", "bytes", "nvidia-smi", "no per-process GPU memory value was reported", observedAt)}
	}
	return samples
}

func splitCSVLine(line string) []string {
	parts := strings.Split(line, ",")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts
}

func parseReportedFloat(raw string) (float64, bool) {
	if raw == "" || strings.EqualFold(raw, "N/A") || strings.EqualFold(raw, "[Not Supported]") {
		return 0, false
	}
	value, err := strconv.ParseFloat(raw, 64)
	return value, err == nil
}
