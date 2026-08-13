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

func TestInferenceRuntimeProcessCollectorUsesOnlyConfiguredCommNames(t *testing.T) {
	procRoot := t.TempDir()
	writeRuntimeProcFixture(t, procRoot, "101", "vllm\n")
	writeRuntimeProcFixture(t, procRoot, "202", "unrelated\n")
	if err := os.WriteFile(filepath.Join(procRoot, "101", "cmdline"), []byte("--prompt=runtime-prompt-canary"), 0600); err != nil {
		t.Fatal(err)
	}
	collector := newInferenceRuntimeProcessCollector([]string{"llama-server", "vllm"}, procRoot)
	samples, err := collector.Collect(context.Background(), time.Unix(100, 0))
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(samples) != 2 {
		t.Fatalf("sample count = %d", len(samples))
	}
	for _, sample := range samples {
		if sample.Metric.Quality != domain.QualityFresh || sample.Metric.Value == nil || strings.Contains(sample.Metric.Semantics, "runtime-prompt-canary") {
			t.Fatalf("unsafe runtime sample %#v", sample)
		}
		switch sample.DeviceID {
		case "runtime:vllm":
			if *sample.Metric.Value != 1 {
				t.Fatalf("vllm value = %v", *sample.Metric.Value)
			}
		case "runtime:llama-server":
			if *sample.Metric.Value != 0 {
				t.Fatalf("llama-server value = %v", *sample.Metric.Value)
			}
		default:
			t.Fatalf("unexpected device ID %q", sample.DeviceID)
		}
	}
}

func TestInferenceRuntimeProcessCollectorReportsExplicitUnavailableWithoutConfiguration(t *testing.T) {
	collector := newInferenceRuntimeProcessCollector(nil, t.TempDir())
	samples, err := collector.Collect(context.Background(), time.Unix(100, 0))
	if err != nil || len(samples) != 1 || samples[0].Metric.Quality != domain.QualityUnavailable || samples[0].Metric.Value != nil {
		t.Fatalf("unexpected unavailable runtime output: samples=%#v err=%v", samples, err)
	}
}

func TestInferenceRuntimeProcessCollectorReportsExplicitUnavailableWhenProcfsCannotBeRead(t *testing.T) {
	collector := newInferenceRuntimeProcessCollector([]string{"vllm"}, filepath.Join(t.TempDir(), "missing"))
	samples, err := collector.Collect(context.Background(), time.Unix(100, 0))
	if err != nil || len(samples) != 1 || samples[0].Metric.Quality != domain.QualityUnavailable || samples[0].Metric.Value != nil {
		t.Fatalf("unexpected unavailable runtime output: samples=%#v err=%v", samples, err)
	}
}

func writeRuntimeProcFixture(t *testing.T, root, pid, comm string) {
	t.Helper()
	directory := filepath.Join(root, pid)
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "comm"), []byte(comm), 0600); err != nil {
		t.Fatal(err)
	}
}
