// Package controlapi implements NodeScope's explicit REST control surface. It
// shares the same server-side authorization and metadata-only service contract
// as MCP; it does not expose inference request or response content.
package controlapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pakgrou-porg/nodescope/internal/mcpserver"
)

type Handler struct {
	Service       mcpserver.Service
	Authenticator mcpserver.TokenAuthenticator
}

func (handler Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil || handler.Authenticator == nil {
		writeProblem(writer, http.StatusServiceUnavailable, "control API is not configured")
		return
	}
	principal, ok := handler.authorize(writer, request)
	if !ok {
		return
	}
	path := strings.TrimPrefix(request.URL.Path, "/api/v1")
	switch {
	case request.Method == http.MethodGet && path == "/fleet":
		handler.fleet(writer, request, principal)
	case request.Method == http.MethodGet && strings.HasPrefix(path, "/hosts/"):
		handler.host(writer, request, principal, strings.TrimPrefix(path, "/hosts/"))
	case request.Method == http.MethodPut && path == "/configuration/collection-interval":
		handler.setInterval(writer, request, principal)
	case request.Method == http.MethodPost && strings.HasPrefix(path, "/alerts/") && strings.HasSuffix(path, "/acknowledge"):
		alertID := strings.TrimSuffix(strings.TrimPrefix(path, "/alerts/"), "/acknowledge")
		handler.acknowledgeAlert(writer, request, principal, alertID)
	case request.Method == http.MethodPost && strings.HasPrefix(path, "/storage-baselines/") && strings.HasSuffix(path, "/refresh"):
		hostID := strings.TrimSuffix(strings.TrimPrefix(path, "/storage-baselines/"), "/refresh")
		handler.refreshBaseline(writer, request, principal, hostID)
	case request.Method == http.MethodGet && path == "/audit":
		writeProblem(writer, http.StatusNotImplemented, "audit query endpoint requires the persisted audit reader")
	default:
		writeProblem(writer, http.StatusNotFound, "control API endpoint was not found")
	}
}

func (handler Handler) authorize(writer http.ResponseWriter, request *http.Request) (mcpserver.Principal, bool) {
	token, ok := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		writeProblem(writer, http.StatusUnauthorized, "client credential is required")
		return mcpserver.Principal{}, false
	}
	principal, err := handler.Authenticator.Authenticate(request.Context(), strings.TrimSpace(token))
	if err != nil {
		writeProblem(writer, http.StatusUnauthorized, "client credential is invalid")
		return mcpserver.Principal{}, false
	}
	return principal, true
}

func require(principal mcpserver.Principal, minimum mcpserver.Role) bool {
	return principal.Role.Allows(minimum)
}
func (handler Handler) fleet(writer http.ResponseWriter, request *http.Request, principal mcpserver.Principal) {
	if !require(principal, mcpserver.RoleViewer) {
		writeProblem(writer, http.StatusForbidden, "viewer role is required")
		return
	}
	hosts, err := handler.Service.FleetStatus(request.Context(), principal)
	if err != nil {
		writeProblem(writer, http.StatusServiceUnavailable, "fleet status is unavailable")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"hosts": hosts})
}
func (handler Handler) host(writer http.ResponseWriter, request *http.Request, principal mcpserver.Principal, hostID string) {
	if !require(principal, mcpserver.RoleViewer) {
		writeProblem(writer, http.StatusForbidden, "viewer role is required")
		return
	}
	host, err := handler.Service.HostStatus(request.Context(), principal, hostID)
	if err != nil {
		writeProblem(writer, http.StatusNotFound, "host was not found")
		return
	}
	writeJSON(writer, http.StatusOK, host)
}
func (handler Handler) setInterval(writer http.ResponseWriter, request *http.Request, principal mcpserver.Principal) {
	if !require(principal, mcpserver.RoleOperator) {
		writeProblem(writer, http.StatusForbidden, "operator role is required")
		return
	}
	var input struct {
		HostID          string `json:"hostId"`
		IntervalSeconds int    `json:"intervalSeconds"`
	}
	if json.NewDecoder(http.MaxBytesReader(writer, request.Body, 16<<10)).Decode(&input) != nil || input.IntervalSeconds < 1 || input.IntervalSeconds > 60 {
		writeProblem(writer, http.StatusBadRequest, "intervalSeconds must be from 1 to 60")
		return
	}
	if err := handler.Service.SetCollectionInterval(request.Context(), principal, input.HostID, input.IntervalSeconds); err != nil {
		writeProblem(writer, http.StatusBadRequest, "collection interval change was rejected")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{"status": "accepted"})
}
func (handler Handler) acknowledgeAlert(writer http.ResponseWriter, request *http.Request, principal mcpserver.Principal, alertID string) {
	if !require(principal, mcpserver.RoleOperator) {
		writeProblem(writer, http.StatusForbidden, "operator role is required")
		return
	}
	var input struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(writer, request.Body, 8<<10)).Decode(&input)
	if err := handler.Service.AcknowledgeAlert(request.Context(), principal, alertID, input.Note); err != nil {
		writeProblem(writer, http.StatusBadRequest, "alert acknowledgement was rejected")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"status": "acknowledged"})
}
func (handler Handler) refreshBaseline(writer http.ResponseWriter, request *http.Request, principal mcpserver.Principal, hostID string) {
	if !require(principal, mcpserver.RoleOperator) {
		writeProblem(writer, http.StatusForbidden, "operator role is required")
		return
	}
	var input struct {
		AcknowledgedDiff bool `json:"acknowledgedDiff"`
	}
	if json.NewDecoder(http.MaxBytesReader(writer, request.Body, 8<<10)).Decode(&input) != nil || !input.AcknowledgedDiff {
		writeProblem(writer, http.StatusBadRequest, "acknowledgedDiff=true is required")
		return
	}
	if err := handler.Service.RefreshStorageBaseline(request.Context(), principal, hostID, true); err != nil {
		writeProblem(writer, http.StatusBadRequest, "storage baseline refresh was rejected")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{"status": "accepted"})
}
func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
func writeProblem(writer http.ResponseWriter, status int, title string) {
	writer.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{"type": "https://nodescope.dev/problems/control-api", "title": title, "status": status})
}
