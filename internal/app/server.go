package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// BuildVersion is overridden by release builds through -ldflags.
var BuildVersion = "0.1.0-dev"

type AgentIngestor interface {
	AuthenticateAgent(context.Context, string) (telemetry.AgentIdentity, error)
	PersistEnvelope(context.Context, telemetry.AgentIdentity, telemetry.Envelope) (telemetry.PersistResult, error)
}

type ServerOption func(*Server)

func WithDatabase(database DatabaseHealth) ServerOption {
	return func(server *Server) { server.database = database }
}

func WithIngestor(ingestor AgentIngestor) ServerOption {
	return func(server *Server) { server.ingestor = ingestor }
}

func WithInferenceProxy(handler http.Handler) ServerOption {
	return func(server *Server) { server.inferenceProxy = handler }
}

func WithMCP(handler http.Handler) ServerOption {
	return func(server *Server) { server.mcpHandler = handler }
}

func WithControlAPI(handler http.Handler) ServerOption {
	return func(server *Server) { server.controlAPI = handler }
}

type Server struct {
	config         Config
	logger         *slog.Logger
	database       DatabaseHealth
	ingestor       AgentIngestor
	inferenceProxy http.Handler
	mcpHandler     http.Handler
	controlAPI     http.Handler
	startedAt      time.Time
	requestCount   atomic.Uint64
}

func NewServer(config Config, logger *slog.Logger, options ...ServerOption) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	server := &Server{
		config:    config,
		logger:    logger,
		startedAt: time.Now().UTC(),
	}
	for _, option := range options {
		if option != nil {
			option(server)
		}
	}
	return server
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /readyz", server.handleReady)
	mux.HandleFunc("GET /api/v1/replica", server.handleReplica)
	mux.HandleFunc("GET /api/v1/ingest/preflight", server.handleIngestPreflight)
	mux.HandleFunc("POST /api/v1/ingest", server.handleIngest)
	mux.Handle("POST /v1/chat/completions", server.proxyHandler())
	mux.Handle("/mcp", server.mcpHTTPHandler())
	mux.Handle("/api/v1/", server.controlAPIHandler())
	mux.HandleFunc("/", server.handleNotFound)
	return server.instrument(mux)
}

func (server *Server) handleHealth(writer http.ResponseWriter, request *http.Request) {
	server.writeJSON(writer, http.StatusOK, map[string]any{
		"status":       "healthy",
		"replicaId":    server.config.ReplicaID,
		"replicaRole":  server.config.ReplicaRole,
		"version":      BuildVersion,
		"startedAt":    server.startedAt,
		"observedAt":   time.Now().UTC(),
		"requestCount": server.requestCount.Load(),
	})
}

func (server *Server) handleReady(writer http.ResponseWriter, request *http.Request) {
	databaseReady := true
	if server.database != nil {
		context, cancel := context.WithTimeout(request.Context(), server.config.ReadyTimeout)
		defer cancel()
		if err := server.database.Ping(context); err != nil {
			databaseReady = false
			server.logger.Warn("runtime database readiness check failed", "error", err)
		}
	}
	status := http.StatusOK
	if !databaseReady {
		status = http.StatusServiceUnavailable
	}
	server.writeJSON(writer, status, map[string]any{
		"status":                    map[bool]string{true: "ready", false: "degraded"}[databaseReady],
		"replicaId":                 server.config.ReplicaID,
		"supabaseConfigured":        server.config.SupabaseURL != "",
		"runtimeDatabaseConfigured": server.config.RuntimeDatabaseURL != "",
		"runtimeDatabaseReady":      databaseReady,
		"tlsConfigured":             server.config.CertificatePath != "",
		"proxyConfigured":           server.inferenceProxy != nil,
		"mcpConfigured":             server.mcpHandler != nil,
		"apiConfigured":             server.controlAPI != nil,
		"observedAt":                time.Now().UTC(),
	})
}

