//go:build linux

package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type persistedState struct {
	BootID   string `json:"bootId"`
	Sequence uint64 `json:"sequence"`
}

type SequenceStore struct {
	path  string
	state persistedState
	mu    sync.Mutex
}

func OpenSequenceStore(directory string) (*SequenceStore, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create agent state directory: %w", err)
	}
	bootID, err := readBootID()
	if err != nil {
		return nil, err
	}
	store := &SequenceStore{path: filepath.Join(directory, "sequence.json"), state: persistedState{BootID: bootID}}
	contents, err := os.ReadFile(store.path)
	if err == nil {
		if err := json.Unmarshal(contents, &store.state); err != nil {
			return nil, fmt.Errorf("decode agent state: %w", err)
		}
		if store.state.BootID != bootID {
			store.state = persistedState{BootID: bootID}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read agent state: %w", err)
	}
	return store, nil
}

func (store *SequenceStore) Next() (bootID string, sequence uint64, err error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.state.Sequence++
	contents, err := json.Marshal(store.state)
	if err != nil {
		return "", 0, fmt.Errorf("encode agent state: %w", err)
	}
	temporary := store.path + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o600); err != nil {
		return "", 0, fmt.Errorf("write agent state: %w", err)
	}
	if err := os.Rename(temporary, store.path); err != nil {
		return "", 0, fmt.Errorf("commit agent state: %w", err)
	}
	return store.state.BootID, store.state.Sequence, nil
}

func readBootID() (string, error) {
	contents, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", fmt.Errorf("read Linux boot ID: %w", err)
	}
	bootID := strings.TrimSpace(string(contents))
	if bootID == "" {
		return "", fmt.Errorf("Linux boot ID is empty")
	}
	if _, err := strconv.ParseUint(strings.ReplaceAll(strings.ReplaceAll(bootID, "-", ""), "a", "a"), 16, 64); err != nil {
		// The UUID can exceed uint64; validate general shape without emitting it.
		if len(strings.ReplaceAll(bootID, "-", "")) != 32 {
			return "", fmt.Errorf("Linux boot ID has invalid shape")
		}
	}
	return bootID, nil
}
