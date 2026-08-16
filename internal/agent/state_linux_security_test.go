//go:build linux

package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenSequenceStoreRejectsSymlinkedSequenceState(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target-state.json")
	if err := os.WriteFile(target, []byte(`{"bootId":"target","sequence":9}`), 0o600); err != nil {
		t.Fatalf("write target state: %v", err)
	}
	if err := os.Symlink(target, filepath.Join(directory, "sequence.json")); err != nil {
		t.Fatalf("create state symlink: %v", err)
	}
	_, err := OpenSequenceStore(directory)
	if err == nil || !strings.Contains(err.Error(), "direct regular file") {
		t.Fatalf("expected symlinked sequence state rejection, err=%v", err)
	}
}

func TestSequenceStoreRejectsSymlinkedTemporaryState(t *testing.T) {
	directory := t.TempDir()
	store, err := OpenSequenceStore(directory)
	if err != nil {
		t.Fatalf("open sequence store: %v", err)
	}
	target := filepath.Join(directory, "target-temporary-state.json")
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatalf("write temporary target: %v", err)
	}
	if err := os.Symlink(target, store.path+".tmp"); err != nil {
		t.Fatalf("create temporary state symlink: %v", err)
	}
	_, _, err = store.Next()
	if err == nil || !strings.Contains(err.Error(), "direct regular file") {
		t.Fatalf("expected symlinked temporary state rejection, err=%v", err)
	}
}
