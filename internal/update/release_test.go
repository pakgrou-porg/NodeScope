package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeClient map[string]*http.Response

func (client fakeClient) Do(request *http.Request) (*http.Response, error) {
	response, ok := client[request.URL.String()]
	if !ok {
		return nil, fmt.Errorf("unexpected URL %s", request.URL)
	}
	return response, nil
}
func response(body []byte) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}
}
func archiveBytes(name string, content []byte) []byte {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	_ = tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o700, Size: int64(len(content)), Typeflag: tar.TypeReg})
	_, _ = tarWriter.Write(content)
	_ = tarWriter.Close()
	_ = gzipWriter.Close()
	return buffer.Bytes()
}
func TestDownloadAndStageVerifiesChecksum(t *testing.T) {
	archive := archiveBytes("nodescope-agent", []byte("binary"))
	digest := sha256.Sum256(archive)
	client := fakeClient{
		"https://example.test/release.tar.gz":        response(archive),
		"https://example.test/release.tar.gz.sha256": response([]byte(fmt.Sprintf("%x  release.tar.gz\n", digest))),
	}
	result, err := DownloadAndStage(context.Background(), client, Artifact{Version: "v1.0.0", ArchiveURL: "https://example.test/release.tar.gz", ChecksumURL: "https://example.test/release.tar.gz.sha256", BinaryName: "nodescope-agent"}, t.TempDir())
	if err != nil {
		t.Fatalf("stage verified release: %v", err)
	}
	if !strings.Contains(result.StagedBinary, "nodescope-agent.v1.0.0") {
		t.Fatalf("unexpected staged path %s", result.StagedBinary)
	}
}
func TestDownloadAndStageRejectsWrongChecksum(t *testing.T) {
	archive := archiveBytes("nodescope-agent", []byte("binary"))
	client := fakeClient{
		"https://example.test/release.tar.gz":        response(archive),
		"https://example.test/release.tar.gz.sha256": response([]byte(strings.Repeat("0", 64) + "  release.tar.gz\n")),
	}
	_, err := DownloadAndStage(context.Background(), client, Artifact{Version: "v1.0.0", ArchiveURL: "https://example.test/release.tar.gz", ChecksumURL: "https://example.test/release.tar.gz.sha256", BinaryName: "nodescope-agent"}, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}
