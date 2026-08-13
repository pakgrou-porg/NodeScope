package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type evidenceAssessment struct {
	Complete bool     `json:"complete"`
	Reasons  []string `json:"reasons"`
}

type report struct {
	evidence
	WindowStart              time.Time          `json:"window_start"`
	GeneratedAt              time.Time          `json:"generated_at"`
	CollectionIntervalSecond int                `json:"collection_interval_seconds"`
	Assessment               evidenceAssessment `json:"assessment"`
	ReportPath               string             `json:"report_path,omitempty"`
}

func assessEvidence(value evidence, intervalSeconds int) evidenceAssessment {
	reasons := make([]string, 0)
	if intervalSeconds < 1 {
		reasons = append(reasons, "collection_interval_seconds_must_be_positive")
	}
	if value.FirstReceivedAt == nil || value.LastReceivedAt == nil {
		reasons = append(reasons, "missing_server_receipt_timestamps")
	}
	if value.ReceivedBatchCount < 1 {
		reasons = append(reasons, "no_received_batches")
	}
	if value.ExpectedBatchCount < 1 {
		reasons = append(reasons, "no_expected_batch_count")
	}
	if math.IsNaN(value.CompletenessPercent) || math.IsInf(value.CompletenessPercent, 0) || value.CompletenessPercent < 0 || value.CompletenessPercent > 100 {
		reasons = append(reasons, "invalid_completeness_percent")
	}
	if math.IsNaN(value.MaxGapSeconds) || math.IsInf(value.MaxGapSeconds, 0) || value.MaxGapSeconds < 0 {
		reasons = append(reasons, "invalid_max_gap_seconds")
	} else if intervalSeconds > 0 && value.MaxGapSeconds > float64(intervalSeconds*3) {
		reasons = append(reasons, "receipt_gap_exceeds_three_collection_intervals")
	}
	if value.MetricCardinality < 1 {
		reasons = append(reasons, "no_metric_cardinality")
	}
	if value.TotalCompressedBytes < 0 || value.TelemetryRelationBytes < 0 || value.TelemetryIndexBytes < 0 || value.RawSampleRelationBytes < 0 || value.RawSampleIndexBytes < 0 {
		reasons = append(reasons, "negative_storage_size_evidence")
	}
	return evidenceAssessment{Complete: len(reasons) == 0, Reasons: reasons}
}

func sanitizeReportComponent(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('_')
		}
	}
	return builder.String()
}

func writeReportAtomic(outputDir string, value report) (string, error) {
	if strings.TrimSpace(outputDir) == "" {
		return "", nil
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return "", fmt.Errorf("create report directory: %w", err)
	}
	filename := fmt.Sprintf("nodescope-storage-evidence-%s-%s.json", sanitizeReportComponent(value.HostSlug), value.GeneratedAt.UTC().Format("20060102T150405Z"))
	path := filepath.Join(outputDir, filename)
	if filepath.Dir(path) != filepath.Clean(outputDir) {
		return "", fmt.Errorf("invalid report path")
	}
	value.ReportPath = path
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report: %w", err)
	}
	temporary, err := os.CreateTemp(outputDir, ".nodescope-storage-evidence-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temporary report: %w", err)
	}
	temporaryPath := temporary.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("set report permissions: %w", err)
	}
	if _, err := temporary.Write(append(payload, '\n')); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("write temporary report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("sync temporary report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close temporary report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return "", fmt.Errorf("atomically publish report: %w", err)
	}
	cleanup = false
	return path, nil
}
