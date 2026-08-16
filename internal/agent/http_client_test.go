package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNewMTLSHTTPClientRejectsPermissivePrivateKeyOnPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows file modes do not provide equivalent POSIX group/world permission semantics")
	}
	privateKeyPath := filepath.Join(t.TempDir(), "agent.key")
	if err := os.WriteFile(privateKeyPath, []byte("not-a-real-key\n"), 0644); err != nil {
		t.Fatalf("write private key fixture: %v", err)
	}
	_, err := newMTLSHTTPClient(Config{
		ClientCertificatePath: testAbsolutePath("agent.crt"),
		ClientPrivateKeyPath:  privateKeyPath,
	}, time.Second)
	if err == nil || !strings.Contains(err.Error(), "must not be group- or world-accessible") {
		t.Fatalf("expected private-key permission rejection, err=%v", err)
	}
}

func TestRequirePrivateTLSKeyRejectsNonRegularFile(t *testing.T) {
	if err := requirePrivateTLSKey(t.TempDir()); err == nil || !strings.Contains(err.Error(), "must be a direct regular file") {
		t.Fatalf("expected non-regular private key to fail, err=%v", err)
	}
}

func TestRequirePrivateTLSKeyRejectsSymlinkOnPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows symlink creation depends on host-specific privilege configuration")
	}
	target := filepath.Join(t.TempDir(), "agent.key")
	if err := os.WriteFile(target, []byte("not-a-real-key\n"), 0600); err != nil {
		t.Fatalf("write private-key target: %v", err)
	}
	link := filepath.Join(t.TempDir(), "agent-link.key")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("create private-key symlink: %v", err)
	}
	if err := requirePrivateTLSKey(link); err == nil || !strings.Contains(err.Error(), "must be a direct regular file") {
		t.Fatalf("expected symlinked private key to fail, err=%v", err)
	}
}

func TestRequireCACertificateFileRejectsNonRegularFile(t *testing.T) {
	if err := requireCACertificateFile(t.TempDir()); err == nil || !strings.Contains(err.Error(), "must be a direct regular file") {
		t.Fatalf("expected non-regular CA certificate file to fail, err=%v", err)
	}
}

func TestRequireCACertificateFileRejectsSymlinkOnPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows symlink creation depends on host-specific privilege configuration")
	}
	target := filepath.Join(t.TempDir(), "root-ca.pem")
	if err := os.WriteFile(target, []byte("not-a-real-certificate\n"), 0644); err != nil {
		t.Fatalf("write CA certificate target: %v", err)
	}
	link := filepath.Join(t.TempDir(), "root-ca-link.pem")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("create CA certificate symlink: %v", err)
	}
	if err := requireCACertificateFile(link); err == nil || !strings.Contains(err.Error(), "must be a direct regular file") {
		t.Fatalf("expected symlinked CA certificate file to fail, err=%v", err)
	}
}
