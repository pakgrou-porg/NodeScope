package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TokenAuthenticator interface {
	Authenticate(context.Context, string) (Principal, error)
}

type StaticTokenAuthenticator struct{ Tokens map[string]Principal }

func (authenticator StaticTokenAuthenticator) Authenticate(_ context.Context, token string) (Principal, error) {
	principal, ok := authenticator.Tokens[token]
	if !ok || principal.ID == "" {
		return Principal{}, fmt.Errorf("MCP client credential is invalid")
	}
	return principal, nil
}

type HTTPConfiguration struct {
	Clients []HTTPClientCredential `json:"clients"`
}
type HTTPClientCredential struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Role  Role   `json:"role"`
}

func LoadHTTPConfiguration(path string) (HTTPConfiguration, error) {
	info, err := os.Stat(path)
	if err != nil {
		return HTTPConfiguration{}, fmt.Errorf("inspect MCP configuration: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return HTTPConfiguration{}, fmt.Errorf("MCP configuration must not be group or world readable")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return HTTPConfiguration{}, fmt.Errorf("read MCP configuration: %w", err)
	}
	var configuration HTTPConfiguration
	if err := json.Unmarshal(contents, &configuration); err != nil {
		return HTTPConfiguration{}, fmt.Errorf("parse MCP configuration: %w", err)
	}
	if len(configuration.Clients) == 0 {
		return HTTPConfiguration{}, fmt.Errorf("MCP configuration requires at least one client")
	}
	for _, client := range configuration.Clients {
		if client.ID == "" || client.Token == "" || (client.Role != RoleViewer && client.Role != RoleOperator && client.Role != RoleAdministrator) {
			return HTTPConfiguration{}, fmt.Errorf("MCP clients require id, token, and a valid role")
		}
	}
	return configuration, nil
}

func (configuration HTTPConfiguration) Authenticator() TokenAuthenticator {
	tokens := make(map[string]Principal, len(configuration.Clients))
	for _, client := range configuration.Clients {
		tokens[client.Token] = Principal{ID: client.ID, Role: client.Role}
	}
	return StaticTokenAuthenticator{Tokens: tokens}
}

func NewHTTPHandler(server *mcp.Server, authenticator TokenAuthenticator) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{
		Stateless: true, JSONResponse: true, MaxRequestBodyBytes: 1 << 20, SessionTimeout: 5 * time.Minute,
	})
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if authenticator == nil {
			writeUnauthorized(writer)
			return
		}
		token, ok := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			writeUnauthorized(writer)
			return
		}
		principal, err := authenticator.Authenticate(request.Context(), strings.TrimSpace(token))
		if err != nil {
			writeUnauthorized(writer)
			return
		}
		mcpHandler.ServeHTTP(writer, request.WithContext(WithPrincipal(request.Context(), principal)))
	})
}

func writeUnauthorized(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(writer).Encode(map[string]any{"type": "https://nodescope.dev/problems/mcp", "title": "MCP client credential is required", "status": http.StatusUnauthorized})
}
