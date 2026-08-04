package agent

import (
	"context"
	"io"
	"net/http"
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