func (server *Server) handleIngest(writer http.ResponseWriter, request *http.Request) {
	if server.ingestor == nil {
		server.writeProblem(writer, http.StatusServiceUnavailable, "ingestion is not configured")
		return
	}
	contentType := request.Header.Get("Content-Type")
	isJSON := strings.HasPrefix(contentType, "application/json")
	isProtobuf := strings.HasPrefix(contentType, "application/x-protobuf") && strings.EqualFold(request.Header.Get("Content-Encoding"), "zstd")
	if !isProtobuf && !(isJSON && server.config.AllowJSONIngest) {
		server.writeProblem(writer, http.StatusUnsupportedMediaType, "ingestion requires application/x-protobuf with zstd content encoding")
		return
	}
	bearerToken, ok := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(bearerToken) == "" {
		server.writeProblem(writer, http.StatusUnauthorized, "agent credential is required")
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, int64(telemetry.MaxUncompressedBytes))
	defer request.Body.Close()
	var envelope telemetry.Envelope
	if isProtobuf {
		compressed, err := io.ReadAll(request.Body)
		if err != nil {
			server.writeProblem(writer, http.StatusBadRequest, "invalid telemetry submission")
			return
		}
		envelope, err = telemetry.DecodeCompressedEnvelope(compressed)
		if err != nil {
			server.writeProblem(writer, http.StatusBadRequest, "invalid telemetry submission")
			return
		}
	} else {
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&envelope); err != nil {
			server.writeProblem(writer, http.StatusBadRequest, "invalid telemetry submission")
			return
		}
	}
	identity, err := server.ingestor.AuthenticateAgent(request.Context(), bearerToken)
	if err != nil {
		server.writeProblem(writer, http.StatusUnauthorized, "agent credential is invalid")
		return
	}
	receipt, err := server.ingestor.PersistEnvelope(request.Context(), identity, envelope)
	if err != nil {
		server.logger.Warn("telemetry submission rejected", "agent_id", identity.AgentID, "error", err)
		server.writeProblem(writer, http.StatusBadRequest, "telemetry submission was rejected")
		return
	}
	status := http.StatusAccepted
	if !receipt.Inserted {
		status = http.StatusOK
	}
	server.writeJSON(writer, status, map[string]any{
		"status":         map[bool]string{true: "accepted", false: "duplicate"}[receipt.Inserted],
		"idempotencyKey": envelope.IdempotencyKey(),
		"observedAt":     time.Now().UTC(),
	})
}

// handleIngestPreflight authenticates a credential without accepting,
// decoding, queueing, or persisting telemetry. It is a GET-only installation
// and replica-failover verification endpoint.
func (server *Server) handleIngestPreflight(writer http.ResponseWriter, request *http.Request) {
	if server.ingestor == nil {
		server.writeProblem(writer, http.StatusServiceUnavailable, "ingestion is not configured")
		return
	}
	bearerToken, ok := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(bearerToken) == "" {
		server.writeProblem(writer, http.StatusUnauthorized, "agent credential is required")
		return
	}
	identity, err := server.ingestor.AuthenticateAgent(request.Context(), bearerToken)
	if err != nil {
		server.writeProblem(writer, http.StatusUnauthorized, "agent credential is invalid")
		return
	}
	server.writeJSON(writer, http.StatusOK, map[string]any{
		"status":     "authenticated",
		"agentId":    identity.AgentID,
		"hostId":     identity.HostID,
		"replicaId":  server.config.ReplicaID,
		"version":    BuildVersion,
		"observedAt": time.Now().UTC(),
	})
}

func (server *Server) proxyHandler() http.Handler {
	if server.inferenceProxy != nil {
		return server.inferenceProxy
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		server.writeProblem(writer, http.StatusServiceUnavailable, "inference proxy is not configured")
	})
}

func (server *Server) mcpHTTPHandler() http.Handler {
	if server.mcpHandler != nil {
		return server.mcpHandler
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		server.writeProblem(writer, http.StatusServiceUnavailable, "MCP server is not configured")
	})
}

func (server *Server) controlAPIHandler() http.Handler {
	if server.controlAPI != nil {
		return server.controlAPI
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		server.writeProblem(writer, http.StatusServiceUnavailable, "control API is not configured")
	})
}

func (server *Server) handleReplica(writer http.ResponseWriter, request *http.Request) {
	server.writeJSON(writer, http.StatusOK, map[string]any{
		"replicaId":           server.config.ReplicaID,
		"role":                server.config.ReplicaRole,
		"version":             BuildVersion,
		"primaryConfigured":   true,
		"secondaryConfigured": true,
		"observedAt":          time.Now().UTC(),
	})
}

func (server *Server) handleNotFound(writer http.ResponseWriter, request *http.Request) {
	server.writeJSON(writer, http.StatusNotFound, map[string]any{
		"type":      "https://nodescope.dev/problems/not-found",
		"title":     "Not found",
		"status":    http.StatusNotFound,
		"requestId": request.Header.Get("X-Request-ID"),
	})
}

func (server *Server) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		server.requestCount.Add(1)
		start := time.Now()
		next.ServeHTTP(writer, request)
		server.logger.Info("http request completed",
			"method", request.Method,
			"path", request.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (server *Server) writeProblem(writer http.ResponseWriter, status int, title string) {
	writer.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"type":   "https://nodescope.dev/problems/telemetry",
		"title":  title,
		"status": status,
	})
}

func (server *Server) writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		server.logger.Error("write JSON response", "error", err)
	}
}
