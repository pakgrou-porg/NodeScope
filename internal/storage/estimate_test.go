package storage

import (
	"testing"
	"time"
)

func TestEstimateRetentionFromObservedTraffic(t *testing.T) {
	estimate, err := EstimateRetention(ProbeObservation{
		HostID:              "framework",
		ObservedDuration:    72 * time.Hour,
		BatchCount:          1000,
		CompressedByteCount: 4_000_000,
		MetricValueCount:    50_000,
	}, DefaultRetentionPlan())
	if err != nil {
		t.Fatalf("estimate retention: %v", err)
	}
	if estimate.AverageCompressedBatchBytes != 4000 {
		t.Fatalf("unexpected average batch size: %f", estimate.AverageCompressedBatchBytes)
	}
	if estimate.ProjectedTotalBytes <= estimate.RawRetentionBytes {
		t.Fatal("total must include rollup storage in addition to raw retention")
	}
	if estimate.OneMinuteRollupBuckets != 2880 || estimate.FiveMinuteRollupBuckets != 2016 || estimate.TenMinuteRollupBuckets != 4320 {
		t.Fatalf("unexpected rollup buckets: %#v", estimate)
	}
}

func TestEstimateRejectsSyntheticEmptyObservation(t *testing.T) {
	if _, err := EstimateRetention(ProbeObservation{HostID: "framework"}, DefaultRetentionPlan()); err == nil {
		t.Fatal("empty observation must not be treated as a feasibility measurement")
	}
}
