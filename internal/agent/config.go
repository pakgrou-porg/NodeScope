// Package agent implements the native NodeScope telemetry service.
package agent

import (
	"fmt"
	"net"
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
	InferenceRuntimeProcesses  []string
	InferenceRuntimeEndpoints  []InferenceRuntimeEndpoint
	AlertedContainers          []string
	ContainerInventoryEnabled  bool
	ContainerInventoryProxyURL string
}

var forbiddenDatabaseConfiguration = []string{
	"NODESCOPE_AGENT_DATABASE_URL",
	"NODESCOPE_AGENT_DB_PASSWORD",
	"NODESCOPE_AGENT_SUPABASE_DB_URL",
}

func LoadConfig(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	for _, key := range forbiddenDatabaseConfiguration {
		if strings.TrimSpace(getenv(key)) != "" {
			return Config{}, fmt.Errorf("%s must not be configured; agents use authenticated HTTPS ingestion and never connect directly to PostgreSQL", key)
		}
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
		InferenceRuntimeProcesses:  splitCSV(getenv("NODESCOPE_INFERENCE_RUNTIME_PROCESS_NAMES")),
		AlertedContainers:          splitCSV(getenv("NODESCOPE_ALERT_CONTAINER_IDS_OR_NAMES")),
		ContainerInventoryEnabled:  false,
		ContainerInventoryProxyURL: strings.TrimSpace(getenv("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL")),
	}
	endpoints, err := parseInferenceRuntimeEndpoints(getenv("NODESCOPE_INFERENCE_RUNTIME_ENDPOINTS"))
	if err != nil {
		return Config{}, err
	}
	config.InferenceRuntimeEndpoints = endpoints
	if config.StateDirectory == "" {
		config.StateDirectory = defaultStateDirectory()
	}
	if !filepath.IsAbs(config.StateDirectory) {
		return Config{}, fmt.Errorf("NODESCOPE_AGENT_STATE_DIRECTORY must be an absolute path")
	}
	for label, path := range map[string]string{
		"NODESCOPE_CA_CERT_PATH":         config.CACertificatePath,
		"NODESCOPE_TLS_CLIENT_CERT_PATH": config.ClientCertificatePath,
		"NODESCOPE_TLS_CLIENT_KEY_PATH":  config.ClientPrivateKeyPath,
	} {
		if path != "" && !filepath.IsAbs(path) {
			return Config{}, fmt.Errorf("%s must be an absolute path", label)
		}
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
		if !validNodeIdentity(value) {
			return Config{}, fmt.Errorf("%s must use 1-64 lowercase letters, numbers, dots, or hyphens; it must start and end with a letter or number and must not contain consecutive dots", label)
		}
	}
	developmentMode := false
	if raw := strings.TrimSpace(getenv("NODESCOPE_DEVELOPMENT_MODE")); raw != "" {
		developmentMode, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("NODESCOPE_DEVELOPMENT_MODE must be a boolean")
		}
	}
	credential, err := loadCredential(config.CredentialFile, getenv, developmentMode)
	if err != nil {
		return Config{}, err
	}
	config.Credential = credential
	allowLoopbackReplicaEndpoints := false
	if raw := strings.TrimSpace(getenv("NODESCOPE_ALLOW_LOOPBACK_REPLICA_ENDPOINTS")); raw != "" {
		allowLoopbackReplicaEndpoints, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("NODESCOPE_ALLOW_LOOPBACK_REPLICA_ENDPOINTS must be a boolean")
		}
		if allowLoopbackReplicaEndpoints && !developmentMode {
			return Config{}, fmt.Errorf("NODESCOPE_ALLOW_LOOPBACK_REPLICA_ENDPOINTS=true requires NODESCOPE_DEVELOPMENT_MODE=true")
		}
	}
	primary, err := parseReplicaEndpoint("NODESCOPE_PRIMARY_ENDPOINT", config.PreferredEndpoint)
	if err != nil {
		return Config{}, err
	}
	secondary, err := parseReplicaEndpoint("NODESCOPE_SECONDARY_ENDPOINT", config.SecondaryEndpoint)
	if err != nil {
		return Config{}, err
	}
	if canonicalReplicaEndpoint(primary) == canonicalReplicaEndpoint(secondary) {
		return Config{}, fmt.Errorf("NODESCOPE_PRIMARY_ENDPOINT and NODESCOPE_SECONDARY_ENDPOINT must identify distinct replicas")
	}
	for label, endpoint := range map[string]*url.URL{
		"NODESCOPE_PRIMARY_ENDPOINT":   primary,
		"NODESCOPE_SECONDARY_ENDPOINT": secondary,
	} {
		if isLoopbackRuntimeHost(endpoint.Hostname()) && !allowLoopbackReplicaEndpoints {
			return Config{}, fmt.Errorf("%s must not use a loopback host unless NODESCOPE_ALLOW_LOOPBACK_REPLICA_ENDPOINTS=true and NODESCOPE_DEVELOPMENT_MODE=true", label)
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
		parsed, err := url.Parse(config.ContainerInventoryProxyURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return Config{}, fmt.Errorf("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL must be an absolute credential-free https URL without query parameters or fragments")
		}
		if proxyIP := net.ParseIP(parsed.Hostname()); proxyIP != nil && proxyIP.IsUnspecified() {
			return Config{}, fmt.Errorf("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL must not use an unspecified wildcard address")
		}
		if err := validateEndpointPort("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL", parsed); err != nil {
			return Config{}, err
		}
	}
	if config.ContainerInventoryEnabled && config.ContainerInventoryProxyURL == "" {
		return Config{}, fmt.Errorf("NODESCOPE_CONTAINER_INVENTORY_PROXY_URL is required when NODESCOPE_DOCKER_INVENTORY_ENABLED=true")
	}
	if config.ContainerInventoryEnabled && config.CACertificatePath == "" {
		return Config{}, fmt.Errorf("NODESCOPE_CA_CERT_PATH is required when NODESCOPE_DOCKER_INVENTORY_ENABLED=true")
	}
	if config.ContainerInventoryEnabled && config.ClientCertificatePath == "" {
		return Config{}, fmt.Errorf("NODESCOPE_TLS_CLIENT_CERT_PATH and NODESCOPE_TLS_CLIENT_KEY_PATH are required when NODESCOPE_DOCKER_INVENTORY_ENABLED=true")
	}
	return config, nil
}

