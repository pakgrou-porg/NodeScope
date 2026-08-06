//go:build windows

package agent

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/windows"
)

// persistedState is written after each accepted collection so a Windows agent
// retains monotonically increasing sequence numbers for its current session.
type persistedState struct {
	BootID   string `json:"bootId"`
	Sequence uint64 `json:"sequence"`
}

// SequenceStore uses a newly generated agent-session identifier on every
// Windows agent start. The value is deliberately not presented as an operating
// system boot identifier because this baseline does not call unqualified
// Windows management APIs for boot-time discovery.
type SequenceStore struct {
	path  string
	state persistedState
	mu    sync.Mutex
}

func OpenSequenceStore(directory string) (*SequenceStore, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create Windows agent state directory: %w", err)
	}
	sessionID, err := newWindowsSessionID()
	if err != nil {
		return nil, err
	}
	return &SequenceStore{
		path:  filepath.Join(directory, "sequence.json"),
		state: persistedState{BootID: sessionID},
	}, nil
}

func (store *SequenceStore) Next() (bootID string, sequence uint64, err error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.state.Sequence++
	contents, err := json.Marshal(store.state)
	if err != nil {
		return "", 0, fmt.Errorf("encode Windows agent state: %w", err)
	}
	temporary := store.path + ".tmp"
	if err := os.WriteFile(temporary, contents, 0o600); err != nil {
		return "", 0, fmt.Errorf("write Windows agent state: %w", err)
	}
	if err := replaceWindowsState(temporary, store.path); err != nil {
		return "", 0, fmt.Errorf("commit Windows agent state: %w", err)
	}
	return store.state.BootID, store.state.Sequence, nil
}

func replaceWindowsState(temporary, destination string) error {
	temporaryPath, err := windows.UTF16PtrFromString(temporary)
	if err != nil {
		return fmt.Errorf("encode temporary state path: %w", err)
	}
	destinationPath, err := windows.UTF16PtrFromString(destination)
	if err != nil {
		return fmt.Errorf("encode destination state path: %w", err)
	}
	if err := windows.MoveFileEx(temporaryPath, destinationPath, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return err
	}
	return nil
}

func newWindowsSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate Windows agent session identifier: %w", err)
	}
	return "windows-agent-session-" + hex.EncodeToString(bytes), nil
}
