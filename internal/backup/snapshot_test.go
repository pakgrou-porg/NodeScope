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
