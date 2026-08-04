// Package app contains the standalone NodeScope server application wiring.
package app

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ReplicaRole string

const (
	ReplicaPreferred ReplicaRole = "preferred"
	ReplicaSecondary ReplicaRole = "secondary"
)

type Config struct {
	ReplicaID          string
	ReplicaRole        ReplicaRole
	ListenAddress      string
	SupabaseURL        string
	RuntimeDatabaseURL string
	AllowJSONIngest    bool
	PrimaryEndpoint    string
	SecondaryEndpoint  string
	CertificatePath    string
	PrivateKeyPath     string
	AgentClientCAPath  string
	RequireAgentMTLS   bool
	ProxyConfigPath    string
	MCPConfigPath      string
	APIConfigPath      string
	ReadyTimeout       time.Duration
}

func LoadConfig(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	config := Config{
		ReplicaID:          strings.TrimSpace(getenv("NODESCOPE_REPLICA_ID")),
		ReplicaRole:        ReplicaRole(strings.TrimSpace(getenv("NODESCOPE_REPLICA_ROLE"))),
		ListenAddress:      strings.TrimSpace(getenv("NODESCOPE_LISTEN_ADDRESS")),
		SupabaseURL:        strings.TrimSpace(getenv("NODESCOPE_SUPABASE_URL")),
		RuntimeDatabaseURL: strings.TrimSpace(getenv("NODESCOPE_RUNTIME_DATABASE_URL")),
		AllowJSONIngest:    strings.EqualFold(strings.TrimSpace(getenv("NODESCOPE_ALLOW_JSON_INGEST")), "true"),
		PrimaryEndpoint:    strings.TrimSpace(getenv("NODESCOPE_PRIMARY_ENDPOINT")),
		SecondaryEndpoint:  strings.TrimSpace(getenv("NODESCOPE_SECONDARY_ENDPOINT")),
		CertificatePath:    strings.TrimSpace(getenv("NODESCOPE_TLS_CERT_PATH")),
		PrivateKeyPath:     strings.TrimSpace(getenv("NODESCOPE_TLS_KEY_PATH")),
		AgentClientCAPath:  strings.TrimSpace(getenv("NODESCOPE_AGENT_CLIENT_CA_CERT_PATH")),
		RequireAgentMTLS:   strings.EqualFold(strings.TrimSpace(getenv("NODESCOPE_REQUIRE_AGENT_MTLS")), "true"),
		ProxyConfigPath:    strings.TrimSpace(getenv("NODESCOPE_PROXY_CONFIG_PATH")),
		MCPConfigPath:      strings.TrimSpace(getenv("NODESCOPE_MCP_CONFIG_PATH")),
		APIConfigPath:      strings.TrimSpace(getenv("NODESCOPE_API_CONFIG_PATH")),
		ReadyTimeout:       5 * time.Second,
	}
	if config.ReplicaID == "" {
		return Config{}, fmt.Errorf("NODESCOPE_REPLICA_ID is required")
	}
	if config.ReplicaRole != ReplicaPreferred && config.ReplicaRole != ReplicaSecondary {
		return Config{}, fmt.Errorf("NODESCOPE_REPLICA_ROLE must be preferred or secondary")
	}
	if config.ListenAddress == "" {
		config.ListenAddress = ":8080"
	}
	for name, raw := range map[string]string{
		"NODESCOPE_PRIMARY_ENDPOINT":   config.PrimaryEndpoint,
		"NODESCOPE_SECONDARY_ENDPOINT": config.SecondaryEndpoint,
	} {
		if raw == "" {
			return Config{}, fmt.Errorf("%s is required", name)
		}
		parsed, err := url.ParseRequestURI(raw)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return Config{}, fmt.Errorf("%s must be an absolute https URL", name)
		}
	}
	if config.SupabaseURL != "" {
		parsed, err := url.ParseRequestURI(config.SupabaseURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return Config{}, fmt.Errorf("NODESCOPE_SUPABASE_URL must be an absolute https URL when configured")
		}
	}
	if config.RuntimeDatabaseURL == "" && strings.TrimSpace(getenv("NODESCOPE_RUNTIME_DB_HOST")) != "" {
		runtimeURL, err := buildRuntimeDatabaseURL(getenv)
		if err != nil {
			return Config{}, err
		}
		config.RuntimeDatabaseURL = runtimeURL
	}
	if config.RuntimeDatabaseURL != "" {
		parsed, err := url.Parse(config.RuntimeDatabaseURL)
		if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" || parsed.User == nil {
			return Config{}, fmt.Errorf("NODESCOPE_RUNTIME_DATABASE_URL must be a PostgreSQL URL with credentials when configured")
		}
	}
	if (config.CertificatePath == "") != (config.PrivateKeyPath == "") {
		return Config{}, fmt.Errorf("NODESCOPE_TLS_CERT_PATH and NODESCOPE_TLS_KEY_PATH must be set together")
	}
	if config.RequireAgentMTLS {
		if config.CertificatePath == "" || config.AgentClientCAPath == "" {
			return Config{}, fmt.Errorf("NODESCOPE_REQUIRE_AGENT_MTLS requires server TLS certificate/key and NODESCOPE_AGENT_CLIENT_CA_CERT_PATH")
		}
	}
	return config, nil
}

