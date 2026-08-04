// Package update verifies and stages NodeScope release artifacts before an
// administrator-approved systemd update. It never follows an unpinned release
// URL and never replaces a running binary until checksum verification succeeds.
package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Artifact struct {
	Version     string
	ArchiveURL  string
	ChecksumURL string
	BinaryName  string
}

type Downloader interface {
	Do(*http.Request) (*http.Response, error)
}

type Result struct {
	StagedBinary string
	SHA256       string
	ArchivePath  string
}

func DownloadAndStage(ctx context.Context, client Downloader, artifact Artifact, stageDirectory string) (Result, error) {
	if client == nil {
		return Result{}, fmt.Errorf("HTTP client is required")
	}
	if artifact.Version == "" || artifact.ArchiveURL == "" || artifact.ChecksumURL == "" || artifact.BinaryName == "" {
		return Result{}, fmt.Errorf("version, archive URL, checksum URL, and binary name are required")
	}
	if !strings.HasPrefix(artifact.ArchiveURL, "https://") || !strings.HasPrefix(artifact.ChecksumURL, "https://") {
		return Result{}, fmt.Errorf("release URLs must use HTTPS")
	}
	if filepath.Base(artifact.BinaryName) != artifact.BinaryName {
		return Result{}, fmt.Errorf("binary name must not contain a path")
	}
	if err := os.MkdirAll(stageDirectory, 0o700); err != nil {
		return Result{}, fmt.Errorf("create stage directory: %w", err)
	}

	checksum, err := fetchChecksum(ctx, client, artifact.ChecksumURL)
	if err != nil {
		return Result{}, err
	}
	archive, err := fetchArchive(ctx, client, artifact.ArchiveURL, stageDirectory)
	if err != nil {
		return Result{}, err
	}
	observed, err := fileSHA256(archive)
	if err != nil {
		return Result{}, err
	}
	if !strings.EqualFold(checksum, observed) {
		return Result{}, fmt.Errorf("release checksum mismatch")
	}
	staged := filepath.Join(stageDirectory, artifact.BinaryName+"."+artifact.Version)
	if err := extractBinary(archive, artifact.BinaryName, staged); err != nil {
		return Result{}, err
	}
	return Result{StagedBinary: staged, SHA256: observed, ArchivePath: archive}, nil
}

func fetchChecksum(ctx context.Context, client Downloader, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download checksum: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum endpoint returned %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(body))
	if len(fields) == 0 || len(fields[0]) != sha256.Size*2 {
		return "", fmt.Errorf("checksum response is invalid")
	}
	if _, err := hex.DecodeString(fields[0]); err != nil {
		return "", fmt.Errorf("checksum response is invalid")
	}
	return strings.ToLower(fields[0]), nil
}

func fetchArchive(ctx context.Context, client Downloader, url, directory string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download archive: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("archive endpoint returned %d", response.StatusCode)
	}
	file, err := os.CreateTemp(directory, "release-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(file.Name())
		}
	}()
	if _, err = io.Copy(file, io.LimitReader(response.Body, 256<<20)); err != nil {
		_ = file.Close()
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func extractBinary(archive, binaryName, destination string) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Name != binaryName {
			continue
		}
		if header.Typeflag != tar.TypeReg || header.Size <= 0 || header.Size > 128<<20 {
			return fmt.Errorf("release binary entry is invalid")
		}
		temporary := destination + ".partial"
		output, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o700)
		if err != nil {
			return err
		}
		if _, err = io.Copy(output, io.LimitReader(tarReader, header.Size+1)); err != nil {
			_ = output.Close()
			_ = os.Remove(temporary)
			return err
		}
		if err = output.Close(); err != nil {
			_ = os.Remove(temporary)
			return err
		}
		return os.Rename(temporary, destination)
	}
	return fmt.Errorf("release archive does not contain %s", binaryName)
}
