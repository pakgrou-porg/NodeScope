package proxy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileConfigurationRejectsInsecurePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proxy.json")
	contents := []byte(`{"routes":[{"id":"r","model":"m","primary_url":"https://backend","enabled":true}],"clients":[{"id":"cli","token":"secret"}]}`)
	if err := os.WriteFile(path, contents, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFileConfiguration(path); err == nil {
		t.Fatal("expected world-readable proxy configuration to be rejected")
	}
}

func TestLoadFileConfigurationBuildsOnlyRouteAndClientContracts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proxy.json")
	contents := []byte(`{"routes":[{"id":"r","model":"m","primary_url":"https://backend","enabled":true}],"clients":[{"id":"cli","token":"secret"}]}`)
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
	configuration, err := LoadFileConfiguration(path)
	if err != nil {
		t.Fatalf("load private config: %v", err)
	}
	if _, err := configuration.Registry().Resolve(context.Background(), "m"); err != nil {
		t.Fatalf("resolve route: %v", err)
	}
	if client, err := configuration.Authenticator().AuthenticateClient(context.Background(), "secret"); err != nil || client != "cli" {
		t.Fatalf("authenticate configured client: %s %v", client, err)
	}
}
