package telemetry

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func testEnvelope() Envelope {
	observed := time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC)
	checksum := sha256.Sum256([]byte("nodescope-test-envelope"))
	value := 42.0
	return Envelope{
		SchemaVersion:     CurrentSchemaVersion,
		Codec:             CodecProtoZstd,
		AgentID:           "agent-framework",
		HostID:            "framework",
		BootID:            "boot-20260721",
		Sequence:          1,
		ObservedAt:        observed,
		SampleCount:       1,
		MetricValueCount:  1,
		UncompressedBytes: 512,
		CompressedBytes:   256,
		ChecksumSHA256:    fmt.Sprintf("%x", checksum),
		Samples: []Sample{{
			DeviceID: "cpu-0",
			Metric: domain.MetricValue{
				Name:       "cpu.utilization",
				Unit:       "percent",
				Value:      &value,
				Quality:    domain.QualityFresh,
				Source:     "procfs",
				Semantics:  "aggregate host CPU utilization",
				ObservedAt: observed,
			},
		}},
	}
}

func TestEnvelopeValidates(t *testing.T) {
	envelope := testEnvelope()
	if err := envelope.Validate(); err != nil {
		t.Fatalf("expected valid envelope, received %v", err)
	}
	if got, want := envelope.IdempotencyKey(), "agent-framework:boot-20260721:1"; got != want {
		t.Fatalf("expected idempotency key %q, got %q", want, got)
	}
}

func TestEnvelopeRejectsOversizedPayload(t *testing.T) {
	envelope := testEnvelope()
	envelope.CompressedBytes = MaxCompressedBytes + 1
	if err := envelope.Validate(); err == nil {
		t.Fatal("expected oversized compressed envelope to be rejected")
	}
}

func TestEnvelopeRejectsUnavailableMetricValue(t *testing.T) {
	envelope := testEnvelope()
	envelope.Samples[0].Metric.Quality = domain.QualityUnavailable
	if err := envelope.Validate(); err == nil {
		t.Fatal("expected unavailable metric with value to be rejected")
	}
}

func TestEnvelopeRejectsCountMismatch(t *testing.T) {
	envelope := testEnvelope()
	envelope.MetricValueCount = 2
	if err := envelope.Validate(); err == nil {
		t.Fatal("expected mismatched count to be rejected")
	}
}

func TestEnvelopeReceiptTimeValidationRejectsMateriallyFutureObservations(t *testing.T) {
	receivedAt := time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC)

	t.Run("within tolerance", func(t *testing.T) {
		envelope := testEnvelope()
		envelope.ObservedAt = receivedAt.Add(MaxFutureObservationSkew)
		envelope.Samples[0].Metric.ObservedAt = envelope.ObservedAt
		if err := envelope.ValidateReceiptTimes(receivedAt); err != nil {
			t.Fatalf("expected tolerated future skew, received %v", err)
		}
	})

	t.Run("envelope ahead", func(t *testing.T) {
		envelope := testEnvelope()
		envelope.ObservedAt = receivedAt.Add(MaxFutureObservationSkew + time.Nanosecond)
		if err := envelope.ValidateReceiptTimes(receivedAt); err == nil {
			t.Fatal("expected future envelope observation to be rejected")
		}
	})

	t.Run("sample ahead", func(t *testing.T) {
		envelope := testEnvelope()
		envelope.ObservedAt = receivedAt
		envelope.Samples[0].Metric.ObservedAt = receivedAt.Add(MaxFutureObservationSkew + time.Nanosecond)
		if err := envelope.ValidateReceiptTimes(receivedAt); err == nil {
			t.Fatal("expected future sample observation to be rejected")
		}
	})
}
