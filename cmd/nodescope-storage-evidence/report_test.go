package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAssessEvidenceUsesReceiptTimeCompletenessSignals(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	valid := evidence{
		FirstReceivedAt:      &now,
		LastReceivedAt:       &now,
		ReceivedBatchCount:   10,
		ExpectedBatchCount:   10,
		CompletenessPercent:  100,
		MaxGapSeconds:        10,
		MetricCardinality:    4,
		TotalCompressedBytes: 10,
	}
	if assessment := assessEvidence(valid, 5); !assessment.Complete || len(assessment.Reasons) != 0 {
		t.Fatalf("expected complete receipt-time evidence, got %#v", assessment)
	}
	valid.MaxGapSeconds = 16
	valid.MetricCardinality = 0
	assessment := assessEvidence(valid, 5)
	if assessment.Complete {
		t.Fatal("expected incomplete evidence")
	}
	if got := strings.Join(assessment.Reasons, ","); !strings.Contains(got, "receipt_gap_exceeds_three_collection_intervals") || !strings.Contains(got, "no_metric_cardinality") {
		t.Fatalf("unexpected reasons: %q", got)
	}
}

func TestAssessEvidenceRejectsUnsupportedCollectionIntervals(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	valid := evidence{
		FirstReceivedAt:      &now,
		LastReceivedAt:       &now,
		ReceivedBatchCount:   1,
		ExpectedBatchCount:   1,
		CompletenessPercent:  100,
		MaxGapSeconds:        0,
		MetricCardinality:    1,
		TotalCompressedBytes: 1,
	}
	for _, intervalSeconds := range []int{0, 61} {
		assessment := assessEvidence(valid, intervalSeconds)
		if assessment.Complete || !strings.Contains(strings.Join(assessment.Reasons, ","), "collection_interval_seconds_must_be_between_1_and_60") {
			t.Fatalf("interval %d was not rejected: %#v", intervalSeconds, assessment)
		}
	}
}

func TestWriteReportAtomicUsesSafeDynamicFilename(t *testing.T) {
	directory := t.TempDir()
	generated := time.Date(2026, 8, 13, 12, 34, 56, 0, time.UTC)
	value := report{
		evidence:    evidence{HostSlug: "framework/primary", ReceivedBatchCount: 1, ExpectedBatchCount: 1, CompletenessPercent: 100, MetricCardinality: 1},
		GeneratedAt: generated,
		Assessment:  evidenceAssessment{Complete: true},
	}
	path, err := writeReportAtomic(directory, value)
	if err != nil {
		t.Fatalf("write report: %v", err)
	}
	if filepath.Dir(path) != directory || filepath.Base(path) != "nodescope-storage-evidence-framework_primary-20260813T123456Z.json" {
		t.Fatalf("unexpected report path %q", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat report: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected report mode 0600, got %o", info.Mode().Perm())
	}
	var written report
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	if err := json.Unmarshal(payload, &written); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if written.ReportPath != path || written.HostSlug != "framework/primary" {
		t.Fatalf("unexpected written report %#v", written)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("list report directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("temporary report was not cleaned up: %s", entry.Name())
		}
	}
}
