package controlapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pakgrou-porg/nodescope/internal/mcpserver"
)

func testHandler() Handler {
	return Handler{
		Service: &mcpserver.MemoryService{Hosts: []mcpserver.FleetHost{{ID: "framework", Name: "Framework", Freshness: "fresh"}}},
		Authenticator: mcpserver.StaticTokenAuthenticator{Tokens: map[string]mcpserver.Principal{
			"viewer-token":   {ID: "viewer", Role: mcpserver.RoleViewer},
			"operator-token": {ID: "operator", Role: mcpserver.RoleOperator},
		}},
	}
}
func TestFleetAllowsViewer(t *testing.T) {
	handler := testHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/fleet", nil)
	request.Header.Set("Authorization", "Bearer viewer-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Framework") {
		t.Fatalf("unexpected fleet response %d %s", response.Code, response.Body.String())
	}
}
func TestMutationDeniesViewer(t *testing.T) {
	handler := testHandler()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/configuration/collection-interval", strings.NewReader(`{"intervalSeconds":10}`))
	request.Header.Set("Authorization", "Bearer viewer-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected viewer mutation denial, got %d", response.Code)
	}
}
func TestMissingCredentialIsUnauthorized(t *testing.T) {
	handler := testHandler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/fleet", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}
