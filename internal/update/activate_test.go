package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActivatePreservesPreviousBinaryAndRollsBack(t *testing.T) {
	directory := t.TempDir()
	active := filepath.Join(directory, "nodescope-agent")
	staged := filepath.Join(directory, "staged-agent")
	if err := os.WriteFile(active, []byte("old"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o700); err != nil {
		t.Fatal(err)
	}
	result, err := Activate(staged, active)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if string(mustRead(t, result.ActiveBinary)) != "new" || string(mustRead(t, result.PreviousBinary)) != "old" {
		t.Fatal("activation did not preserve expected versions")
	}
	if err := Rollback(active); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if string(mustRead(t, active)) != "old" {
		t.Fatal("rollback did not restore previous binary")
	}
}
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	value, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
