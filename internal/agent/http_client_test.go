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
