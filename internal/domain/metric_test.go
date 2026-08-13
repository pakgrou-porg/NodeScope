package domain

import (
	"testing"
	"time"
)

func float64Ptr(value float64) *float64 {
	return &value
}

func TestMetricValueRejectsUnavailableValue(t *testing.T) {
	value := MetricValue{
		Name:       "gpu.memory.free",
		Unit:       "bytes",
		Value:      float64Ptr(0),
		Quality:    QualityUnavailable,
		Source:     "nvidia-smi",
		Semantics:  "vendor field unavailable",
		ObservedAt: time.Now().UTC(),
	}

	if err := value.Validate(); err == nil {
		t.Fatal("expected unavailable metric with a value to be rejected")
	}
}

func TestMetricValueRequiresProvenanceAndSemantics(t *testing.T) {
	reading := MetricValue{
		Name:       "memory.available",
		Unit:       "bytes",
		Value:      float64Ptr(1024),
		Quality:    QualityFresh,
		ObservedAt: time.Now().UTC(),
	}

	if err := reading.Validate(); err == nil {
		t.Fatal("expected missing provenance and semantics to be rejected")
	}
}

func TestExperimentalMetricRequiresValueAndIsNotAlertEligible(t *testing.T) {
	reading := MetricValue{
		Name:       "gpu.utilization",
		Unit:       "percent",
		Value:      float64Ptr(61),
		Quality:    QualityExperimental,
		Source:     "sysfs-experimental",
		Semantics:  "unqualified Fedora AMD DRM evidence",
		ObservedAt: time.Now().UTC(),
	}
	if err := reading.Validate(); err != nil {
		t.Fatalf("expected value-bearing experimental metric to validate: %v", err)
	}
	if reading.Quality.EligibleForAutomaticAlerting() {
		t.Fatal("experimental metric must not be eligible for automatic alerts")
	}

	reading.Value = nil
	if err := reading.Validate(); err == nil {
		t.Fatal("expected experimental metric without a reading to be rejected")
	}
}

func TestUMAReadingDoesNotRequireDedicatedVRAM(t *testing.T) {
	reading := MemoryReading{
		MetricValue: MetricValue{
			Name:       "memory.os.available",
			Unit:       "bytes",
			Value:      float64Ptr(64 * 1024 * 1024 * 1024),
			Quality:    QualityFresh,
			Source:     "/proc/meminfo",
			Semantics:  "host OS MemAvailable under unified memory",
			ObservedAt: time.Now().UTC(),
		},
		Kind: MemoryOSAvailable,
	}

	if err := reading.Validate(); err != nil {
		t.Fatalf("expected valid UMA reading, received %v", err)
	}
}

func TestDedicatedMetricCannotClaimUnifiedSemantics(t *testing.T) {
	reading := MemoryReading{
		MetricValue: MetricValue{
			Name:       "gpu.memory.framebuffer.free",
			Unit:       "bytes",
			Value:      float64Ptr(1024),
			Quality:    QualityFresh,
			Source:     "vendor-api",
			Semantics:  "unified memory estimate",
			ObservedAt: time.Now().UTC(),
		},
		Kind: MemoryDedicatedFramebuffer,
	}

	if err := reading.Validate(); err == nil {
		t.Fatal("expected dedicated framebuffer reading with unified semantics to be rejected")
	}
}
