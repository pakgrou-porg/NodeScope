//go:build linux

package agent

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// SelectedProcessCollector observes only allowlisted process names. It never
// reads command arguments, environments, prompts, or response payloads.
type SelectedProcessCollector struct {
	processNames []string
}

func NewSelectedProcessCollector(processNames []string) *SelectedProcessCollector {
	return &SelectedProcessCollector{processNames: append([]string(nil), processNames...)}
}

func (collector *SelectedProcessCollector) Name() string { return "selected_processes" }

func (collector *SelectedProcessCollector) Collect(_ context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	if len(collector.processNames) == 0 {
		return []telemetry.Sample{unavailableSample("process-selection", "process.selected.running", "state", "procfs", "no selected process names are configured", observedAt)}, nil
	}
	observed := map[string]bool{}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if _, err := strconv.ParseUint(entry.Name(), 10, 64); err != nil {
			continue
		}
		comm, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		for _, wanted := range collector.processNames {
			if name == wanted {
				observed[wanted] = true
			}
		}
	}
	samples := make([]telemetry.Sample, 0, len(collector.processNames))
	for _, name := range collector.processNames {
		value := 0.0
		semantics := "selected process is not running"
		if observed[name] {
			value = 1
			semantics = "selected process is running"
		}
		samples = append(samples, numericSample("process:"+name, "process.selected.running", "state", value, "procfs", semantics, observedAt))
	}
	return samples, nil
}
