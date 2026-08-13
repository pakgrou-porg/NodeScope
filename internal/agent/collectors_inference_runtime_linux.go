//go:build linux

package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// InferenceRuntimeProcessCollector observes only configured process comm names
// for locally approved inference runtimes such as vLLM, llama.cpp, or LM
// Studio. It intentionally never reads command arguments, environment files,
// network endpoints, prompts, completions, request bodies, or response bodies.
type InferenceRuntimeProcessCollector struct {
	processNames []string
	procRoot     string
}

func NewInferenceRuntimeProcessCollector(processNames []string) *InferenceRuntimeProcessCollector {
	return newInferenceRuntimeProcessCollector(processNames, "/proc")
}

func newInferenceRuntimeProcessCollector(processNames []string, procRoot string) *InferenceRuntimeProcessCollector {
	return &InferenceRuntimeProcessCollector{processNames: append([]string(nil), processNames...), procRoot: procRoot}
}

func (collector *InferenceRuntimeProcessCollector) Name() string {
	return "inference_runtime_processes"
}

func (collector *InferenceRuntimeProcessCollector) Collect(_ context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	if len(collector.processNames) == 0 {
		return []telemetry.Sample{unavailableSample("runtime-discovery", "inference.runtime.running", "state", "procfs-comm", "no inference runtime process names are configured", observedAt)}, nil
	}
	entries, err := os.ReadDir(collector.procRoot)
	if err != nil {
		return []telemetry.Sample{unavailableSample("runtime-discovery", "inference.runtime.running", "state", "procfs-comm", "runtime process table is unavailable", observedAt)}, nil
	}
	configured := make(map[string]bool, len(collector.processNames))
	for _, name := range collector.processNames {
		configured[name] = true
	}
	observed := make(map[string]bool, len(configured))
	for _, entry := range entries {
		if _, err := strconv.ParseUint(entry.Name(), 10, 64); err != nil {
			continue
		}
		// /proc/<pid>/comm contains the kernel process name only. Do not replace
		// this read with cmdline or environ; either can expose protected content.
		contents, err := os.ReadFile(filepath.Join(collector.procRoot, entry.Name(), "comm"))
		if err != nil {
			continue
		}
		if name := strings.TrimSpace(string(contents)); configured[name] {
			observed[name] = true
		}
	}
	names := append([]string(nil), collector.processNames...)
	sort.Strings(names)
	samples := make([]telemetry.Sample, 0, len(names))
	for _, name := range names {
		value := 0.0
		semantics := fmt.Sprintf("configured inference runtime process %q is not running", name)
		if observed[name] {
			value = 1
			semantics = fmt.Sprintf("configured inference runtime process %q is running", name)
		}
		samples = append(samples, numericSample("runtime:"+name, "inference.runtime.running", "state", value, "procfs-comm", semantics, observedAt))
	}
	return samples, nil
}
