// Package agent implements the native NodeScope telemetry service.
package agent

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AgentID                    string
	HostID                     string
	Credential                 string
	CredentialFile             string
	PreferredEndpoint          string
	SecondaryEndpoint          string
	CollectionInterval         time.Duration
	StateDirectory             string
	CACertificatePath          string
	ClientCertificatePath      string
	ClientPrivateKeyPath       string
	RequireClientMTLS          bool
	SelectedProcesses          []string
	AlertedContainers          []string
	ContainerInventoryEnabled  bool
	ContainerInventoryProxyURL string
}

func LoadConfig(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	config := Config{
		AgentID:                    strings.TrimSpace(getenv("NODESCOPE_AGENT_ID")),
		HostID:                     strings.TrimSpace(getenv("NODESCOPE_HOST_ID")),
		CredentialFile:             strings.TrimSpace(getenv("NODESCOPE_AGENT_CREDENTIAL_FILE")),
		PreferredEndpoint:          strings.TrimSpace(getenv("NODESCOPE_PRIMARY_ENDPOINT")),
		SecondaryEndpoint:          strings.TrimSpace(getenv("NODESCOPE_SECONDARY_ENDPOINT")),
		StateDirectory:             strings.TrimSpace(getenv("NODESCOPE_AGENT_STATE_DIRECTORY")),
		CACertificatePath:          strings.TrimSpace(getenv("NODESCOPE_CA_CERT_PATH")),
		ClientCertificatePath:      strings.TrimSpace(getenv("NODESCOPE_TLS_CLIENT_CERT_PATH")),
		ClientPrivateKeyPath:       strings.TrimSpace(getenv("NODESCOPE_TLS_CLIENT_KEY_PATH")),
		RequireClientMTLS:          false,
		SelectedProcesses:          splitCSV(getenv("NODESCOPE_SELECTED_PROCESS_NAMES")),
		AlertedContainers:          splitCSV(getenv("NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES")),
		ContainerInventoryEnabled:  false,
		ContainerInventoryProxyURL: strings.TrimSpace(getenv("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL")),
	}
	if config.StateDirectory == "" {
		config.StateDirectory = defaultStateDirectory()
	}
	if (config.ClientCertificatePath == "") != (config.ClientPrivateKeyPath == "") {
		return Config{}, fmt.Errorf("NODESCOPE_TLS_CLIENT_CERT_PATH and NODESCOPE_TLS_CLIENT_KEY_PATH must be set together")
	}
	if raw := strings.TrimSpace(getenv("NODESCOPE_REQUIRE_CLIENT_MTLS")); raw != "" {
		required, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("NODESCOPE_REQUIRE_CLIENT_MTLS must be a boolean")
		}
		config.RequireClientMTLS = required
	}
	if config.RequireClientMTLS {
		if config.CACertificatePath == "" {
			return Config{}, fmt.Errorf("NODESCOPE_CA_CERT_PATH is required when NODESCOPE_REQUIRE_CLIENT_MTLS=true")
		}
		if config.ClientCertificatePath == "" {
			return Config{}, fmt.Errorf("NODESCOPE_TLS_CLIENT_CERT_PATH and NODESCOPE_TLS_CLIENT_KEY_PATH are required when NODESCOPE_REQUIRE_CLIENT_MTLS=true")
		}
	}
	for label, value := range map[string]string{
		"NODESCOPE_AGENT_ID": config.AgentID,
		"NODESCOPE_HOST_ID":  config.HostID,
	} {
		if value == "" {
			return Config{}, fmt.Errorf("%s is required", label)
		}
	}
	credential, err := loadCredential(config.CredentialFile, getenv)
	if err != nil {
		return Config{}, err
	}
	config.Credential = credential
	for label, endpoint := range map[string]string{
		"NODESCOPE_PRIMARY_ENDPOINT":   config.PreferredEndpoint,
		"NODESCOPE_SECONDARY_ENDPOINT": config.SecondaryEndpoint,
	} {
		parsed, err := url.ParseRequestURI(endpoint)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return Config{}, fmt.Errorf("%s must be an absolute https URL", label)
		}
	}
	intervalSeconds := 5
	if raw := strings.TrimSpace(getenv("NODESCOPE_COLLECTION_INTERVAL_SECONDS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 60 {
			return Config{}, fmt.Errorf("NODESCOPE_COLLECTION_INTERVAL_SECONDS must be an integer from 1 to 60")
		}
		intervalSeconds = parsed
	}
	config.CollectionInterval = time.Duration(intervalSeconds) * time.Second
	if raw := strings.TrimSpace(getenv("NODESCOPE_DOCKER_INVENTORY_ENABLED")); raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("NODESCOPE_DOCKER_INVENTORY_ENABLED must be a boolean")
		}
		config.ContainerInventoryEnabled = enabled
	}
	if config.ContainerInventoryProxyURL != "" {
		parsed, err := url.ParseRequestURI(config.ContainerInventoryProxyURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return Config{}, fmt.Errorf("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL must be an absolute https URL")
		}
	}
	if config.ContainerInventoryEnabled && config.ContainerInventoryProxyURL == "" {
		return Config{}, fmt.Errorf("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL is required when NODESCOPE_DOCKER_INVENTORY_ENABLED=true")
	}
	if config.ContainerInventoryEnabled && config.ClientCertificatePath == "" {
		return Config{}, fmt.Errorf("NODESCOPE_TLS_CLIENT_CERT_PATH and NODESCOPE_TLS_CLIENT_KEY_PATH are required when NODESCOPE_DOCKER_INVENTORY_ENABLED=true")
	}
	return config, nil
}

