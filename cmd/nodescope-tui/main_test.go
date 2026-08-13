package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/consoleclient"
)

func TestRenderRemoteStatusDoesNotInventInterval(t *testing.T) {
	var output bytes.Buffer
	render(&output, refreshMessage{rows: []consoleclient.HostStatus{{DisplayName: "Framework", FreshnessState: "fresh"}}}, 5*time.Second)
	if !strings.Contains(output.String(), "n/a") || strings.Contains(output.String(), "0s") {
		t.Fatalf("rendered TUI = %q", output.String())
	}
	if strings.Contains(strings.ToLower(output.String()), "prompt") && !strings.Contains(output.String(), "Prompt and response content are never queried") {
		t.Fatalf("rendered TUI did not retain the no-content boundary: %q", output.String())
	}
}

func TestRenderFailureDeclaresNoSubstituteValues(t *testing.T) {
	var output bytes.Buffer
	render(&output, refreshMessage{err: errors.New("unavailable")}, 5*time.Second)
	if !strings.Contains(output.String(), "No substitute values are shown.") {
		t.Fatalf("rendered TUI = %q", output.String())
	}
}
