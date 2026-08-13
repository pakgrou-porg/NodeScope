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
