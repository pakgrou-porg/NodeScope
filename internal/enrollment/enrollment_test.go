package enrollment

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCredentialProducesDigestAndNonEmptyHint(t *testing.T) {
	credential, hint, digest, err := newCredential()
	if err != nil {
		t.Fatalf("new credential: %v", err)
	}
	if len(credential) != 64 || len(hint) != 8 {
		t.Fatalf("unexpected credential or hint lengths: %d / %d", len(credential), len(hint))
	}
	if digest == [32]byte{} {
		t.Fatal("credential digest must not be empty")
	}
}

func TestWriteCredentialFileIsExclusiveAndPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets", "agent-token")
	if err := writeCredentialFile(path, "secret-value"); err != nil {
		t.Fatalf("write credential file: %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != "secret-value\n" {
		t.Fatalf("unexpected credential contents %q err=%v", string(contents), err)
	}
	if err := writeCredentialFile(path, "replacement"); err == nil {
		t.Fatal("credential output must not overwrite an existing file")
	}
}

func TestRequestRequiresFutureExpiryAndIdentityFields(t *testing.T) {
	request := Request{Slug: "framework", DisplayName: "Framework", Platform: "fedora", Address: "10.116.2.145", CredentialPath: "/tmp/token", ExpiresAt: time.Now().Add(time.Hour)}
	if err := request.Validate(); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	request.ExpiresAt = time.Now()
	if err := request.Validate(); err == nil {
		t.Fatal("expired request must be rejected")
	}
}
