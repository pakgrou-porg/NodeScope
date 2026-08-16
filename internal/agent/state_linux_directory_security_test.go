//go:build linux

package agent

import (
	"os"
	"strings"
	"testing"
)

func TestOpenSequenceStoreRejectsGroupOrWorldWritableDirectory(t *testing.T) {
	directory := t.TempDir()
	if err := os.Chmod(directory, 0o777); err != nil {
		t.Fatalf("chmod state directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(directory, 0o700) })

	_, err := OpenSequenceStore(directory)
	if err == nil || !strings.Contains(err.Error(), "must not be group- or world-writable") {
		t.Fatalf("expected writable state directory rejection, got %v", err)
	}
}
