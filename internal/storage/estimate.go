package storage

import (
	"fmt"
	"math"
	"time"
)

// ProbeObservation is a summary emitted from real NodeScope ingestion traffic.
// The estimator deliberately accepts observed compressed bytes rather than
// fabricating metric sizes or pretending a synthetic load is a feasibility run.
type ProbeObservation struct {
	HostID              string
	ObservedDuration    time.Duration
	BatchCount          int64
	CompressedByteCount int64
	MetricValueCount    int64
}

type RetentionPlan struct {
	RawRetention        time.Duration
	OneMinuteRetention  time.Duration
	FiveMinuteRetention time.Duration
	TenMinuteRetention  time.Duration
}

func DefaultRetentionPlan() RetentionPlan {
	return RetentionPlan{
		RawRetention:        48 * time.Hour,
		OneMinuteRetention:  48 * time.Hour,
		FiveMinuteRetention: 7 * 24 * time.Hour,
		TenMinuteRetention:  30 * 24 * time.Hour,
	}
}

type Estimate struct {
	AverageCompressedBatchBytes float64
	BatchesPerDay               float64
	RawRetentionBytes           int64
	OneMinuteRollupBuckets      int64
	FiveMinuteRollupBuckets     int64
	TenMinuteRollupBuckets      int64
	ProjectedTotalBytes         int64
}

func EstimateRetention(observation ProbeObservation, plan RetentionPlan) (Estimate, error) {
	if observation.HostID == "" {
		return Estimate{}, fmt.Errorf("host ID is required")
	}
	if observation.ObservedDuration <= 0 || observation.BatchCount <= 0 || observation.CompressedByteCount <= 0 {
		return Estimate{}, fmt.Errorf("observed duration, batch count, and compressed bytes must be positive")
	}
	if plan.RawRetention <= 0 || plan.OneMinuteRetention <= 0 || plan.FiveMinuteRetention <= 0 || plan.TenMinuteRetention <= 0 {
		return Estimate{}, fmt.Errorf("all retention durations must be positive")
	}

	averageBatchBytes := float64(observation.CompressedByteCount) / float64(observation.BatchCount)
	batchesPerDay := float64(24*time.Hour) / float64(observation.ObservedDuration) * float64(observation.BatchCount)
	rawBytes := int64(math.Ceil(averageBatchBytes * batchesPerDay * plan.RawRetention.Hours() / 24))

	// Rollups contain aggregates rather than raw sample payloads. The conservative
	// envelope below budgets 160 bytes per metric-series bucket plus a 96-byte
	// DDSketch payload, then adds a 25% index/row overhead. Probe data supplies
	// the actual observed metric series count.
	seriesCount := math.Max(1, math.Ceil(float64(observation.MetricValueCount)/float64(observation.BatchCount)))
	bucketBytes := int64(math.Ceil(seriesCount * (160 + 96) * 1.25))
	oneMinute := int64(plan.OneMinuteRetention.Minutes())
	fiveMinute := int64(plan.FiveMinuteRetention.Minutes() / 5)
	tenMinute := int64(plan.TenMinuteRetention.Minutes() / 10)
	rollupBytes := bucketBytes * (oneMinute + fiveMinute + tenMinute)

	return Estimate{
		AverageCompressedBatchBytes: averageBatchBytes,
		BatchesPerDay:               batchesPerDay,
		RawRetentionBytes:           rawBytes,
		OneMinuteRollupBuckets:      oneMinute,
		FiveMinuteRollupBuckets:     fiveMinute,
		TenMinuteRollupBuckets:      tenMinute,
		ProjectedTotalBytes:         rawBytes + rollupBytes,
	}, nil
}
