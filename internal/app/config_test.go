package app

import (
	"testing"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func baseEnv() map[string]string {
	return map[string]string{
		"NODESCOPE_REPLICA_ID":         "framework",
		"NODESCOPE_REPLICA_ROLE":       "preferred",
		"NODESCOPE_PRIMARY_ENDPOINT":   "https://10.116.2.145:8443",
		"NODESCOPE_SECONDARY_ENDPOINT": "https://10.116.2.56:8443",
		"NODESCOPE_SUPABASE_URL":       "https://example.supabase.co",
	}
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig(env(baseEnv()))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.ReplicaRole != ReplicaPreferred {
		t.Fatalf("expected preferred role, got %q", config.ReplicaRole)
	}
	if config.ListenAddress != ":8080" {
		t.Fatalf("expected default listen address, got %q", config.ListenAddress)
	}
}

func TestLoadConfigRejectsIncompleteTLS(t *testing.T) {
	values := baseEnv()
	values["NODESCOPE_TLS_CERT_PATH"] = "/srv/nodescope/server.crt"
	if _, err := LoadConfig(env(values)); err == nil {
		t.Fatal("expected incomplete TLS configuration to fail")
	}
}

func TestLoadConfigRejectsNonHTTPSReplicaEndpoint(t *testing.T) {
	values := baseEnv()
	values["NODESCOPE_SECONDARY_ENDPOINT"] = "http://10.116.2.56:8080"
	if _, err := LoadConfig(env(values)); err == nil {
		t.Fatal("expected non-HTTPS replica endpoint to fail")
	}
}

func TestLoadConfigRejectsUnsafeOrDuplicateReplicaEndpoints(t *testing.T) {
	for name, mutate := range map[string]func(map[string]string){
		"primary credential": func(values map[string]string) {
			values["NODESCOPE_PRIMARY_ENDPOINT"] = "https://credential@10.116.2.145:8443"
		},
		"secondary query": func(values map[string]string) {
			values["NODESCOPE_SECONDARY_ENDPOINT"] = "https://10.116.2.56:8443?token=forbidden"
		},
		"same explicit destination": func(values map[string]string) {
			values["NODESCOPE_SECONDARY_ENDPOINT"] = "https://10.116.2.145:8443/"
		},
		"same default port destination": func(values map[string]string) {
			values["NODESCOPE_PRIMARY_ENDPOINT"] = "https://replica.nodescope.lan"
			values["NODESCOPE_SECONDARY_ENDPOINT"] = "https://replica.nodescope.lan:443"
		},
	} {
		t.Run(name, func(t *testing.T) {
			values := baseEnv()
			mutate(values)
			if _, err := LoadConfig(env(values)); err == nil {
				t.Fatal("expected unsafe or duplicate replica endpoints to fail")
			}
		})
	}
}

func TestRedactedSummaryDoesNotIncludeSupabaseURL(t *testing.T) {
	config, err := LoadConfig(env(baseEnv()))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	summary := config.RedactedSummary()
	if _, exists := summary["supabase_url"]; exists {
		t.Fatal("redacted summary must not expose Supabase URL")
	}
	if summary["supabase_configured"] != "true" {
		t.Fatal("expected configuration presence flag")
	}
}

func TestLoadConfigBuildsRuntimeDatabaseURLFromProtectedFields(t *testing.T) {
	values := baseEnv()
	values["NODESCOPE_RUNTIME_DB_HOST"] = "db.example.supabase.co"
	values["NODESCOPE_RUNTIME_DB_PASSWORD"] = "symbols & @ are encoded"
	config, err := LoadConfig(env(values))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.RuntimeDatabaseURL == "" {
		t.Fatal("expected runtime database URL to be constructed")
	}
	if config.RedactedSummary()["runtime_database_configured"] != "true" {
		t.Fatal("expected redacted summary presence flag")
	}
	if config.RedactedSummary()["runtime_database_url"] != "" {
		t.Fatal("redacted summary must not expose runtime database URL")
	}
}

func TestLoadConfigRejectsIncompleteRuntimeDatabaseFields(t *testing.T) {
	values := baseEnv()
	values["NODESCOPE_RUNTIME_DB_HOST"] = "db.example.supabase.co"
	if _, err := LoadConfig(env(values)); err == nil {
		t.Fatal("expected missing runtime password to fail")
	}
}

func TestLoadConfigRejectsAgentMTLSWithoutRequiredMaterial(t *testing.T) {
	values := baseEnv()
	values["NODESCOPE_REQUIRE_AGENT_MTLS"] = "true"
	if _, err := LoadConfig(env(values)); err == nil {
		t.Fatal("expected required agent mTLS without certificate material to fail")
	}
}
