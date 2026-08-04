package update

import (
	"fmt"
	"os"
	"path/filepath"
)

type Activation struct {
	PreviousBinary string
	ActiveBinary   string
}

// Activate atomically replaces the active binary only after a verified staged
// candidate exists. The previous binary is retained for operator rollback.
func Activate(stagedBinary, activeBinary string) (Activation, error) {
	info, err := os.Stat(stagedBinary)
	if err != nil {
		return Activation{}, fmt.Errorf("inspect staged binary: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return Activation{}, fmt.Errorf("staged binary is not executable")
	}
	if err := os.MkdirAll(filepath.Dir(activeBinary), 0o755); err != nil {
		return Activation{}, err
	}
	previous := activeBinary + ".previous"
	if _, err := os.Stat(activeBinary); err == nil {
		_ = os.Remove(previous)
		if err := os.Rename(activeBinary, previous); err != nil {
			return Activation{}, fmt.Errorf("preserve previous binary: %w", err)
		}
	}
	if err := os.Rename(stagedBinary, activeBinary); err != nil {
		if _, restoreErr := os.Stat(previous); restoreErr == nil {
			_ = os.Rename(previous, activeBinary)
		}
		return Activation{}, fmt.Errorf("activate staged binary: %w", err)
	}
	return Activation{PreviousBinary: previous, ActiveBinary: activeBinary}, nil
}

func Rollback(activeBinary string) error {
	previous := activeBinary + ".previous"
	if _, err := os.Stat(previous); err != nil {
		return fmt.Errorf("previous binary is unavailable: %w", err)
	}
	failed := activeBinary + ".failed"
	_ = os.Remove(failed)
	if err := os.Rename(activeBinary, failed); err != nil {
		return fmt.Errorf("preserve failed binary: %w", err)
	}
	if err := os.Rename(previous, activeBinary); err != nil {
		return fmt.Errorf("restore previous binary: %w", err)
	}
	return nil
}
