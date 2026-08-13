package consoleclient

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFleetHTTPSUsesCredentialFileAndCustomCA(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/fleet" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer verifier-token" {
			t.Fatalf("authorization = %q", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"hosts":[{"id":"framework","name":"Framework","platform":"linux/amd64","freshness":"fresh","metric_count":12}]}`))
	}))
	defer server.Close()

	directory := t.TempDir()
	credentialPath := filepath.Join(directory, "credential")
	if err := os.WriteFile(credentialPath, []byte(" verifier-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	caPath := filepath.Join(directory, "ca.pem")
	if err := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600); err != nil {
		t.Fatal(err)
	}

	hosts, err := LoadFleet(context.Background(), Config{Endpoint: server.URL, CredentialFile: credentialPath, CAFile: caPath})
	if err != nil {
		t.Fatalf("LoadFleet() error = %v", err)
	}
	if len(hosts) != 1 || hosts[0].HostSlug != "framework" || hosts[0].DisplayName != "Framework" || hosts[0].CurrentMetricCount != 12 {
		t.Fatalf("hosts = %#v", hosts)
	}
}

func TestValidateRejectsInsecureOrCredentialBearingHTTPSConfiguration(t *testing.T) {
	for name, config := range map[string]Config{
		"http":       {Endpoint: "http://nodescope.lan", CredentialFile: "/run/credential"},
		"basic-auth": {Endpoint: "https://operator@example.test", CredentialFile: "/run/credential"},
		"query":      {Endpoint: "https://nodescope.lan?token=not-allowed", CredentialFile: "/run/credential"},
		"missing":    {Endpoint: "https://nodescope.lan"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := config.Validate(); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func TestLoadFleetSSHUsesReadOnlyJSONCommand(t *testing.T) {
	previous := executeSSH
	t.Cleanup(func() { executeSSH = previous })
	executeSSH = func(_ context.Context, target, command string) ([]byte, error) {
		if target != "framework.lan" {
			t.Fatalf("target = %q", target)
		}
		if command != "nodescope-cli --format json" {
			t.Fatalf("command = %q", command)
		}
		return []byte(`[{"host_slug":"framework","display_name":"Framework","freshness_state":"fresh"}]`), nil
	}

	hosts, err := LoadFleet(context.Background(), Config{SSHTarget: "framework.lan"})
	if err != nil {
		t.Fatalf("LoadFleet() error = %v", err)
	}
	if len(hosts) != 1 || hosts[0].FreshnessState != "fresh" {
		t.Fatalf("hosts = %#v", hosts)
	}
}

func TestReadCredentialRejectsOversizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", maxCredentialBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readCredential(path); err == nil {
		t.Fatal("readCredential() error = nil")
	}
}