func parseReplicaEndpoint(label, value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, fmt.Errorf("%s must be an absolute https URL", label)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("%s must not contain credentials, query parameters, or fragments", label)
	}
	if hostIP := net.ParseIP(parsed.Hostname()); hostIP != nil && hostIP.IsUnspecified() {
		return nil, fmt.Errorf("%s must not use an unspecified wildcard address", label)
	}
	if err := validateEndpointPort(label, parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func canonicalReplicaEndpoint(parsed *url.URL) string {
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	hostname := strings.TrimRight(strings.ToLower(parsed.Hostname()), ".")
	port := parsed.Port()
	if port == "" {
		port = "443"
	}
	if portNumber, err := strconv.Atoi(port); err == nil {
		port = strconv.Itoa(portNumber)
	}
	host := net.JoinHostPort(hostname, port)
	return strings.ToLower(parsed.Scheme) + "://" + host + path
}

func validateEndpointPort(label string, parsed *url.URL) error {
	rawPort := parsed.Port()
	if rawPort == "" {
		return nil
	}
	if len(rawPort) > 1 && rawPort[0] == '0' {
		return fmt.Errorf("%s must use a canonical decimal port without leading zeros", label)
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil || port > 65535 {
		return fmt.Errorf("%s must use a port from 1 through 65535", label)
	}
	if port == 0 {
		return fmt.Errorf("%s must not use port zero", label)
	}
	return nil
}

func loadCredential(path string, getenv func(string) string, developmentMode bool) (string, error) {
	legacyEnvironmentCredential := false
	if raw := strings.TrimSpace(getenv("NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return "", fmt.Errorf("NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL must be a boolean")
		}
		legacyEnvironmentCredential = parsed
	}
	if legacyEnvironmentCredential && !developmentMode {
		return "", fmt.Errorf("NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL=true requires NODESCOPE_DEVELOPMENT_MODE=true")
	}
	if path != "" {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("NODESCOPE_AGENT_CREDENTIAL_FILE must be an absolute path")
		}
		info, err := os.Lstat(path)
		if err != nil {
			return "", fmt.Errorf("lstat NodeScope agent credential file: %w", err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("NodeScope agent credential file must be a direct regular file")
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
			return "", fmt.Errorf("NodeScope agent credential file must not be group- or world-accessible")
		}
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
	// local development and test use. Production systemd units never set both.
	if legacyEnvironmentCredential && developmentMode {
		if credential := strings.TrimSpace(getenv("NODESCOPE_AGENT_CREDENTIAL")); credential != "" {
			return credential, nil
		}
	}
	return "", fmt.Errorf("NODESCOPE_AGENT_CREDENTIAL_FILE is required; environment credentials require NODESCOPE_ALLOW_LEGACY_ENV_CREDENTIAL=true and NODESCOPE_DEVELOPMENT_MODE=true")
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

// validNodeIdentity accepts the stable identifier format shared by agent and
// host identities. Keeping this deliberately narrower than display names
// prevents whitespace, path-like segments, case aliases, and delimiter-based
// injection from reaching API routes, logs, storage keys, or metric labels.
func validNodeIdentity(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	if value[0] == '.' || value[0] == '-' || value[len(value)-1] == '.' || value[len(value)-1] == '-' || strings.Contains(value, "..") {
		return false
	}
	for _, runeValue := range value {
		if runeValue >= 'a' && runeValue <= 'z' || runeValue >= '0' && runeValue <= '9' || runeValue == '.' || runeValue == '-' {
			continue
		}
		return false
	}
	return true
}

// InferenceRuntimeEndpoint identifies an administrator-approved local
// OpenAI-compatible server. It does not carry headers, credentials, models,
// or any caller-provided inference data.
type InferenceRuntimeEndpoint struct {
	ID      string
	Kind    string
	BaseURL string
}

func parseInferenceRuntimeEndpoints(raw string) ([]InferenceRuntimeEndpoint, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	seen := map[string]bool{}
	seenDestinations := map[string]bool{}
	endpoints := []InferenceRuntimeEndpoint{}
	for _, candidate := range strings.Split(raw, ";") {
		parts := strings.Split(strings.TrimSpace(candidate), "|")
		if len(parts) != 3 {
			return nil, fmt.Errorf("NODESCOPE_INFERENCE_RUNTIME_ENDPOINTS entries must use id|kind|base-url")
		}
		endpoint := InferenceRuntimeEndpoint{ID: strings.TrimSpace(parts[0]), Kind: strings.TrimSpace(parts[1]), BaseURL: strings.TrimSuffix(strings.TrimSpace(parts[2]), "/")}
		canonicalID := strings.ToLower(endpoint.ID)
		if !validRuntimeEndpointID(endpoint.ID) || seen[canonicalID] {
			return nil, fmt.Errorf("inference runtime endpoint IDs must be unique and use letters, numbers, dot, underscore, or hyphen")
		}
		if endpoint.Kind != "vllm" && endpoint.Kind != "llama_cpp" && endpoint.Kind != "lm_studio" {
			return nil, fmt.Errorf("inference runtime endpoint %q must use kind vllm, llama_cpp, or lm_studio", endpoint.ID)
		}
		parsed, err := url.Parse(endpoint.BaseURL)
		if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, fmt.Errorf("inference runtime endpoint %q must be a credential-free HTTP(S) base URL", endpoint.ID)
		}
		if runtimeIP := net.ParseIP(parsed.Hostname()); runtimeIP != nil && runtimeIP.IsUnspecified() {
			return nil, fmt.Errorf("inference runtime endpoint %q must not use an unspecified wildcard address", endpoint.ID)
		}
		if err := validateEndpointPort(fmt.Sprintf("inference runtime endpoint %q", endpoint.ID), parsed); err != nil {
			return nil, err
		}
		if parsed.Scheme == "http" && !isLoopbackRuntimeHost(parsed.Hostname()) {
			return nil, fmt.Errorf("inference runtime endpoint %q must use HTTPS unless its host is loopback", endpoint.ID)
		}
		canonicalDestination := canonicalInferenceRuntimeEndpoint(parsed)
		if seenDestinations[canonicalDestination] {
			return nil, fmt.Errorf("inference runtime endpoint %q duplicates an existing runtime destination", endpoint.ID)
		}
		seen[canonicalID] = true
		seenDestinations[canonicalDestination] = true
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func validRuntimeEndpointID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	if value[0] == '.' || value[len(value)-1] == '.' || strings.Contains(value, "..") {
		return false
	}
	for _, runeValue := range value {
		if runeValue >= 'a' && runeValue <= 'z' || runeValue >= 'A' && runeValue <= 'Z' || runeValue >= '0' && runeValue <= '9' || runeValue == '.' || runeValue == '_' || runeValue == '-' {
			continue
		}
		return false
	}
	return true
}

func canonicalInferenceRuntimeEndpoint(parsed *url.URL) string {
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	} else if parsedPort, err := strconv.Atoi(port); err == nil {
		port = strconv.Itoa(parsedPort)
	}
	hostname := strings.TrimRight(strings.ToLower(parsed.Hostname()), ".")
	return strings.ToLower(parsed.Scheme) + "://" + net.JoinHostPort(hostname, port)
}

func isLoopbackRuntimeHost(host string) bool {
	normalizedHost := strings.TrimRight(strings.TrimSpace(host), ".")
	switch strings.ToLower(normalizedHost) {
	case "localhost", "localhost.localdomain", "ip6-localhost", "ip6-loopback":
		return true
	}
	parsed := net.ParseIP(normalizedHost)
	return parsed != nil && parsed.IsLoopback()
}

func (config Config) RedactedSummary() map[string]string {
	return map[string]string{
		"agent_id_configured":                  fmt.Sprintf("%t", config.AgentID != ""),
		"host_id_configured":                   fmt.Sprintf("%t", config.HostID != ""),
		"credential_file_configured":           fmt.Sprintf("%t", config.CredentialFile != ""),
		"primary_configured":                   fmt.Sprintf("%t", config.PreferredEndpoint != ""),
		"secondary_configured":                 fmt.Sprintf("%t", config.SecondaryEndpoint != ""),
		"collection_interval_second":           fmt.Sprintf("%d", int(config.CollectionInterval.Seconds())),
		"state_directory_configured":           fmt.Sprintf("%t", config.StateDirectory != ""),
		"custom_ca_configured":                 fmt.Sprintf("%t", config.CACertificatePath != ""),
		"client_certificate_configured":        fmt.Sprintf("%t", config.ClientCertificatePath != ""),
		"client_mtls_required":                 fmt.Sprintf("%t", config.RequireClientMTLS),
		"selected_process_count":               fmt.Sprintf("%d", len(config.SelectedProcesses)),
		"inference_runtime_process_count":      fmt.Sprintf("%d", len(config.InferenceRuntimeProcesses)),
		"inference_runtime_endpoint_count":     fmt.Sprintf("%d", len(config.InferenceRuntimeEndpoints)),
		"alerted_container_count":              fmt.Sprintf("%d", len(config.AlertedContainers)),
		"docker_inventory_enabled":             fmt.Sprintf("%t", config.ContainerInventoryEnabled),
		"container_inventory_proxy_configured": fmt.Sprintf("%t", config.ContainerInventoryProxyURL != ""),
	}
}
