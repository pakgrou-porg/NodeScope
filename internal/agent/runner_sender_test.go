package agent

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type recordingSender struct{ envelope telemetry.Envelope }

func (sender *recordingSender) Send(_ context.Context, envelope telemetry.Envelope) error {
	sender.envelope = envelope
	return nil
}

type fakeCollector struct {
	name    string
	samples []telemetry.Sample
}

func (collector fakeCollector) Name() string { return collector.name }
func (collector fakeCollector) Collect(_ context.Context, _ time.Time) ([]telemetry.Sample, error) {
	return collector.samples, nil
}

type fakeInventoryCollector struct{ fakeCollector }

func (collector fakeInventoryCollector) CollectContainerInventory(_ context.Context, _ time.Time) ([]telemetry.Sample, []telemetry.ContainerInventory, error) {
	return collector.samples, []telemetry.ContainerInventory{{ContainerID: "abc", Name: "vllm", Image: "vllm:latest", State: "running", Health: "unreported"}}, nil
}

func TestRunnerIncludesContainerInventory(t *testing.T) {
	directory := t.TempDir()
	state, err := OpenSequenceStore(directory)
	if err != nil {
		t.Fatalf("open state: %v", err)
	}
	value := 1.0
	sender := &recordingSender{}
	runner, err := NewRunner(Config{AgentID: "agent", HostID: "host"}, []Collector{fakeInventoryCollector{fakeCollector{name: "docker", samples: []telemetry.Sample{{DeviceID: "docker", Metric: domain.MetricValue{Name: "container.inventory.available", Unit: "state", Value: &value, Quality: domain.QualityFresh, Source: "test", Semantics: "test", ObservedAt: time.Now()}}}}}}, sender, state)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if err := runner.CollectOnce(context.Background()); err != nil {
		t.Fatalf("collect once: %v", err)
	}
	if len(sender.envelope.Containers) != 1 || sender.envelope.Containers[0].Name != "vllm" {
		t.Fatalf("unexpected containers %#v", sender.envelope.Containers)
	}
}

type scriptedRoundTripper struct {
	mu       sync.Mutex
	statuses []int
	calls    []string
}

func (roundTripper *scriptedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	roundTripper.mu.Lock()
	defer roundTripper.mu.Unlock()
	roundTripper.calls = append(roundTripper.calls, request.URL.Host)
	status := roundTripper.statuses[len(roundTripper.calls)-1]
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header), Request: request}, nil
}

type preflightRoundTripper struct {
	mu       sync.Mutex
	statuses []int
	bodies   []string
	calls    []string
}

func (roundTripper *preflightRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	roundTripper.mu.Lock()
	defer roundTripper.mu.Unlock()
	roundTripper.calls = append(roundTripper.calls, request.URL.Host)
	index := len(roundTripper.calls) - 1
	return &http.Response{StatusCode: roundTripper.statuses[index], Body: io.NopCloser(strings.NewReader(roundTripper.bodies[index])), Header: make(http.Header), Request: request}, nil
}

func validAgentEnvelope() telemetry.Envelope {
	observedAt := time.Date(2026, 7, 22, 12, 30, 0, 0, time.UTC)
	value := 23.0
	return telemetry.Envelope{SchemaVersion: telemetry.CurrentSchemaVersion, Codec: telemetry.CodecProtoZstd, AgentID: "agent", HostID: "host", BootID: "boot", Sequence: 1, ObservedAt: observedAt, SampleCount: 1, MetricValueCount: 1, UncompressedBytes: 1, CompressedBytes: 1, ChecksumSHA256: strings.Repeat("0", 64), Samples: []telemetry.Sample{{DeviceID: "cpu-0", Metric: domain.MetricValue{Name: "cpu.utilization", Unit: "percent", Value: &value, Quality: domain.QualityFresh, Source: "test", Semantics: "test", ObservedAt: observedAt}}}}
}

