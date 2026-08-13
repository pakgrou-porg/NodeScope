package proxy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

type FileConfiguration struct {
	Routes  []BackendRoute     `json:"routes"`
	Clients []ClientCredential `json:"clients"`
}

type ClientCredential struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// LoadFileConfiguration reads only a root-managed private file. The client
// tokens are placed in memory for verification but are never logged, exposed in
// health output, forwarded upstream, or persisted by the proxy.
func LoadFileConfiguration(path string) (FileConfiguration, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileConfiguration{}, fmt.Errorf("inspect proxy configuration: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return FileConfiguration{}, fmt.Errorf("proxy configuration must not be group or world readable")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return FileConfiguration{}, fmt.Errorf("read proxy configuration: %w", err)
	}
	var configuration FileConfiguration
	if err := json.Unmarshal(contents, &configuration); err != nil {
		return FileConfiguration{}, fmt.Errorf("parse proxy configuration: %w", err)
	}
	if len(configuration.Routes) == 0 {
		return FileConfiguration{}, fmt.Errorf("proxy configuration requires at least one approved route")
	}
	for _, route := range configuration.Routes {
		if route.ID == "" || route.Model == "" || route.PrimaryURL == "" || !route.Enabled {
			return FileConfiguration{}, fmt.Errorf("all configured routes require id, model, primary URL, and enabled=true")
		}
		if err := validateBackendURL(route.PrimaryURL); err != nil {
			return FileConfiguration{}, fmt.Errorf("validate primary backend URL for route %q: %w", route.ID, err)
		}
		if route.SecondaryURL != "" {
			if err := validateBackendURL(route.SecondaryURL); err != nil {
				return FileConfiguration{}, fmt.Errorf("validate secondary backend URL for route %q: %w", route.ID, err)
			}
		}
	}
	if len(configuration.Clients) == 0 {
		return FileConfiguration{}, fmt.Errorf("proxy configuration requires at least one client credential")
	}
	for _, client := range configuration.Clients {
		if client.ID == "" || client.Token == "" {
			return FileConfiguration{}, fmt.Errorf("client credentials require id and token")
		}
	}
	return configuration, nil
}

func (configuration FileConfiguration) Registry() RouteRegistry {
	return NewMemoryRegistry(configuration.Routes)
}
func (configuration FileConfiguration) Authenticator() ClientAuthenticator {
	tokens := make(map[string]string, len(configuration.Clients))
	for _, client := range configuration.Clients {
		tokens[client.Token] = client.ID
	}
	return StaticClientAuthenticator{Tokens: tokens}
}

func validateBackendURL(value string) error {
	endpoint, err := url.ParseRequestURI(value)
	if err != nil || endpoint.Host == "" || endpoint.Scheme != "http" && endpoint.Scheme != "https" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return fmt.Errorf("must be a credential-free HTTP(S) base URL")
	}
	return nil
}
