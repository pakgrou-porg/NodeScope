package ingest

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

func ingestEnvelope(sequence uint64) telemetry.Envelope {
	observedAt := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	checksum := sha256.Sum256([]byte(fmt.Sprintf("ingest-%d", sequence)))
	value := 1.0
	return telemetry.Envelope{
		SchemaVersion:     telemetry.CurrentSchemaVersion,
		Codec:             telemetry.CodecProtoZstd,
		AgentID:           "agent-framework",
		HostID:            "framework",
		BootID:            "boot-1",
		Sequence:          sequence,
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
				Semantics:  "aggregate host CPU utilization",
				ObservedAt: observedAt,
			},
		}},
	}
}

func TestAcceptsFirstEnvelopeAndDeduplicatesRetry(t *testing.T) {
	service, err := NewService(DefaultPolicy())
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	envelope := ingestEnvelope(1)

	first, err := service.Accept("agent-framework", envelope)
	if err != nil || first.Outcome != OutcomeAccepted {
		t.Fatalf("expected accepted receipt, got %#v err=%v", first, err)
	}
	retry, err := service.Accept("agent-framework", envelope)
	if err != nil || retry.Outcome != OutcomeDuplicate {
		t.Fatalf("expected duplicate receipt, got %#v err=%v", retry, err)
	}
}

func TestRejectsAgentIdentityMismatch(t *testing.T) {
	service, err := NewService(DefaultPolicy())
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	if _, err := service.Accept("agent-asus", ingestEnvelope(1)); err == nil {
		t.Fatal("expected mismatched authenticated agent to be rejected")
	}
}

func TestThrottlesBoundedAgentRate(t *testing.T) {
	policy := Policy{MaxRequestsPerMinute: 1, Burst: 1, DeduplicationTTL: time.Hour}
	service, err := NewService(policy)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	for sequence := uint64(1); sequence <= 2; sequence++ {
		receipt, err := service.Accept("agent-framework", ingestEnvelope(sequence))
		if err != nil || receipt.Outcome != OutcomeAccepted {
			t.Fatalf("request %d expected accepted, got %#v err=%v", sequence, receipt, err)
		}
	}
	receipt, err := service.Accept("agent-framework", ingestEnvelope(3))
	if err != nil {
		t.Fatalf("expected throttled receipt, got error %v", err)
	}
	if receipt.Outcome != OutcomeThrottled {
		t.Fatalf("expected throttled receipt, got %#v", receipt)
	}
}

func TestDeduplicationExpires(t *testing.T) {
	policy := Policy{MaxRequestsPerMinute: 10, Burst: 1, DeduplicationTTL: time.Minute}
	service, err := NewService(policy)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	envelope := ingestEnvelope(1)
	if receipt, err := service.Accept("agent-framework", envelope); err != nil || receipt.Outcome != OutcomeAccepted {
		t.Fatalf("first receipt %#v err=%v", receipt, err)
	}
	now = now.Add(2 * time.Minute)
	if receipt, err := service.Accept("agent-framework", envelope); err != nil || receipt.Outcome != OutcomeAccepted {
		t.Fatalf("expired retry should be accepted, got %#v err=%v", receipt, err)
	}
}

func TestResetsPerAgentRateWindowAfterOneMinute(t *testing.T) {
	service, err := NewService(Policy{MaxRequestsPerMinute: 1, Burst: 1, DeduplicationTTL: time.Hour})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	if receipt, err := service.Accept("agent-framework", ingestEnvelope(1)); err != nil || receipt.Outcome != OutcomeAccepted {
		t.Fatalf("first receipt %#v err=%v", receipt, err)
	}
	if receipt, err := service.Accept("agent-framework", ingestEnvelope(2)); err != nil || receipt.Outcome != OutcomeAccepted {
		t.Fatalf("burst receipt %#v err=%v", receipt, err)
	}
	if receipt, err := service.Accept("agent-framework", ingestEnvelope(3)); err != nil || receipt.Outcome != OutcomeThrottled {
		t.Fatalf("bounded receipt %#v err=%v", receipt, err)
	}
	now = now.Add(time.Minute)
	if receipt, err := service.Accept("agent-framework", ingestEnvelope(4)); err != nil || receipt.Outcome != OutcomeAccepted {
		t.Fatalf("reset receipt %#v err=%v", receipt, err)
	}
}
