package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

func testServer(t *testing.T, options ...ServerOption) *Server {
	t.Helper()
	config, err := LoadConfig(env(baseEnv()))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	// Unit fixtures deliberately exercise the JSON adapter; production defaults
	// to binary protobuf-zstd transport.
	config.AllowJSONIngest = true
	return NewServer(config, nil, options...)
}

func TestHealthEndpoint(t *testing.T) {
	server := testServer(t)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", response.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body["status"] != "healthy" || body["replicaId"] != "framework" {
		t.Fatalf("unexpected health payload: %#v", body)
	}
}

func TestReadyEndpoint(t *testing.T) {
	server := testServer(t)
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ready status 200, got %d", response.Code)
	}
}

func TestUnknownRouteIsNoStoreProblem(t *testing.T) {
	server := testServer(t)
	request := httptest.NewRequest(http.MethodGet, "/unexpected", nil)
	request.Header.Set("X-Request-ID", "request-test")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("expected errors to be non-cacheable")
	}
}

type fakeIngestor struct {
	identity     telemetry.AgentIdentity
	token        string
	receipt      telemetry.PersistResult
	envelope     telemetry.Envelope
	persistCalls int
}

func (ingestor *fakeIngestor) AuthenticateAgent(_ context.Context, token string) (telemetry.AgentIdentity, error) {
	if token != ingestor.token {
		return telemetry.AgentIdentity{}, fmt.Errorf("invalid token")
	}
	return ingestor.identity, nil
}

func (ingestor *fakeIngestor) PersistEnvelope(_ context.Context, identity telemetry.AgentIdentity, envelope telemetry.Envelope) (telemetry.PersistResult, error) {
	if identity != ingestor.identity {
		return telemetry.PersistResult{}, fmt.Errorf("identity mismatch")
	}
	ingestor.persistCalls++
	ingestor.envelope = envelope
	return ingestor.receipt, nil
}

func TestIngestPreflightAuthenticatesWithoutPersistingTelemetry(t *testing.T) {
	fixture := &fakeIngestor{
		identity: telemetry.AgentIdentity{AgentID: "agent-framework", HostID: "framework-host"},
		token:    "agent-secret",
	}
	server := testServer(t, WithIngestor(fixture))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ingest/preflight", nil)
	request.Header.Set("Authorization", "Bearer agent-secret")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected authenticated preflight, got %d: %s", response.Code, response.Body.String())
	}
	if fixture.persistCalls != 0 || fixture.envelope.AgentID != "" {
		t.Fatalf("preflight must not persist telemetry: %#v", fixture)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode preflight response: %v", err)
	}
	if body["status"] != "authenticated" || body["agentId"] != "agent-framework" || body["hostId"] != "framework-host" || body["replicaId"] != "framework" {
		t.Fatalf("unexpected preflight response: %#v", body)
	}
}

func validEnvelope() telemetry.Envelope {
	observedAt := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	value := 55.0
	checksum := sha256.Sum256([]byte("submission"))
	return telemetry.Envelope{
		SchemaVersion:     telemetry.CurrentSchemaVersion,
		Codec:             telemetry.CodecProtoZstd,
		AgentID:           "agent-framework",
		HostID:            "framework-host",
		BootID:            "boot-1",
		Sequence:          1,
		ObservedAt:        observedAt,
		SampleCount:       1,
		MetricValueCount:  1,
		UncompressedBytes: 128,
		CompressedBytes:   64,
		ChecksumSHA256:    fmt.Sprintf("%x", checksum),
		Samples: []telemetry.Sample{{
			DeviceID: "cpu-0",
			Metric: domain.MetricValue{
				Name:       "cpu.utilization",
				Unit:       "percent",
				Value:      &value,
				Quality:    domain.QualityFresh,
				Source:     "procfs",
				Semantics:  "host CPU utilization",
				ObservedAt: observedAt,
			},
		}},
	}
}

func TestIngestRequiresConfiguredStore(t *testing.T) {
	server := testServer(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable without ingestor, got %d", response.Code)
	}
}

func TestIngestAcceptsAuthenticatedEnvelopeWithoutLoggingContent(t *testing.T) {
	fixture := &fakeIngestor{
		identity: telemetry.AgentIdentity{AgentID: "agent-framework", HostID: "framework-host"},
		token:    "agent-secret",
		receipt:  telemetry.PersistResult{Inserted: true},
	}
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	config, err := LoadConfig(env(baseEnv()))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	config.AllowJSONIngest = true
	server := NewServer(config, logger, WithIngestor(fixture))
	body, err := json.Marshal(validEnvelope())
	if err != nil {
		t.Fatalf("encode envelope: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer agent-secret")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected accepted submission, got %d: %s", response.Code, response.Body.String())
	}
	if fixture.envelope.AgentID != "agent-framework" {
		t.Fatalf("unexpected persisted envelope %#v", fixture.envelope)
	}
	if strings.Contains(logs.String(), "agent-secret") || strings.Contains(logs.String(), "cpu.utilization") || strings.Contains(logs.String(), "host CPU utilization") {
		t.Fatalf("ingestion logs must not retain credential or metric content: %s", logs.String())
	}
}

func TestIngestAcceptsProtobufZstdWithoutJSONAdapter(t *testing.T) {
	fixture := &fakeIngestor{
		identity: telemetry.AgentIdentity{AgentID: "agent-framework", HostID: "framework-host"},
		token:    "binary-token",
		receipt:  telemetry.PersistResult{Inserted: true},
	}
	config, err := LoadConfig(env(baseEnv()))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	server := NewServer(config, nil, WithIngestor(fixture))
	body, err := telemetry.EncodeCompressedEnvelope(validEnvelope())
	if err != nil {
		t.Fatalf("encode binary envelope: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Content-Encoding", "zstd")
	request.Header.Set("Authorization", "Bearer binary-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("expected binary submission to be accepted, got %d: %s", response.Code, response.Body.String())
	}
}

func TestIngestRejectsJSONWithoutExplicitDevelopmentSwitch(t *testing.T) {
	fixture := &fakeIngestor{identity: telemetry.AgentIdentity{AgentID: "agent-framework", HostID: "framework-host"}, token: "correct-token"}
	config, err := LoadConfig(env(baseEnv()))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	server := NewServer(config, nil, WithIngestor(fixture))
	body, err := json.Marshal(validEnvelope())
	if err != nil {
		t.Fatalf("encode envelope: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer correct-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected JSON to be rejected in production mode, got %d", response.Code)
	}
}

func TestIngestRejectsBadCredentialWithoutLeakingBody(t *testing.T) {
	fixture := &fakeIngestor{identity: telemetry.AgentIdentity{AgentID: "agent-framework", HostID: "framework-host"}, token: "correct-token"}
	server := testServer(t, WithIngestor(fixture))
	body, err := json.Marshal(validEnvelope())
	if err != nil {
		t.Fatalf("encode envelope: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer incorrect-token")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized submission, got %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "cpu.utilization") || strings.Contains(response.Body.String(), "incorrect-token") {
		t.Fatalf("error response must not echo telemetry or credentials: %s", response.Body.String())
	}
}
