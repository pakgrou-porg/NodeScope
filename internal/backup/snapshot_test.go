package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeLeaser struct{ current bool }

func (leaser fakeLeaser) Acquire(context.Context, string, string, time.Duration) (Lease, error) {
	return Lease{FencingToken: 7, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (leaser fakeLeaser) Current(context.Context, string, string, int64) (bool, error) {
	return leaser.current, nil
}

type sequencedLeaser struct {
	results []bool
	calls   int
}

func (leaser *sequencedLeaser) Acquire(context.Context, string, string, time.Duration) (Lease, error) {
	return Lease{FencingToken: 7, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (leaser *sequencedLeaser) Current(context.Context, string, string, int64) (bool, error) {
	if leaser.calls >= len(leaser.results) {
		return false, nil
	}
	result := leaser.results[leaser.calls]
	leaser.calls++
	return result, nil
}

type fakeDump struct{ arguments []string }

func (executor *fakeDump) Dump(_ context.Context, _ Connection, arguments []string, destination string) error {
	executor.arguments = append([]string(nil), arguments...)
	return os.WriteFile(destination, []byte("pg_dump"), 0o600)
}

func TestDefaultBackupExcludesRawTelemetry(t *testing.T) {
	executor := &fakeDump{}
	runner := Runner{Leaser: fakeLeaser{current: true}, Executor: executor, Now: func() time.Time { return time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) }}
	output, err := runner.Run(context.Background(), Request{ReplicaID: "framework", OutputDirectory: t.TempDir(), Mode: ModeDefault, RetentionDays: 10})
	if err != nil {
		t.Fatalf("run backup: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("backup archive missing: %v", err)
	}
	if !strings.Contains(strings.Join(executor.arguments, " "), "--exclude-table-data=nodescope.raw_metric_samples") {
		t.Fatalf("default backup included raw telemetry: %#v", executor.arguments)
	}
}
func TestBackupRefusesPublicationAfterLeaseLoss(t *testing.T) {
	executor := &fakeDump{}
	runner := Runner{Leaser: fakeLeaser{current: false}, Executor: executor}
	_, err := runner.Run(context.Background(), Request{ReplicaID: "framework", OutputDirectory: filepath.Join(t.TempDir(), "backup"), Mode: ModeDefault, RetentionDays: 10})
	if err == nil || !strings.Contains(err.Error(), "lease was lost") {
		t.Fatalf("expected fencing refusal, got %v", err)
	}
}

func TestBackupRefusesFinalPublicationWhenLeaseIsLostDuringArchiveCreation(t *testing.T) {
	directory := t.TempDir()
	leaser := &sequencedLeaser{results: []bool{true, false}}
	runner := Runner{Leaser: leaser, Executor: &fakeDump{}, Now: func() time.Time { return time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) }}
	_, err := runner.Run(context.Background(), Request{ReplicaID: "framework", OutputDirectory: directory, Mode: ModeDefault, RetentionDays: 10})
	if err == nil || !strings.Contains(err.Error(), "lease was lost before final publication") {
		t.Fatalf("expected final fencing refusal, got %v", err)
	}
	if leaser.calls != 2 {
		t.Fatalf("expected lease checks before and after archive creation, got %d", leaser.calls)
	}
	entries, readErr := os.ReadDir(directory)
	if readErr != nil {
		t.Fatalf("read output directory: %v", readErr)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") || strings.HasSuffix(entry.Name(), ".partial") {
			t.Fatalf("stale replica left publishable archive artifact %q", entry.Name())
		}
	}
}

func TestArchiveCreationNeverOverwritesOrFollowsExistingPartialPath(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "nodescope.dump"), []byte("pg_dump"), 0o600); err != nil {
		t.Fatalf("write source dump: %v", err)
	}
	for name, prepare := range map[string]func(t *testing.T, destination string){
		"regular file": func(t *testing.T, destination string) {
			t.Helper()
			if err := os.WriteFile(destination, []byte("preserve"), 0o600); err != nil {
				t.Fatalf("write existing partial: %v", err)
			}
		},
		"symlink": func(t *testing.T, destination string) {
			t.Helper()
			target := filepath.Join(t.TempDir(), "outside.tar.gz")
			if err := os.WriteFile(target, []byte("preserve"), 0o600); err != nil {
				t.Fatalf("write symlink target: %v", err)
			}
			if err := os.Symlink(target, destination); err != nil {
				t.Fatalf("create partial symlink: %v", err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "archive.tar.gz.partial")
			prepare(t, destination)
			if err := tarDirectory(source, destination); err == nil {
				t.Fatal("expected existing partial path to be rejected")
			}
			info, err := os.Lstat(destination)
			if err != nil {
				t.Fatalf("existing partial path disappeared: %v", err)
			}
			if name == "symlink" && info.Mode()&os.ModeSymlink == 0 {
				t.Fatal("existing partial symlink was replaced")
			}
			contents, err := os.ReadFile(destination)
			if err != nil || string(contents) != "preserve" {
				t.Fatalf("existing partial path or symlink target was changed: contents=%q err=%v", contents, err)
			}
		})
	}
}

func TestArchiveCreationRejectsSymlinkInStagingSource(t *testing.T) {
	source := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside-dump")
	if err := os.WriteFile(target, []byte("outside-canary"), 0o600); err != nil {
		t.Fatalf("write outside target: %v", err)
	}
	if err := os.Symlink(target, filepath.Join(source, "nodescope.dump")); err != nil {
		t.Fatalf("create staged symlink: %v", err)
	}
	destination := filepath.Join(t.TempDir(), "archive.tar.gz.partial")
	err := tarDirectory(source, destination)
	if err == nil || !strings.Contains(err.Error(), "unsupported non-regular") {
		t.Fatalf("expected staged symlink rejection, got %v", err)
	}
	contents, readErr := os.ReadFile(target)
	if readErr != nil || string(contents) != "outside-canary" {
		t.Fatalf("outside target changed while archiving staged symlink: contents=%q err=%v", contents, readErr)
	}
}