// RedactedSummary is safe to log. It deliberately omits secrets and never
// serializes the Supabase URL because deployment logs are public diagnostics.
func buildRuntimeDatabaseURL(getenv func(string) string) (string, error) {
	host := strings.TrimSpace(getenv("NODESCOPE_RUNTIME_DB_HOST"))
	password := getenv("NODESCOPE_RUNTIME_DB_PASSWORD")
	if host == "" || password == "" {
		return "", fmt.Errorf("NODESCOPE_RUNTIME_DB_HOST and NODESCOPE_RUNTIME_DB_PASSWORD are required together")
	}
	port := strings.TrimSpace(getenv("NODESCOPE_RUNTIME_DB_PORT"))
	if port == "" {
		port = "5432"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("NODESCOPE_RUNTIME_DB_PORT must be numeric")
	}
	user := strings.TrimSpace(getenv("NODESCOPE_RUNTIME_DB_USER"))
	if user == "" {
		user = "nodescope_runtime_login"
	}
	databaseName := strings.TrimSpace(getenv("NODESCOPE_RUNTIME_DB_NAME"))
	if databaseName == "" {
		databaseName = "postgres"
	}
	return (&url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + databaseName,
		RawQuery: "sslmode=require",
	}).String(), nil
}

func (c Config) RedactedSummary() map[string]string {
	return map[string]string{
		"replica_id":                  c.ReplicaID,
		"replica_role":                string(c.ReplicaRole),
		"listen_address":              c.ListenAddress,
		"primary_configured":          fmt.Sprintf("%t", c.PrimaryEndpoint != ""),
		"secondary_configured":        fmt.Sprintf("%t", c.SecondaryEndpoint != ""),
		"supabase_configured":         fmt.Sprintf("%t", c.SupabaseURL != ""),
		"runtime_database_configured": fmt.Sprintf("%t", c.RuntimeDatabaseURL != ""),
		"json_ingest_enabled":         fmt.Sprintf("%t", c.AllowJSONIngest),
		"tls_configured":              fmt.Sprintf("%t", c.CertificatePath != ""),
		"agent_mtls_required":         fmt.Sprintf("%t", c.RequireAgentMTLS),
		"proxy_configured":            fmt.Sprintf("%t", c.ProxyConfigPath != ""),
		"mcp_configured":              fmt.Sprintf("%t", c.MCPConfigPath != ""),
		"api_configured":              fmt.Sprintf("%t", c.APIConfigPath != ""),
	}
}