func loadCredential(path string, getenv func(string) string) (string, error) {
	if path != "" {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read NodeScope agent credential file: %w", err)
		}
		credential := strings.TrimSpace(string(contents))
		if credential == "" {
			return "", fmt.Errorf("NodeScope agent credential file is empty")
		}
		return credential, nil
	}
	// A legacy environment credential is permitted only for explicitly marked,
	// local development and test use. Production systemd units never set this.
	if getenv("NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL") == "true" {
		if credential := strings.TrimSpace(getenv("NODESCOPE_AGENT_CREDENTIAL")); credential != "" {
			return credential, nil
		}
	}
	return "", fmt.Errorf("NODESCOPE_AGENT_CREDENTIAL_FILE is required; environment credentials require NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL=true")
}

func defaultStateDirectory() string {
	if runtime.GOOS != "windows" {
		return "/var/lib/nodescope-agent"
	}
	programData := strings.TrimSpace(os.Getenv("ProgramData"))
	if programData == "" {
		programData = `C:\\ProgramData`
	}
	return filepath.Join(programData, "NodeScope", "state")
}

func splitCSV(raw string) []string {
	seen := map[string]bool{}
	values := []string{}
	for _, candidate := range strings.Split(raw, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		values = append(values, candidate)
	}
	return values
}

func (config Config) RedactedSummary() map[string]string {
	return map[string]string{
		"agent_id_configured":                  fmt.Sprintf("%t", config.AgentID != ""),
		"host_id_configured":                   fmt.Sprintf("%t", config.HostID != ""),
		"credential_file_configured":           fmt.Sprintf("%t", config.CredentialFile != ""),
		"primary_configured":                   fmt.Sprintf("%t", config.PreferredEndpoint != ""),
		"secondary_configured":                 fmt.Sprintf("%t", config.SecondaryEndpoint != ""),
		"collection_interval_second":           fmt.Sprintf("%d", int(config.CollectionInterval.Seconds())),
		"state_directory":                      config.StateDirectory,
		"custom_ca_configured":                 fmt.Sprintf("%t", config.CACertificatePath != ""),
		"client_certificate_configured":        fmt.Sprintf("%t", config.ClientCertificatePath != ""),
		"client_mtls_required":                 fmt.Sprintf("%t", config.RequireClientMTLS),
		"selected_process_count":               fmt.Sprintf("%d", len(config.SelectedProcesses)),
		"alerted_container_count":              fmt.Sprintf("%d", len(config.AlertedContainers)),
		"docker_inventory_enabled":             fmt.Sprintf("%t", config.ContainerInventoryEnabled),
		"container_inventory_proxy_configured": fmt.Sprintf("%t", config.ContainerInventoryProxyURL != ""),
	}
}
