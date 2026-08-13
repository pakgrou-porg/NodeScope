//go:build linux

package agent

import (
	"context"
	"os/exec"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// XDNACollector intentionally records NPU availability only in Release 1. It
// does not persist xrt-smi text or JSON reports, which can contain unrelated
// host details. NPU utilization remains explicit unavailable until a stable,
// vendor-supported structured metric interface is configured for the host.
type XDNACollector struct{}

func NewXDNACollector() *XDNACollector        { return &XDNACollector{} }
func (collector *XDNACollector) Name() string { return "amd_xdna" }

func (collector *XDNACollector) Collect(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	const source = "xrt-smi-experimental"
	const qualification = "experimental Fedora AMD XDNA telemetry; not qualified until NodeScope publishes an exact Fedora, kernel, firmware, ROCm, XRT, and XDNA matrix"
	binary, err := exec.LookPath("xrt-smi")
	if err != nil {
		return []telemetry.Sample{
			unavailableSample("npu-0", "npu.availability", "state", source, qualification+"; xrt-smi is not installed or not on PATH", observedAt),
			unavailableSample("npu-0", "npu.utilization", "percent", source, qualification+"; NPU utilization is unavailable without a supported structured collector", observedAt),
		}, nil
	}
	probe, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(probe, binary, "examine", "--batch").Run(); err != nil {
		return []telemetry.Sample{
			unavailableSample("npu-0", "npu.availability", "state", source, qualification+"; xrt-smi did not complete a readable NPU examination", observedAt),
			unavailableSample("npu-0", "npu.utilization", "percent", source, qualification+"; NPU utilization is unavailable without a supported structured collector", observedAt),
		}, nil
	}
	return []telemetry.Sample{
		numericSample("npu-0", "npu.availability", "state", 1, source, qualification+"; xrt-smi reports an accessible AMD XDNA NPU", observedAt),
		unavailableSample("npu-0", "npu.utilization", "percent", source, qualification+"; NPU utilization is unavailable until a structured xrt-smi metric contract is configured", observedAt),
	}, nil
}
