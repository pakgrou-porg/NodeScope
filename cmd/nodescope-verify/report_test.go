package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAssessStatusUsesServerReceiptAndState(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	receipt := now.Add(-10 * time.Second)
	value := status{LatestReceipt: &receipt, CurrentMetricCount: 3}
	if assessment := assessStatus(value, now, 15); !assessment.Fresh {
		t.Fatalf("expected fresh receipt-time state, got %#v", assessment)
	}
	value.LatestReceipt = nil
	value.CurrentMetricCount = 0
	value.StaleMetricCount = 1
	assessment := assessStatus(value, now, 15)
	if assessment.Fresh {
		t.Fatal("expected incomplete verification state")
	}
	reasons := strings.Join(assessment.Reasons, ",")
	for _, expected := range []string{"missing_server_receipt", "no_current_metric_state", "stale_metric_state_present"} {
		if !strings.Contains(reasons, expected) {
			t.Fatalf("missing %q in reasons %q", expected, reasons)
		}
	}
}

func TestWriteVerificationReportAtomic(t *testing.T) {
	directory := t.TempDir()
	generated := time.Date(2026, 8, 13, 12, 34, 56, 0, time.UTC)
	value := verificationReport{
		status:      status{HostSlug: "framework/primary", CurrentMetricCount: 1},
		GeneratedAt: generated,
		Assessment:  verificationAssessment{Fresh: true},
	}
	path, err := writeVerificationReportAtomic(directory, value)
	if err != nil {
		t.Fatalf("write verification report: %v", err)
	}
	if filepath.Dir(path) != directory || filepath.Base(path) != "nodescope-host-verification-framework_primary-20260813T123456Z.json" {
		t.Fatalf("unexpected report path %q", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat report: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected report mode 0600, got %o", info.Mode().Perm())
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var written verificationReport
	if err := json.Unmarshal(payload, &written); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if written.ReportPath != path || written.HostSlug != "framework/primary" || !written.Assessment.Fresh {
		t.Fatalf("unexpected written report %#v", written)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("list report directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("temporary verification report was not cleaned up: %s", entry.Name())
		}
	}
}
