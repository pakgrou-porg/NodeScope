package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type verificationAssessment struct {
	Fresh   bool     `json:"fresh"`
	Reasons []string `json:"reasons"`
}

type verificationReport struct {
	status
	GeneratedAt      time.Time              `json:"generated_at"`
	MaxReceiptAgeSec int                    `json:"max_receipt_age_seconds"`
	Assessment       verificationAssessment `json:"assessment"`
	ReportPath       string                 `json:"report_path,omitempty"`
}

func assessStatus(value status, now time.Time, maxReceiptAgeSeconds int) verificationAssessment {
	reasons := make([]string, 0)
	if maxReceiptAgeSeconds < 1 {
		reasons = append(reasons, "max_receipt_age_seconds_must_be_positive")
	}
	if value.LatestReceipt == nil {
		reasons = append(reasons, "missing_server_receipt")
	} else if maxReceiptAgeSeconds > 0 && now.Sub(value.LatestReceipt.UTC()) > time.Duration(maxReceiptAgeSeconds)*time.Second {
		reasons = append(reasons, "server_receipt_exceeds_freshness_window")
	}
	if value.CurrentMetricCount < 1 {
		reasons = append(reasons, "no_current_metric_state")
	}
	if value.StaleMetricCount > 0 {
		reasons = append(reasons, "stale_metric_state_present")
	}
	return verificationAssessment{Fresh: len(reasons) == 0, Reasons: reasons}
}

func verificationFilenameComponent(value string) string {
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

func writeVerificationReportAtomic(outputDir string, value verificationReport) (string, error) {
	if strings.TrimSpace(outputDir) == "" {
		return "", nil
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return "", fmt.Errorf("create report directory: %w", err)
	}
	filename := fmt.Sprintf("nodescope-host-verification-%s-%s.json", verificationFilenameComponent(value.HostSlug), value.GeneratedAt.UTC().Format("20060102T150405Z"))
	path := filepath.Join(outputDir, filename)
	if filepath.Dir(path) != filepath.Clean(outputDir) {
		return "", fmt.Errorf("invalid report path")
	}
	value.ReportPath = path
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal verification report: %w", err)
	}
	temporary, err := os.CreateTemp(outputDir, ".nodescope-host-verification-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temporary verification report: %w", err)
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
		return "", fmt.Errorf("write temporary verification report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("sync temporary verification report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close temporary verification report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return "", fmt.Errorf("atomically publish verification report: %w", err)
	}
	cleanup = false
	return path, nil
}
