package agent

import (
	"context"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// Collector supplies host telemetry while preserving each metric's provenance
// and quality. Platform-specific collectors must explicitly report unsupported
// or unavailable data rather than substitute an inferred value.
type Collector interface {
	Name() string
	Collect(context.Context, time.Time) ([]telemetry.Sample, error)
}

// ContainerInventoryCollector optionally supplies complete container rows in
// addition to metric samples. Only platform collectors that can safely expose
// a read-only inventory implement this interface.
type ContainerInventoryCollector interface {
	Collector
	CollectContainerInventory(context.Context, time.Time) ([]telemetry.Sample, []telemetry.ContainerInventory, error)
}
