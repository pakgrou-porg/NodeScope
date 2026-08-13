// Package domain contains NodeScope's stable, vendor-neutral telemetry vocabulary.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// MetricQuality describes how a metric should be interpreted and rendered.
// A value must never be assumed to be fresh simply because a sample exists.
type MetricQuality string

const (
	QualityFresh       MetricQuality = "fresh"
	QualityStale       MetricQuality = "stale"
	QualityUnavailable MetricQuality = "unavailable"
	QualityUnsupported MetricQuality = "unsupported"
	QualityEstimated   MetricQuality = "estimated"
	// QualityExperimental carries a reading from an unqualified platform or
	// runtime path. It is rendered with explicit provenance and must not be
	// promoted to ordinary alert evidence until qualification is recorded.
	QualityExperimental MetricQuality = "experimental"
)

func (q MetricQuality) Valid() bool {
	switch q {
	case QualityFresh, QualityStale, QualityUnavailable, QualityUnsupported, QualityEstimated, QualityExperimental:
		return true
	default:
		return false
	}
}

// EligibleForAutomaticAlerting reports whether a quality can participate in
// automatic policy evaluation. Experimental values remain visible with their
// provenance but cannot trigger alerts until their collection path is
// explicitly qualified.
func (q MetricQuality) EligibleForAutomaticAlerting() bool {
	return q != QualityExperimental && q.Valid()
}

// MemorySemantics distinguishes dedicated-memory values from unified-memory
// values. It prevents the UI and API from collapsing GX10 UMA values into a
// fabricated generic "VRAM free" figure.
type MemorySemantics string

const (
	MemoryDedicatedFramebuffer MemorySemantics = "dedicated_framebuffer"
	MemoryOSAvailable          MemorySemantics = "os_mem_available"
	MemorySwapFree             MemorySemantics = "swap_free"
	MemoryHugePages            MemorySemantics = "huge_pages"
	MemoryRuntimeAllocatable   MemorySemantics = "runtime_allocatable"
	MemoryProcessGPU           MemorySemantics = "per_process_gpu"
)

func (m MemorySemantics) Valid() bool {
	switch m {
	case MemoryDedicatedFramebuffer, MemoryOSAvailable, MemorySwapFree, MemoryHugePages, MemoryRuntimeAllocatable, MemoryProcessGPU:
		return true
	default:
		return false
	}
}

// MetricValue is one measured or explicitly unavailable value. Value is nil
// for unavailable and unsupported readings; callers must not replace it with
// zero during storage, API serialization, or display.
type MetricValue struct {
	Name       string        `json:"name"`
	Unit       string        `json:"unit"`
	Value      *float64      `json:"value,omitempty"`
	Quality    MetricQuality `json:"quality"`
	Source     string        `json:"source"`
	Semantics  string        `json:"semantics"`
	ObservedAt time.Time     `json:"observedAt"`
}

func (m MetricValue) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("metric name is required")
	}
	if !m.Quality.Valid() {
		return fmt.Errorf("metric %q has invalid quality %q", m.Name, m.Quality)
	}
	if strings.TrimSpace(m.Source) == "" {
		return fmt.Errorf("metric %q requires a provenance source", m.Name)
	}
	if strings.TrimSpace(m.Semantics) == "" {
		return fmt.Errorf("metric %q requires semantics", m.Name)
	}
	if m.ObservedAt.IsZero() {
		return fmt.Errorf("metric %q requires an observation time", m.Name)
	}
	if (m.Quality == QualityUnavailable || m.Quality == QualityUnsupported) && m.Value != nil {
		return fmt.Errorf("metric %q cannot have a value when quality is %q", m.Name, m.Quality)
	}
	if (m.Quality == QualityFresh || m.Quality == QualityStale || m.Quality == QualityEstimated || m.Quality == QualityExperimental) && m.Value == nil {
		return fmt.Errorf("metric %q requires a value when quality is %q", m.Name, m.Quality)
	}
	return nil
}

// MemoryReading preserves the source-specific facts that make up a host's
// memory picture. A dedicated framebuffer value and UMA values are deliberately
// separate records, never a fallback hierarchy.
type MemoryReading struct {
	MetricValue
	Kind MemorySemantics `json:"kind"`
}

func (m MemoryReading) Validate() error {
	if err := m.MetricValue.Validate(); err != nil {
		return err
	}
	if !m.Kind.Valid() {
		return fmt.Errorf("memory metric %q has invalid semantics %q", m.Name, m.Kind)
	}
	if m.Kind == MemoryDedicatedFramebuffer && strings.Contains(strings.ToLower(m.Semantics), "unified") {
		return fmt.Errorf("dedicated framebuffer metric %q cannot claim unified-memory semantics", m.Name)
	}
	return nil
}