func TestSenderFallsBackOnlyAfterServerFailure(t *testing.T) {
	roundTripper := &scriptedRoundTripper{statuses: []int{503, 202}}
	sender := &Sender{client: &http.Client{Transport: roundTripper}, endpoints: []string{"https://framework.test", "https://asus.test"}, credential: "token"}
	if err := sender.Send(context.Background(), validAgentEnvelope()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if strings.Join(roundTripper.calls, ",") != "framework.test,asus.test" {
		t.Fatalf("unexpected fallback path %#v", roundTripper.calls)
	}
}

func TestSenderDoesNotFailOverOnAuthorizationFailure(t *testing.T) {
	roundTripper := &scriptedRoundTripper{statuses: []int{401}}
	sender := &Sender{client: &http.Client{Transport: roundTripper}, endpoints: []string{"https://framework.test", "https://asus.test"}, credential: "token"}
	if err := sender.Send(context.Background(), validAgentEnvelope()); err == nil {
		t.Fatal("expected authorization failure")
	}
	if len(roundTripper.calls) != 1 {
		t.Fatalf("authorization failure must not fail over: %#v", roundTripper.calls)
	}
}

func TestSenderPreflightFailsOverOnlyAfterTransientFailure(t *testing.T) {
	roundTripper := &preflightRoundTripper{
		statuses: []int{http.StatusServiceUnavailable, http.StatusOK},
		bodies:   []string{"{}", `{"status":"authenticated","agentId":"agent","hostId":"host","replicaId":"asus","version":"v1"}`},
	}
	sender := &Sender{client: &http.Client{Transport: roundTripper}, endpoints: []string{"https://framework.test", "https://asus.test"}, credential: "token"}
	result, err := sender.Preflight(context.Background())
	if err != nil {
		t.Fatalf("preflight: %v", err)
	}
	if result.ReplicaID != "asus" || result.AgentID != "agent" || strings.Join(roundTripper.calls, ",") != "framework.test,asus.test" {
		t.Fatalf("unexpected preflight result %#v calls %#v", result, roundTripper.calls)
	}
}

func TestSenderPreflightDoesNotFailOverOnCredentialRejection(t *testing.T) {
	roundTripper := &preflightRoundTripper{statuses: []int{http.StatusUnauthorized}, bodies: []string{"{}"}}
	sender := &Sender{client: &http.Client{Transport: roundTripper}, endpoints: []string{"https://framework.test", "https://asus.test"}, credential: "token"}
	if _, err := sender.Preflight(context.Background()); err == nil {
		t.Fatal("expected preflight credential rejection")
	}
	if len(roundTripper.calls) != 1 {
		t.Fatalf("credential rejection must not fail over: %#v", roundTripper.calls)
	}
}

func TestSenderCircuitSkipsRepeatedlyFailingPreferredReplica(t *testing.T) {
	roundTripper := &scriptedRoundTripper{statuses: []int{503, 202, 503, 202, 503, 202, 202}}
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	sender := &Sender{
		client:           &http.Client{Transport: roundTripper},
		endpoints:        []string{"https://framework.test", "https://asus.test"},
		credential:       "token",
		circuits:         map[string]endpointCircuit{},
		now:              func() time.Time { return now },
		failureThreshold: 3,
		cooldown:         time.Minute,
	}
	for attempt := 0; attempt < 4; attempt++ {
		if err := sender.Send(context.Background(), validAgentEnvelope()); err != nil {
			t.Fatalf("send attempt %d: %v", attempt, err)
		}
	}
	if got, want := strings.Join(roundTripper.calls, ","), "framework.test,asus.test,framework.test,asus.test,framework.test,asus.test,asus.test"; got != want {
		t.Fatalf("unexpected circuit path %q, want %q", got, want)
	}
}

func TestSenderReturnsToPreferredReplicaAfterCircuitCooldown(t *testing.T) {
	roundTripper := &scriptedRoundTripper{statuses: []int{503, 202, 202}}
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	sender := &Sender{
		client:           &http.Client{Transport: roundTripper},
		endpoints:        []string{"https://framework.test", "https://asus.test"},
		credential:       "token",
		circuits:         map[string]endpointCircuit{},
		now:              func() time.Time { return now },
		failureThreshold: 1,
		cooldown:         time.Minute,
	}
	if err := sender.Send(context.Background(), validAgentEnvelope()); err != nil {
		t.Fatalf("initial preferred-replica failure should fall back: %v", err)
	}
	now = now.Add(time.Minute)
	if err := sender.Send(context.Background(), validAgentEnvelope()); err != nil {
		t.Fatalf("preferred replica should be retried after cooldown: %v", err)
	}
	if got, want := strings.Join(roundTripper.calls, ","), "framework.test,asus.test,framework.test"; got != want {
		t.Fatalf("unexpected failback path %q, want %q", got, want)
	}
}

func TestSenderDoesNotFollowIngestionRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("ingestion redirect target received telemetry")
	}))
	defer target.Close()
	primary := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusTemporaryRedirect)
	}))
	defer primary.Close()
	sender := &Sender{client: primary.Client(), endpoints: []string{primary.URL}, credential: "credential-canary"}
	if err := sender.Send(context.Background(), validAgentEnvelope()); err == nil {
		t.Fatal("expected redirected ingestion to be treated as unavailable")
	}
}

func TestSenderDoesNotFollowPreflightRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("preflight redirect target received credentials")
	}))
	defer target.Close()
	primary := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusFound)
	}))
	defer primary.Close()
	sender := &Sender{client: primary.Client(), endpoints: []string{primary.URL}, credential: "credential-canary"}
	if _, err := sender.Preflight(context.Background()); err == nil {
		t.Fatal("expected redirected preflight to be rejected")
	}
}
