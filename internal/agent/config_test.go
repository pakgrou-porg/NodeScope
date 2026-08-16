package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func testAbsolutePath(name string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(`C:\nodescope-agent`, name)
	}
	return filepath.Join("/etc/nodescope-agent", name)
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig(testEnv(validEnv(t)))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.CollectionInterval.Seconds() != 5 {
		t.Fatalf("expected 5-second default, got %s", config.CollectionInterval)
	}
	summary := config.RedactedSummary()
	if summary["credential"] != "" || summary["credential_file_configured"] != "true" {
		t.Fatal("redacted summary must expose only credential-file presence")
	}
	if summary["state_directory"] != "" || summary["state_directory_configured"] != "true" || strings.Contains(fmt.Sprintf("%#v", summary), config.StateDirectory) {
		t.Fatal("redacted summary must expose only state-directory presence")
	}
}

func TestLoadConfigRejectsOutOfRangeInterval(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_COLLECTION_INTERVAL_SECONDS"] = "61"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected invalid interval to fail")
	}
}

func TestLoadConfigValidatesCanonicalAgentAndHostIDs(t *testing.T) {
	for name, mutate := range map[string]func(map[string]string){
		"agent identifier uses uppercase": func(values map[string]string) {
			values["NODESCOPE_AGENT_ID"] = "Framework-Agent"
		},
		"host identifier includes whitespace": func(values map[string]string) {
			values["NODESCOPE_HOST_ID"] = "framework host"
		},
		"agent identifier has path segment": func(values map[string]string) {
			values["NODESCOPE_AGENT_ID"] = "framework..agent"
		},
		"host identifier begins with delimiter": func(values map[string]string) {
			values["NODESCOPE_HOST_ID"] = "-framework"
		},
		"agent identifier ends with delimiter": func(values map[string]string) {
			values["NODESCOPE_AGENT_ID"] = "framework-agent."
		},
		"host identifier is too long": func(values map[string]string) {
			values["NODESCOPE_HOST_ID"] = strings.Repeat("a", 65)
		},
	} {
		t.Run(name, func(t *testing.T) {
			values := validEnv(t)
			mutate(values)
			if _, err := LoadConfig(testEnv(values)); err == nil || !strings.Contains(err.Error(), "must use 1-64 lowercase letters") {
				t.Fatalf("expected canonical ID validation failure, err=%v", err)
			}
		})
	}
}

func TestLoadConfigAcceptsCanonicalAgentAndHostIDs(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_AGENT_ID"] = "framework-agent-1"
	values["NODESCOPE_HOST_ID"] = "framework.lan"
	if _, err := LoadConfig(testEnv(values)); err != nil {
		t.Fatalf("expected canonical IDs to be accepted: %v", err)
	}
}

func TestLoadConfigRejectsRelativeStateDirectory(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_AGENT_STATE_DIRECTORY"] = "relative/nodescope-agent"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected relative agent state directory to fail")
	}
}

func TestLoadConfigRejectsRelativeCredentialFile(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_AGENT_CREDENTIAL_FILE"] = "relative/agent-token"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected relative agent credential file to fail")
	}
}

func TestLoadConfigRejectsPermissiveCredentialFileOnPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows file modes do not provide equivalent POSIX group/world permission semantics")
	}
	credentialPath := filepath.Join(t.TempDir(), "permissive-agent-token")
	if err := os.WriteFile(credentialPath, []byte("credential\n"), 0644); err != nil {
		t.Fatalf("write credential fixture: %v", err)
	}
	values := validEnv(t)
	values["NODESCOPE_AGENT_CREDENTIAL_FILE"] = credentialPath
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected group/world-readable credential file to fail")
	}
}

func TestLoadConfigRejectsRelativeTLSMaterialPaths(t *testing.T) {
	for name, key := range map[string]string{
		"CA certificate":     "NODESCOPE_CA_CERT_PATH",
		"client certificate": "NODESCOPE_TLS_CLIENT_CERT_PATH",
		"client private key": "NODESCOPE_TLS_CLIENT_KEY_PATH",
	} {
		t.Run(name, func(t *testing.T) {
			values := validEnv(t)
			values[key] = "relative/tls-material"
			if _, err := LoadConfig(testEnv(values)); err == nil {
				t.Fatalf("expected relative TLS material path for %s to fail", key)
			}
		})
	}
}

