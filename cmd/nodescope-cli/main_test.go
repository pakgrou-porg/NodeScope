package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pakgrou-porg/nodescope/internal/consoleclient"
)

func TestWriteOutputPreservesMachineReadableFormats(t *testing.T) {
	rows := []consoleclient.HostStatus{{HostSlug: "framework", DisplayName: "Framework", FreshnessState: "fresh", CurrentMetricCount: 8}}
	for name, format := range map[string]string{"json": "json", "ndjson": "ndjson"} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			if err := writeOutput(&output, format, rows); err != nil {
				t.Fatalf("writeOutput() error = %v", err)
			}
			if !strings.Contains(output.String(), `"host_slug":"framework"`) || strings.Contains(output.String(), "prompt") {
				t.Fatalf("output = %q", output.String())
			}
		})
	}
}

func TestWriteOutputTableDoesNotInventRemoteInterval(t *testing.T) {
	var output bytes.Buffer
	if err := writeOutput(&output, "table", []consoleclient.HostStatus{{DisplayName: "Framework", FreshnessState: "fresh"}}); err != nil {
		t.Fatalf("writeOutput() error = %v", err)
	}
	if !strings.Contains(output.String(), "n/a") || strings.Contains(output.String(), "0s") {
		t.Fatalf("table = %q", output.String())
	}
}
