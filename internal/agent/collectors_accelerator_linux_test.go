//go:build linux

package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func TestAMDDRMCollectorMarksEvidenceExperimental(t *testing.T) {
	card := t.TempDir()
	device := filepath.Join(card, "device")
	if err := os.Mkdir(device, 0755); err != nil {
		t.Fatalf("create DRM device fixture: %v", err)
	}
	for name, value := range map[string]string{
		"gpu_busy_percent":    "42\n",
		"mem_info_vram_total": "8589934592\n",
		"mem_info_vram_used":  "2147483648\n",
	} {
		if err := os.WriteFile(filepath.Join(device, name), []byte(value), 0644); err != nil {
			t.Fatalf("write %s fixture: %v", name, err)
		}
	}
	samples := collectAMDDRM("card0", card, time.Now().UTC())
	if len(samples) == 0 {
		t.Fatal("expected AMD DRM evidence")
	}
	for _, sample := range samples {
		if sample.Metric.Source != "sysfs-experimental" || !strings.Contains(sample.Metric.Semantics, "experimental Fedora AMD DRM telemetry") {
			t.Fatalf("AMD DRM sample must retain experimental provenance: %#v", sample)
		}
		if sample.Metric.Value != nil && sample.Metric.Quality != domain.QualityExperimental {
			t.Fatalf("value-bearing AMD DRM sample must be experimental: %#v", sample)
		}
		if sample.Metric.Value == nil && sample.Metric.Quality != domain.QualityUnavailable {
			t.Fatalf("missing AMD DRM sample must remain explicitly unavailable: %#v", sample)
		}
	}
}

func TestXDNACollectorMarksUnavailableEvidenceExperimental(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	samples, err := NewXDNACollector().Collect(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("collect XDNA unavailable evidence: %v", err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected availability and utilization evidence, got %#v", samples)
	}
	for _, sample := range samples {
		if sample.Metric.Source != "xrt-smi-experimental" || !strings.Contains(sample.Metric.Semantics, "experimental Fedora AMD XDNA telemetry") {
			t.Fatalf("XDNA sample must retain experimental provenance: %#v", sample)
		}
	}
}

func TestXDNACollectorMarksAvailableEvidenceExperimental(t *testing.T) {
	binDir := t.TempDir()
	probe := filepath.Join(binDir, "xrt-smi")
	if err := os.WriteFile(probe, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write xrt-smi fixture: %v", err)
	}
	t.Setenv("PATH", binDir)
	samples, err := NewXDNACollector().Collect(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("collect XDNA available evidence: %v", err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected availability and utilization evidence, got %#v", samples)
	}
	if samples[0].Metric.Quality != domain.QualityExperimental || samples[0].Metric.Value == nil {
		t.Fatalf("available XDNA evidence must be experimental and value-bearing: %#v", samples[0])
	}
	if samples[1].Metric.Quality != domain.QualityUnavailable || samples[1].Metric.Value != nil {
		t.Fatalf("unsupported XDNA utilization must remain unavailable: %#v", samples[1])
	}
}