func TestLoadConfigRejectsInsecureEndpoint(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_SECONDARY_ENDPOINT"] = "http://10.116.2.56:8080"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected non-HTTPS secondary endpoint to fail")
	}
}

func TestLoadConfigRejectsUnsafeOrDuplicateReplicaEndpoints(t *testing.T) {
	for name, mutate := range map[string]func(map[string]string){
		"primary credentials": func(values map[string]string) {
			values["NODESCOPE_PRIMARY_ENDPOINT"] = "https://token@10.116.2.145:8443"
		},
		"secondary query": func(values map[string]string) {
			values["NODESCOPE_SECONDARY_ENDPOINT"] = "https://10.116.2.56:8443?credential=token"
		},
		"same endpoint":        func(values map[string]string) { values["NODESCOPE_SECONDARY_ENDPOINT"] = "https://10.116.2.145:8443" },
		"trailing slash alias": func(values map[string]string) { values["NODESCOPE_SECONDARY_ENDPOINT"] = "https://10.116.2.145:8443/" },
	} {
		t.Run(name, func(t *testing.T) {
			values := validEnv(t)
			mutate(values)
			if _, err := LoadConfig(testEnv(values)); err == nil {
				t.Fatal("expected unsafe or duplicate replica endpoint to fail")
			}
		})
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

func TestLoadConfigRejectsDirectDatabaseConfiguration(t *testing.T) {
	for _, key := range forbiddenDatabaseConfiguration {
		t.Run(key, func(t *testing.T) {
			values := validEnv(t)
			values[key] = "must-not-reach-the-agent"
			if _, err := LoadConfig(testEnv(values)); err == nil || !strings.Contains(err.Error(), "never connect directly to PostgreSQL") {
				t.Fatalf("expected %s to be rejected as a direct database setting, err=%v", key, err)
			}
		})
	}
}

func TestLoadConfigAllowsExplicitLegacyEnvironmentCredentialOnlyForDevelopment(t *testing.T) {
	values := validEnv(t)
	delete(values, "NODESCOPE_AGENT_CREDENTIAL_FILE")
	values["NODESCOPE_AGENT_CREDENTIAL"] = "legacy-secret"
	values["NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL"] = "true"
	values["NODESCOPE_DEVELOPMENT_MODE"] = "true"
	config, err := LoadConfig(testEnv(values))
	if err != nil || config.Credential != "legacy-secret" {
		t.Fatalf("expected explicit development legacy credential support, config=%#v err=%v", config.RedactedSummary(), err)
	}
}

func TestLoadConfigRejectsLegacyEnvironmentCredentialOutsideDevelopmentMode(t *testing.T) {
	values := validEnv(t)
	delete(values, "NODESCOPE_AGENT_CREDENTIAL_FILE")
	values["NODESCOPE_AGENT_CREDENTIAL"] = "legacy-secret"
	values["NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL"] = "true"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected legacy environment credential without development mode to fail")
	}
}

func TestParseInferenceRuntimeEndpointsRejectsFragments(t *testing.T) {
	_, err := parseInferenceRuntimeEndpoints("local-vllm|vllm|https://127.0.0.1:8000#diagnostic")
	if err == nil {
		t.Fatal("expected fragment-bearing inference runtime endpoint to fail")
	}
}

func TestParseInferenceRuntimeEndpointsRejectsCaseInsensitiveDuplicateIDs(t *testing.T) {
	_, err := parseInferenceRuntimeEndpoints("Local-VLLM|vllm|http://127.0.0.1:8000;local-vllm|llama_cpp|http://127.0.0.1:8080")
	if err == nil {
		t.Fatal("expected case-insensitive duplicate inference runtime endpoint IDs to fail")
	}
}

func TestParseReplicaEndpointRejectsFragments(t *testing.T) {
	_, err := parseReplicaEndpoint("NODESCOPE_PRIMARY_ENDPOINT", "https://primary.example.invalid#routing")
	if err == nil {
		t.Fatal("expected fragment-bearing ingestion replica endpoint to fail")
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
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = testAbsolutePath("inventory.crt")
	values["NODESCOPE_TLS_CLIENT_KEY_PATH"] = testAbsolutePath("inventory.key")
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
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = testAbsolutePath("inventory.crt")
	values["NODESCOPE_TLS_CLIENT_KEY_PATH"] = testAbsolutePath("inventory.key")
	if _, err := LoadConfig(testEnv(values)); err != nil {
		t.Fatalf("expected inventory proxy mTLS configuration to be accepted: %v", err)
	}
}

func TestLoadConfigRejectsUnsafeContainerInventoryProxyURL(t *testing.T) {
	for name, value := range map[string]string{
		"credentials": "https://token@inventory-proxy.lan/v1/containers",
		"query":       "https://inventory-proxy.lan/v1/containers?credential=token",
		"fragment":    "https://inventory-proxy.lan/v1/containers#alternate",
	} {
		t.Run(name, func(t *testing.T) {
			values := validEnv(t)
			values["NODESCOPE_CONTAINER_INVENTORY_PROXY_URL"] = value
			if _, err := LoadConfig(testEnv(values)); err == nil {
				t.Fatalf("expected unsafe inventory proxy URL %q to fail", value)
			}
		})
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
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = testAbsolutePath("agent.crt")
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
	values["NODESCOPE_CA_CERT_PATH"] = testAbsolutePath("ca.pem")
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected client mTLS policy without client credentials to fail")
	}
	values["NODESCOPE_TLS_CLIENT_CERT_PATH"] = testAbsolutePath("agent.crt")
	values["NODESCOPE_TLS_CLIENT_KEY_PATH"] = testAbsolutePath("agent.key")
	config, err := LoadConfig(testEnv(values))
	if err != nil || !config.RequireClientMTLS {
		t.Fatalf("expected explicit client mTLS configuration to be accepted: config=%#v err=%v", config.RedactedSummary(), err)
	}
}

func TestLoadConfigParsesInferenceRuntimeProcessNamesWithoutExposingNamesInSummary(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_INFERENCE_RUNTIME_PROCESS_NAMES"] = "vllm, llama-server, vllm, LM Studio"
	config, err := LoadConfig(testEnv(values))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := len(config.InferenceRuntimeProcesses); got != 3 || config.RedactedSummary()["inference_runtime_process_count"] != "3" || strings.Contains(fmt.Sprintf("%#v", config.RedactedSummary()), "llama-server") {
		t.Fatalf("unexpected runtime process config summary: processes=%#v summary=%#v", config.InferenceRuntimeProcesses, config.RedactedSummary())
	}
}

func TestLoadConfigParsesLocalInferenceRuntimeEndpointsWithoutExposingLocations(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_INFERENCE_RUNTIME_ENDPOINTS"] = "framework-vllm|vllm|http://127.0.0.1:8000;msi-lmstudio|lm_studio|https://msi.example.lan:1234"
	config, err := LoadConfig(testEnv(values))
	if err != nil || len(config.InferenceRuntimeEndpoints) != 2 || config.RedactedSummary()["inference_runtime_endpoint_count"] != "2" || strings.Contains(fmt.Sprintf("%#v", config.RedactedSummary()), "msi.example.lan") {
		t.Fatalf("unexpected endpoint config summary: endpoints=%#v summary=%#v err=%v", config.InferenceRuntimeEndpoints, config.RedactedSummary(), err)
	}
	values["NODESCOPE_INFERENCE_RUNTIME_ENDPOINTS"] = "unsafe|vllm|http://framework.example.lan:8000"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected non-loopback HTTP endpoint to be rejected")
	}
	values["NODESCOPE_INFERENCE_RUNTIME_ENDPOINTS"] = "unsafe|vllm|https://token@example.lan:8000"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected credential-bearing endpoint to be rejected")
	}
}

func TestLoadConfigRejectsInvalidClientMTLSPolicy(t *testing.T) {
	values := validEnv(t)
	values["NODESCOPE_REQUIRE_CLIENT_MTLS"] = "required"
	if _, err := LoadConfig(testEnv(values)); err == nil {
		t.Fatal("expected invalid client mTLS policy value to fail")
	}
}
