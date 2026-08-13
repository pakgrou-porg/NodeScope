package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func testEnv(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func validEnv(t *testing.T) map[string]string {
	t.Helper()
	credentialPath := filepath.Join(t.TempDir(), "agent-token")
	if err := os.WriteFile(credentialPath, []byte("credential\n"), 0600); err != nil {
		t.Fatalf("write credential fixture: %v", err)
	}
	return map[string]string{
		"NODESCOPE_AGENT_ID":              "agent-id",
		"NODESCOPE_HOST_ID":               "host-id",
		"NODESCOPE_AGENT_CREDENTIAL_FILE": credentialPath,
		"NODESCOPE_PRIMARY_ENDPOINT":      "https://10.116.2.145:8443",
		"NODESCOPE_SECONDARY_ENDPOINT":    "https://10.116.2.56:8443",
	}
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig(testEnv(validEnv(t)))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.CollectionInterval.Seconds() != 5 {
		t.Fatalf("expected 5-second default, got %s", config.CollectionInterval)
	}
	if config.RedactedSummary()["credential"] != "" || config.RedactedSummary()["credential_file_configured"] != "true" {
		t.Fatal("redacted summary must expose only credential-file presence")
	}
}

func TestLoadConfigRejectsOutOfRangeInterval(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_COLLECTION_INTERVAL_SECONDS"] = "61"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected invalid interval to fail")
	}
}

func TestLoadConfigRejectsInsecureEndpoint(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_SECONDARY_ENDPOINT"] = "http://10.116.2.56:8080"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected non-HTTPS secondary endpoint to fail")
	}
}

func TestLoadConfigRejectsEnvironmentCredentialByDefault(t *testing.T) {
	values := validEnv(t)
	delete(values, "NODESCOPE_AGENT_CREDENTIAL_FILE")
	values["NODESCOPE_AGENT_CREDENTIAL"] = "legacy-secret"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected legacy environment credential to be rejected")
	}
}

func TestLoadConfigAllowsExplicitLegacyEnvironmentCredentialOnlyForDevelopment(t *testing.T) {
	values := validEnv(t)
	delete(values, "NODESCOPE_AGENT_CREDENTIAL_FILE")
	values["NODESCOPE_AGENT_CREDENTIAL"] = "legacy-secret"
	values["NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL"] = "true"
	config, err := LoadConfig(testEnv(values))
	if err != nil || config.Credential != "legacy-secret" {
		t.Fatalf("expected explicit development legacy credential support, config=%#v err=%v", config.RedactedSummary(), err)
	}
}

func TestLoadConfigKeepsDockerInventoryDisabledByDefault(t *testing.T) {
	config, err := LoadConfig(testEnv(validEnv(t)))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.ContainerInventoryEnabled {
		t.Fatal("Docker inventory must be disabled by default")
	}
}

func TestLoadConfigEnablesDockerInventoryOnlyWithExplicitBoolean(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_DOCKER_INVENTORY_ENABLED"] = "true"
	values["NODESCOPE_CONTAINER_INVENTORY_PROXY_URL"] = "https://inventory-proxy.lan/v1/containers"
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = "/etc/nodescope-agent/inventory.crt"
	values["NODESCOPE_TLS_CLIENT_KEY_PATH"] = "/etc/nodescope-agent/inventory.key"
	config, err := LoadConfig(testEnv(values))
	if err != nil || !config.ContainerInventoryEnabled {
		t.Fatalf("expected explicit Docker opt-in, config=%#v err=%v", config.RedactedSummary(), err)
	}
}

func TestLoadConfigRequiresHTTPSInventoryProxyForDockerOptIn(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_DOCKER_INVENTORY_ENABLED"] = "true"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected Docker inventory opt-in without an approved proxy URL to fail")
	}
	values["NODESCOPE_CONTAINER_INVENTORY_PROXY_URL"] = "http://inventory-proxy.lan/v1/containers"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected non-HTTPS inventory proxy URL to fail")
	}
	values["NODESCOPE_CONTAINER_INVENTORY_PROXY_URL"] = "https://inventory-proxy.lan/v1/containers"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected inventory proxy opt-in without mTLS client credentials to fail")
	}
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = "/etc/nodescope-agent/inventory.crt"
	values["NODESCOPE_TLS_CLIENT_KEY_PATH"] = "/etc/nodescope-agent/inventory.key"
	if _, err := LoadConfig(testEnv(values)); err != nil {
		t.Fatalf("expected inventory proxy mTLS configuration to be accepted: %v", err)
	}
}

func TestLoadConfigRejectsInvalidDockerInventoryBoolean(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_DOCKER_INVENTORY_ENABLED"] = "approved"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected invalid Docker opt-in value to fail")
	}
}

func TestLoadConfigRejectsIncompleteClientCertificateConfiguration(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = "/etc/nodescope-agent/agent.crt"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected incomplete client certificate configuration to fail")
	}
}

func TestLoadConfigRequiresInternalCAAndClientCertificateForExplicitMTLS(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_REQUIRE_CLIENT_MTLS"] = "true"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected client mTLS policy without internal CA and credentials to fail")
	}
	values["NODESCOPE_CA_CERT_PATH"] = "/etc/nodescope-agent/ca.pem"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected client mTLS policy without client credentials to fail")
	}
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = "/etc/nodescope-agent/agent.crt"
	values["NODESCOPE_TLS_CLIENT_KEY_PATH"] = "/etc/nodescope-agent/agent.key"
	config, err := LoadConfig(testEnv(values))
	if err != nil || !config.RequireClientMTLS {
		t.Fatalf("expected explicit client mTLS configuration to be accepted: config=%#v err=%v", config.RedactedSummary(), err)
	}
}

func TestLoadConfigRejectsInvalidClientMTLSPolicy(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_REQUIRE_CLIENT_MTLS"] = "required"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected invalid client mTLS policy value to fail")
	}
}
