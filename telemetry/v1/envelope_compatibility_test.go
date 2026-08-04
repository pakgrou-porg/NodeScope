package telemetryv1

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestEnvelopeV1RoundTripsKnownFields(t *testing.T) {
	original := &Envelope{
		SchemaVersion:   1,
		Codec:           "protobuf+zstd",
		AgentId:         "agent-framework",
		HostId:          "framework",
		BootId:          "boot-001",
		Sequence:        42,
		SampleCount:     1,
		MetricValueCount: 1,
		Samples: []*MetricSample{{
			DeviceId:       "cpu-0",
			MetricName:     "cpu.utilization",
			Unit:           "percent",
			Quality:        MetricQuality_METRIC_QUALITY_FRESH,
			Source:         "procfs",
			Semantics:      "aggregate host CPU utilization",
			ScalarValue: &MetricSample_NumberValue{
				NumberValue: 47.5,
			},
		}},
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal v1 envelope: %v", err)
	}
	var decoded Envelope
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("unmarshal v1 envelope: %v", err)
	}

	if decoded.GetAgentId() != original.GetAgentId() || decoded.GetSequence() != original.GetSequence() {
		t.Fatalf("expected known v1 fields to survive round trip: %#v", decoded)
	}
	if len(decoded.GetSamples()) != 1 || decoded.GetSamples()[0].GetNumberValue() != 47.5 {
		t.Fatalf("expected known v1 sample to survive round trip: %#v", decoded.GetSamples())
	}
}

func TestEnvelopePreservesUnknownFutureFields(t *testing.T) {
	known, err := proto.Marshal(&Envelope{SchemaVersion: 1, AgentId: "agent-framework", HostId: "framework", BootId: "boot-001", Sequence: 1})
	if err != nil {
		t.Fatalf("marshal known v1 envelope: %v", err)
	}

	// Field 99, wire type 0, value 1. A future schema can add this field while
	// an older receiver must retain it when it reads and forwards the message.
	futureField := []byte{0x98, 0x06, 0x01}
	wireWithFutureField := append(known, futureField...)

	var olderReader Envelope
	if err := proto.Unmarshal(wireWithFutureField, &olderReader); err != nil {
		t.Fatalf("older reader should accept unknown future field: %v", err)
	}
	roundTripped, err := proto.Marshal(&olderReader)
	if err != nil {
		t.Fatalf("re-marshal older reader: %v", err)
	}
	if !bytes.Contains(roundTripped, futureField) {
		t.Fatalf("older reader discarded unknown future field: %x", roundTripped)
	}
}
