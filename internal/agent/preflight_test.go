package agent

import (
	"testing"
	"time"
)

func TestInspectPreflightProducesStableCoreCapabilities(t *testing.T) {
	report := InspectPreflight(func() time.Time {
		return time.Date(2026, 7, 22, 16, 30, 0, 0, time.UTC)
	})
	if report.GeneratedAt.IsZero() || report.Architecture == "" || report.OperatingSystem == "" {
		t.Fatalf("incomplete report %#v", report)
	}
	capabilities := map[string]Capability{}
	for _, capability := range report.Capabilities {
		capabilities[capability.ID] = capability
	}
	for _, id := range []string{"amd_smi", "xrt_smi", "nvidia_smi", "procfs", "docker_socket"} {
		if _, exists := capabilities[id]; !exists {
			t.Fatalf("missing capability %q", id)
		}
	}
	if capabilities["amd_smi"].State == CapabilityUnavailable && len(capabilities["amd_smi"].RemediationSteps) == 0 {
		t.Fatal("missing AMD SMI must provide remediation guidance")
	}
}
