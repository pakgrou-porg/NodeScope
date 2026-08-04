package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type scriptedSender struct {
	calls int
	errs  []error
}

func (sender *scriptedSender) Send(_ context.Context, _ telemetry.Envelope) error {
	index := sender.calls
	sender.calls++
	if index >= len(sender.errs) {
		return nil
	}
	return sender.errs[index]
}

func TestRetryPolicyBoundsAndJitter(t *testing.T) {
	policy := DefaultRetryPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("validate policy: %v", err)
	}
	if delay := policy.Delay(0, func() float64 { return 0 }); delay != 800*time.Millisecond {
		t.Fatalf("unexpected lower jitter bound %s", delay)
	}
	if delay := policy.Delay(10, func() float64 { return 1 }); delay != 36*time.Second {
		t.Fatalf("unexpected capped upper jitter bound %s", delay)
	}
}

func TestRunnerRetriesSameEnvelopeOnlyForTransientFailure(t *testing.T) {
	state, err := OpenSequenceStore(t.TempDir())
	if err != nil {
		t.Fatalf("open state: %v", err)
	}
	value := 1.0
	sender := &scriptedSender{errs: []error{
		&DeliveryError{Reason: "temporary", CanRetry: true},
		nil,
	}}
	runner, err := NewRunner(Config{AgentID: "agent", HostID: "host"}, []Collector{fakeCollector{name: "cpu", samples: []telemetry.Sample{{DeviceID: "cpu", Metric: domain.MetricValue{Name: "cpu.utilization", Unit: "percent", Value: &value, Quality: domain.QualityFresh, Source: "test", Semantics: "test", ObservedAt: time.Now()}}}}}, sender, state)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	runner.retryPolicy = RetryPolicy{MaxAttempts: 2, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond, JitterFraction: 0}
	runner.sleep = func(context.Context, time.Duration) error { return nil }
	if err := runner.CollectOnce(context.Background()); err != nil {
		t.Fatalf("collect once: %v", err)
	}
	if sender.calls != 2 {
		t.Fatalf("expected two delivery attempts, got %d", sender.calls)
	}
}

func TestRunnerFailsClosedOnAuthorizationFailure(t *testing.T) {
	state, err := OpenSequenceStore(t.TempDir())
	if err != nil {
		t.Fatalf("open state: %v", err)
	}
	value := 1.0
	sender := &scriptedSender{errs: []error{&DeliveryError{Reason: "unauthorized", CanRetry: false, StatusCode: 401}}}
	runner, err := NewRunner(Config{AgentID: "agent", HostID: "host"}, []Collector{fakeCollector{name: "cpu", samples: []telemetry.Sample{{DeviceID: "cpu", Metric: domain.MetricValue{Name: "cpu.utilization", Unit: "percent", Value: &value, Quality: domain.QualityFresh, Source: "test", Semantics: "test", ObservedAt: time.Now()}}}}}, sender, state)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	err = runner.CollectOnce(context.Background())
	if !errors.As(err, new(*DeliveryError)) {
		t.Fatalf("expected delivery error, got %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("terminal authorization error must not retry, got %d calls", sender.calls)
	}
}
