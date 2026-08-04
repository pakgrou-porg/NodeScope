package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Mode string

const (
	ModeDefault Mode = "default"
	ModeFull    Mode = "full"
)

type Connection struct {
	Host         string
	Port         string
	Database     string
	User         string
	PasswordFile string
}
type DumpExecutor interface {
	Dump(context.Context, Connection, []string, string) error
}
type Request struct {
	ReplicaID       string
	OutputDirectory string
	Mode            Mode
	RetentionDays   int
	Connection      Connection
}
type Manifest struct {
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	ReplicaID    string    `json:"replica_id"`
	Mode         Mode      `json:"mode"`
	FencingToken int64     `json:"fencing_token"`
	Files        []string  `json:"files"`
}
type Runner struct {
	Leaser   Leaser
	Executor DumpExecutor
	Now      func() time.Time
}

func (runner Runner) Run(ctx context.Context, request Request) (string, error) {
	if runner.Leaser == nil || runner.Executor == nil {
		return "", fmt.Errorf("backup leaser and dump executor are required")
	}
	if request.ReplicaID == "" || request.OutputDirectory == "" {
		return "", fmt.Errorf("replica ID and output directory are required")
	}
	if request.Mode != ModeDefault && request.Mode != ModeFull {
		return "", fmt.Errorf("backup mode must be default or full")
	}
	if request.RetentionDays < 1 || request.RetentionDays > 365 {
		return "", fmt.Errorf("retention must be 1..365 days")
	}
	lease, err := runner.Leaser.Acquire(ctx, "daily_backup", request.ReplicaID, 30*time.Minute)
	if err != nil {
		return "", fmt.Errorf("acquire backup lease: %w", err)
	}
	if err := os.MkdirAll(request.OutputDirectory, 0o750); err != nil {
		return "", err
	}
	now := time.Now().UTC()
	if runner.Now != nil {
		now = runner.Now().UTC()
	}
	basename := fmt.Sprintf("nodescope-%s-%s-%d", request.Mode, now.Format("20060102T150405Z"), lease.FencingToken)
	staging := filepath.Join(request.OutputDirectory, "."+basename+".partial")
	final := filepath.Join(request.OutputDirectory, basename+".tar.gz")
	if err := os.Mkdir(staging, 0o700); err != nil {
		return "", err
	}
	defer os.RemoveAll(staging)
	dump := filepath.Join(staging, "nodescope.dump")
	if err := runner.Executor.Dump(ctx, request.Connection, dumpArguments(request.Mode), dump); err != nil {
		return "", fmt.Errorf("dump NodeScope schema: %w", err)
	}
	manifest := Manifest{Version: 1, CreatedAt: now, ReplicaID: request.ReplicaID, Mode: request.Mode, FencingToken: lease.FencingToken, Files: []string{"nodescope.dump"}}
	manifestFile := filepath.Join(staging, "manifest.json")
	bytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(manifestFile, append(bytes, '\n'), 0o600); err != nil {
		return "", err
	}
	manifest.Files = append(manifest.Files, "manifest.json")
	current, err := runner.Leaser.Current(ctx, "daily_backup", request.ReplicaID, lease.FencingToken)
	if err != nil {
		return "", err
	}
	if !current {
		return "", fmt.Errorf("backup lease was lost before publication")
	}
	archivePartial := final + ".partial"
	if err := tarDirectory(staging, archivePartial); err != nil {
		return "", err
	}
	if err := os.Rename(archivePartial, final); err != nil {
		return "", err
	}
	if err := prune(request.OutputDirectory, now.AddDate(0, 0, -request.RetentionDays)); err != nil {
		return "", err
	}
	return final, nil
}

func dumpArguments(mode Mode) []string {
	if mode == ModeFull {
		return []string{"--schema=nodescope"}
	}
	return []string{
		"--schema=nodescope",
		"--exclude-table-data=nodescope.raw_metric_samples",
		"--exclude-table-data=nodescope.ingest_receipts",
	}
}
func tarDirectory(source, destination string) error {
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	return filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil || strings.Contains(relative, "..") {
			return fmt.Errorf("invalid backup path")
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relative
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		_, err = io.Copy(tarWriter, input)
		return err
	})
}
func prune(directory string, before time.Time) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	paths := make([]string, 0)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "nodescope-") && strings.HasSuffix(entry.Name(), ".tar.gz") {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.ModTime().Before(before) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}
