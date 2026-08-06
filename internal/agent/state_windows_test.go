//go:build windows

package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsSequenceStorePersistsRepeatedUpdates(t *testing.T) {
	store, err := OpenSequenceStore(t.TempDir())
	if err != nil {
		t.Fatalf("open Windows sequence store: %v", err)
	}
	bootID, first, err := store.Next()
	if err != nil {
		t.Fatalf("persist first sequence: %v", err)
	}
	returnedBootID, second, err := store.Next()
	if err != nil {
		t.Fatalf("persist second sequence: %v", err)
	}
	if bootID == "" || returnedBootID != bootID || first != 1 || second != 2 {
		t.Fatalf("unexpected sequence progression boot=%q returned=%q first=%d second=%d", bootID, returnedBootID, first, second)
	}

	contents, err := os.ReadFile(filepath.Join(filepath.Dir(store.path), "sequence.json"))
	if err != nil {
		t.Fatalf("read committed Windows sequence state: %v", err)
	}
	var persisted persistedState
	if err := json.Unmarshal(contents, &persisted); err != nil {
		t.Fatalf("decode committed Windows sequence state: %v", err)
	}
	if persisted.BootID != bootID || persisted.Sequence != 2 {
		t.Fatalf("unexpected committed Windows state %#v", persisted)
	}
}
