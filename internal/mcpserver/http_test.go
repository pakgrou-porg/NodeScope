package mcpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPHandlerRejectsMissingOrInvalidBearerCredentials(t *testing.T) {
	handler, _ := testHTTPHandler()
	for name, authorization := range map[string]string{"missing": "", "invalid": "Bearer invalid"} {
		t.Run(name, func(t *testing.T) {
			response := postMCP(handler, authorization, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)
			if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "MCP client credential is required") {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestHTTPHandlerAcceptsAgentZeroCompatibleInitialize(t *testing.T) {
	handler, _ := testHTTPHandler()
	response := postMCP(handler, "Bearer viewer-token", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)
	if response.Code != http.StatusOK {
		t.Fatalf("initialize status = %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Result struct {
			ServerInfo struct {
				Name string `json:"name"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode initialize response: %v", err)
	}
	if payload.Result.ServerInfo.Name != "nodescope" {
		t.Fatalf("server name = %q", payload.Result.ServerInfo.Name)
	}
}

func TestHTTPHandlerPropagatesBearerRoleToMCPTools(t *testing.T) {
	handler, service := testHTTPHandler()
	viewer := postMCP(handler, "Bearer viewer-token", `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"nodescope_set_collection_interval","arguments":{"host_id":"framework","interval_seconds":10}}}`)
	if viewer.Code != http.StatusOK || !strings.Contains(viewer.Body.String(), "operator role is required") || len(service.Actions) != 0 {
		t.Fatalf("viewer tool response = %d %s; actions = %#v", viewer.Code, viewer.Body.String(), service.Actions)
	}
	operator := postMCP(handler, "Bearer operator-token", `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"nodescope_set_collection_interval","arguments":{"host_id":"framework","interval_seconds":10}}}`)
	if operator.Code != http.StatusOK || len(service.Actions) != 1 || service.Actions[0] != "operator:interval:framework:10" {
		t.Fatalf("operator tool response = %d %s; actions = %#v", operator.Code, operator.Body.String(), service.Actions)
	}
}

func testHTTPHandler() (http.Handler, *MemoryService) {
	service := &MemoryService{Hosts: []FleetHost{{ID: "framework", Name: "Framework", Freshness: "fresh"}}}
	mcpTools := Server{Service: service}.New()
	return NewHTTPHandler(mcpTools, StaticTokenAuthenticator{Tokens: map[string]Principal{
		"viewer-token":   {ID: "viewer", Role: RoleViewer},
		"operator-token": {ID: "operator", Role: RoleOperator},
	}}), service
}

func postMCP(handler http.Handler, authorization, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestAgentZeroConfigurationExampleIsRemoteHTTPSAndSecretFree(t *testing.T) {
	path := filepath.Join("..", "..", "integrations", "agentzero", "nodescope-mcp.json.example")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read AgentZero example: %v", err)
	}
	var configuration struct {
		MCPServers map[string]struct {
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(contents, &configuration); err != nil {
		t.Fatalf("parse AgentZero example: %v", err)
	}
	nodeScope, ok := configuration.MCPServers["nodescope"]
	if !ok || !strings.HasPrefix(nodeScope.URL, "https://") || !strings.HasSuffix(nodeScope.URL, "/mcp") {
		t.Fatalf("NodeScope MCP URL = %q", nodeScope.URL)
	}
	if nodeScope.Headers["Authorization"] != "Bearer ${NODESCOPE_AGENTZERO_MCP_TOKEN}" || strings.Contains(string(contents), "Bearer ey") {
		t.Fatalf("AgentZero example must use only the environment token placeholder: %s", contents)
	}
}

func TestLoadHTTPConfigurationRejectsAmbiguousClientEntries(t *testing.T) {
	for name, contents := range map[string]string{
		"duplicate identity": `{"clients":[{"id":"agentzero","token":"token-a","role":"viewer"},{"id":"agentzero","token":"token-b","role":"operator"}]}`,
		"duplicate token":    `{"clients":[{"id":"agentzero","token":"shared-token","role":"viewer"},{"id":"operator","token":"shared-token","role":"operator"}]}`,
		"blank after trim":   `{"clients":[{"id":"  ","token":"token-a","role":"viewer"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mcp-clients.json")
			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatalf("write configuration: %v", err)
			}
			if _, err := LoadHTTPConfiguration(path); err == nil {
				t.Fatal("expected ambiguous MCP configuration to be rejected")
			}
		})
	}
}
