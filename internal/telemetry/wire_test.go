package telemetry

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func codecEnvelope() Envelope {
	observedAt := time.Date(2026, 7, 22, 12, 10, 0, 0, time.UTC)
	value := 73.25
	checksum := sha256.Sum256([]byte("bootstrap"))
	return Envelope{
		SchemaVersion:     CurrentSchemaVersion,
		Codec:             CodecProtoZstd,
		AgentID:           "agent-id",
		HostID:            "host-id",
		BootID:            "boot-id",
		Sequence:          1,
		ObservedAt:        observedAt,
		SampleCount:       1,
		MetricValueCount:  1,
		UncompressedBytes: 1,
		CompressedBytes:   1,
		ChecksumSHA256:    fmt.Sprintf("%x", checksum),
		Samples: []Sample{{
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

func TestCompressedEnvelopeRoundTrip(t *testing.T) {
	input := codecEnvelope()
	encoded, err := EncodeCompressedEnvelope(input)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := DecodeCompressedEnvelope(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.AgentID != input.AgentID || decoded.HostID != input.HostID || len(decoded.Samples) != 1 {
		t.Fatalf("unexpected decoded envelope %#v", decoded)
	}
	if decoded.Samples[0].Metric.Value == nil || *decoded.Samples[0].Metric.Value != 73.25 {
		t.Fatalf("unexpected decoded metric %#v", decoded.Samples[0].Metric)
	}
}

func TestCompressedEnvelopePreservesExperimentalQuality(t *testing.T) {
	input := codecEnvelope()
	input.Samples[0].Metric.Quality = domain.QualityExperimental
	input.Samples[0].Metric.Source = "sysfs-experimental"
	input.Samples[0].Metric.Semantics = "unqualified Fedora AMD DRM evidence"

	encoded, err := EncodeCompressedEnvelope(input)
	if err != nil {
		t.Fatalf("encode experimental evidence: %v", err)
	}
	decoded, err := DecodeCompressedEnvelope(encoded)
	if err != nil {
		t.Fatalf("decode experimental evidence: %v", err)
	}
	if decoded.Samples[0].Metric.Quality != domain.QualityExperimental {
		t.Fatalf("experimental quality was not preserved: %#v", decoded.Samples[0].Metric)
	}
}

func TestCompressedEnvelopeRejectsCorruption(t *testing.T) {
	encoded, err := EncodeCompressedEnvelope(codecEnvelope())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	encoded[len(encoded)-1] ^= 0xff
	if _, err := DecodeCompressedEnvelope(encoded); err == nil {
		t.Fatal("corrupted compressed payload must be rejected")
	}
}

func TestCompressedEnvelopeRejectsOversize(t *testing.T) {
	oversize := make([]byte, MaxCompressedBytes+1)
	if _, err := DecodeCompressedEnvelope(oversize); err == nil {
		t.Fatal("oversized compressed payload must be rejected")
	}
}
