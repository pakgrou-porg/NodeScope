package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type Sender struct {
	client           *http.Client
	endpoints        []string
	credential       string
	mu               sync.Mutex
	circuits         map[string]endpointCircuit
	now              func() time.Time
	failureThreshold int
	cooldown         time.Duration
}

type endpointCircuit struct {
	Failures  int
	OpenUntil time.Time
}

const defaultCircuitFailureThreshold = 3
const defaultCircuitCooldown = 30 * time.Second

// PreflightResult contains only endpoint and canonical identity evidence from
// a non-mutating credential authentication request. It never contains a token
// or telemetry payload.
type PreflightResult struct {
	Endpoint  string `json:"endpoint"`
	AgentID   string `json:"agentId"`
	HostID    string `json:"hostId"`
	ReplicaID string `json:"replicaId"`
	Version   string `json:"version"`
}

func NewSender(config Config) (*Sender, error) {
	client, err := newMTLSHTTPClient(config, 15*time.Second)
	if err != nil {
		return nil, err
	}
	return &Sender{
		client:           client,
		endpoints:        []string{strings.TrimRight(config.PreferredEndpoint, "/"), strings.TrimRight(config.SecondaryEndpoint, "/")},
		credential:       config.Credential,
		circuits:         map[string]endpointCircuit{},
		now:              time.Now,
		failureThreshold: defaultCircuitFailureThreshold,
		cooldown:         defaultCircuitCooldown,
	}, nil
}

// Preflight verifies the configured agent credential without constructing,
// queueing, transmitting, or persisting telemetry. It follows the ordered
// replica policy, fails closed for credential rejection, and skips a replica
// whose transient-failure circuit is open.
func (sender *Sender) Preflight(ctx context.Context) (PreflightResult, error) {
	var failures []string
	for _, endpoint := range sender.endpoints {
		if !sender.endpointAvailable(endpoint) {
			failures = append(failures, endpoint+": circuit open")
			continue
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/v1/ingest/preflight", nil)
		if err != nil {
			return PreflightResult{}, &DeliveryError{Reason: "create preflight request", CanRetry: false}
		}
		request.Header.Set("Authorization", "Bearer "+sender.credential)
		request.Header.Set("Accept", "application/json")
		response, requestErr := sender.client.Do(request)
		if requestErr != nil {
			sender.recordTransientFailure(endpoint)
			failures = append(failures, endpoint+": transport failure")
			continue
		}
		var payload struct {
			Status    string `json:"status"`
			AgentID   string `json:"agentId"`
			HostID    string `json:"hostId"`
			ReplicaID string `json:"replicaId"`
			Version   string `json:"version"`
		}
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, 8<<10)).Decode(&payload)
		response.Body.Close()
		if response.StatusCode == http.StatusOK && decodeErr == nil && payload.Status == "authenticated" && payload.AgentID != "" && payload.HostID != "" && payload.ReplicaID != "" {
			sender.recordSuccess(endpoint)
			return PreflightResult{Endpoint: endpoint, AgentID: payload.AgentID, HostID: payload.HostID, ReplicaID: payload.ReplicaID, Version: payload.Version}, nil
		}
		if response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != http.StatusTooManyRequests {
			return PreflightResult{}, &DeliveryError{Reason: "preflight credential rejected", CanRetry: false, StatusCode: response.StatusCode}
		}
		sender.recordTransientFailure(endpoint)
		failures = append(failures, fmt.Sprintf("%s: preflight status %d", endpoint, response.StatusCode))
	}
	return PreflightResult{}, &DeliveryError{Reason: "all endpoints unavailable for preflight: " + strings.Join(failures, "; "), CanRetry: true}
}

// Send always tries the preferred endpoint first. A request body is rebuilt for
// each endpoint to preserve deterministic idempotent payloads. Client and
// authorization failures fail closed; only transient failures can be retried.
func (sender *Sender) Send(ctx context.Context, envelope telemetry.Envelope) error {
	payload, err := telemetry.EncodeCompressedEnvelope(envelope)
	if err != nil {
		return &DeliveryError{Reason: "encode envelope", CanRetry: false}
	}
	var failures []string
	for _, endpoint := range sender.endpoints {
		if !sender.endpointAvailable(endpoint) {
			failures = append(failures, endpoint+": circuit open")
			continue
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/api/v1/ingest", bytes.NewReader(payload))
		if err != nil {
			return &DeliveryError{Reason: "create ingestion request", CanRetry: false}
		}
		request.Header.Set("Content-Type", "application/x-protobuf")
		request.Header.Set("Content-Encoding", "zstd")
		request.Header.Set("Authorization", "Bearer "+sender.credential)
		response, requestErr := sender.client.Do(request)
		if requestErr != nil {
			sender.recordTransientFailure(endpoint)
			failures = append(failures, endpoint+": transport failure")
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			sender.recordSuccess(endpoint)
			return nil
		}
		if response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != http.StatusTooManyRequests {
			return &DeliveryError{Reason: "ingestion rejected", CanRetry: false, StatusCode: response.StatusCode}
		}
		sender.recordTransientFailure(endpoint)
		failures = append(failures, fmt.Sprintf("%s: status %d", endpoint, response.StatusCode))
	}
	return &DeliveryError{Reason: "all endpoints unavailable: " + strings.Join(failures, "; "), CanRetry: true}
}

func (sender *Sender) endpointAvailable(endpoint string) bool {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	state, ok := sender.circuits[endpoint]
	if !ok || state.OpenUntil.IsZero() {
		return true
	}
	return !sender.currentTime().Before(state.OpenUntil)
}

func (sender *Sender) recordTransientFailure(endpoint string) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.circuits == nil {
		sender.circuits = map[string]endpointCircuit{}
	}
	state := sender.circuits[endpoint]
	state.Failures++
	if state.Failures >= sender.circuitThreshold() {
		state.OpenUntil = sender.currentTime().Add(sender.circuitCooldown())
		state.Failures = 0
	}
	sender.circuits[endpoint] = state
}

func (sender *Sender) recordSuccess(endpoint string) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.circuits == nil {
		sender.circuits = map[string]endpointCircuit{}
	}
	sender.circuits[endpoint] = endpointCircuit{}
}

func (sender *Sender) currentTime() time.Time {
	if sender.now != nil {
		return sender.now()
	}
	return time.Now()
}

func (sender *Sender) circuitThreshold() int {
	if sender.failureThreshold > 0 {
		return sender.failureThreshold
	}
	return defaultCircuitFailureThreshold
}

func (sender *Sender) circuitCooldown() time.Duration {
	if sender.cooldown > 0 {
		return sender.cooldown
	}
	return defaultCircuitCooldown
}
