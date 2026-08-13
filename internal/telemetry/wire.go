package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/pakgrou-porg/nodescope/internal/domain"
	telemetryv1 "github.com/pakgrou-porg/nodescope/telemetry/v1"
	"google.golang.org/protobuf/proto"
)

// DecodeCompressedEnvelope enforces the transport size bounds before mapping a
// native-agent submission into NodeScope's provenance-aware domain envelope.
func DecodeCompressedEnvelope(compressed []byte) (Envelope, error) {
	if len(compressed) == 0 || len(compressed) > MaxCompressedBytes {
		return Envelope{}, fmt.Errorf("compressed payload must be between 1 and %d bytes", MaxCompressedBytes)
	}
	decoder, err := zstd.NewReader(nil, zstd.WithDecoderMaxMemory(uint64(MaxUncompressedBytes)*2))
	if err != nil {
		return Envelope{}, fmt.Errorf("create zstd decoder: %w", err)
	}
	decompressed, err := decoder.DecodeAll(compressed, nil)
	decoder.Close()
	if err != nil {
		return Envelope{}, fmt.Errorf("decode zstd telemetry payload: %w", err)
	}
	if len(decompressed) == 0 || len(decompressed) > MaxUncompressedBytes {
		return Envelope{}, fmt.Errorf("uncompressed payload must be between 1 and %d bytes", MaxUncompressedBytes)
	}
	message := &telemetryv1.Envelope{}
	if err := proto.Unmarshal(decompressed, message); err != nil {
		return Envelope{}, fmt.Errorf("decode telemetry protobuf: %w", err)
	}
	if int(message.GetCompressedBytes()) != len(compressed) || int(message.GetUncompressedBytes()) != len(decompressed) {
		return Envelope{}, fmt.Errorf("declared transport sizes do not match decoded payload")
	}
	if err := verifyWireChecksum(message); err != nil {
		return Envelope{}, err
	}
	return envelopeFromProto(message)
}

// EncodeCompressedEnvelope produces the only production wire format accepted
// by NodeScope agents. The embedded checksum covers a canonical protobuf form
// with the checksum field omitted, avoiding a self-referential digest.
func EncodeCompressedEnvelope(envelope Envelope) ([]byte, error) {
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	message, err := envelopeToProto(envelope)
	if err != nil {
		return nil, err
	}
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, fmt.Errorf("create zstd encoder: %w", err)
	}
	defer encoder.Close()

	var compressed []byte
	for attempt := 0; attempt < 6; attempt++ {
		message.ChecksumSha256 = nil
		canonical, err := proto.Marshal(message)
		if err != nil {
			return nil, fmt.Errorf("marshal canonical telemetry protobuf: %w", err)
		}
		digest := sha256.Sum256(canonical)
		message.ChecksumSha256 = digest[:]
		raw, err := proto.Marshal(message)
		if err != nil {
			return nil, fmt.Errorf("marshal telemetry protobuf: %w", err)
		}
		message.UncompressedBytes = uint32(len(raw))
		compressed = encoder.EncodeAll(raw, nil)
		message.CompressedBytes = uint32(len(compressed))
	}
	message.ChecksumSha256 = nil
	canonical, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal final canonical telemetry protobuf: %w", err)
	}
	digest := sha256.Sum256(canonical)
	message.ChecksumSha256 = digest[:]
	raw, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal final telemetry protobuf: %w", err)
	}
	compressed = encoder.EncodeAll(raw, nil)
	if len(raw) > MaxUncompressedBytes || len(compressed) > MaxCompressedBytes {
		return nil, fmt.Errorf("encoded telemetry exceeds transport bounds")
	}
	if int(message.UncompressedBytes) != len(raw) || int(message.CompressedBytes) != len(compressed) {
		return nil, fmt.Errorf("telemetry transport lengths did not converge")
	}
	return compressed, nil
}

func envelopeFromProto(message *telemetryv1.Envelope) (Envelope, error) {
	qualityMap := map[telemetryv1.MetricQuality]domain.MetricQuality{
		telemetryv1.MetricQuality_METRIC_QUALITY_FRESH:        domain.QualityFresh,
		telemetryv1.MetricQuality_METRIC_QUALITY_STALE:        domain.QualityStale,
		telemetryv1.MetricQuality_METRIC_QUALITY_UNAVAILABLE:  domain.QualityUnavailable,
		telemetryv1.MetricQuality_METRIC_QUALITY_UNSUPPORTED:  domain.QualityUnsupported,
		telemetryv1.MetricQuality_METRIC_QUALITY_ESTIMATED:    domain.QualityEstimated,
		telemetryv1.MetricQuality_METRIC_QUALITY_EXPERIMENTAL: domain.QualityExperimental,
	}
	envelope := Envelope{
		SchemaVersion:        message.GetSchemaVersion(),
		Codec:                message.GetCodec(),
		AgentID:              message.GetAgentId(),
		HostID:               message.GetHostId(),
		BootID:               message.GetBootId(),
		Sequence:             message.GetSequence(),
		ObservedAt:           time.Unix(0, message.GetObservedUnixNano()).UTC(),
		MonotonicElapsedNano: message.GetMonotonicElapsedNano(),
		SampleCount:          message.GetSampleCount(),
		MetricValueCount:     message.GetMetricValueCount(),
		UncompressedBytes:    message.GetUncompressedBytes(),
		CompressedBytes:      message.GetCompressedBytes(),
		ChecksumSHA256:       hex.EncodeToString(message.GetChecksumSha256()),
		Samples:              make([]Sample, 0, len(message.GetSamples())),
		Containers:           make([]ContainerInventory, 0, len(message.GetContainers())),
	}
	for _, sample := range message.GetSamples() {
		quality, exists := qualityMap[sample.GetQuality()]
		if !exists {
			return Envelope{}, fmt.Errorf("unsupported metric quality %s", sample.GetQuality().String())
		}
		var value *float64
		switch scalar := sample.GetScalarValue().(type) {
		case *telemetryv1.MetricSample_NumberValue:
			candidate := scalar.NumberValue
			value = &candidate
		case nil:
			// Unavailable and unsupported values deliberately remain nil.
		default:
			return Envelope{}, fmt.Errorf("metric %q uses a scalar type unsupported by Release 1", sample.GetMetricName())
		}
		envelope.Samples = append(envelope.Samples, Sample{
			DeviceID: sample.GetDeviceId(),
			Metric: domain.MetricValue{
				Name:       sample.GetMetricName(),
				Unit:       sample.GetUnit(),
				Value:      value,
				Quality:    quality,
				Source:     sample.GetSource(),
				Semantics:  sample.GetSemantics(),
				ObservedAt: time.Unix(0, sample.GetObservedUnixNano()).UTC(),
			},
		})
	}
	for _, container := range message.GetContainers() {
		envelope.Containers = append(envelope.Containers, ContainerInventory{
			ContainerID:         container.GetContainerId(),
			Name:                container.GetName(),
			Image:               container.GetImage(),
			State:               container.GetState(),
			Health:              container.GetHealth(),
			SelectedForAlerting: container.GetSelectedForAlerting(),
		})
	}
	if err := envelope.Validate(); err != nil {
		return Envelope{}, err
	}
	return envelope, nil
}

func envelopeToProto(envelope Envelope) (*telemetryv1.Envelope, error) {
	qualityMap := map[domain.MetricQuality]telemetryv1.MetricQuality{
		domain.QualityFresh:        telemetryv1.MetricQuality_METRIC_QUALITY_FRESH,
		domain.QualityStale:        telemetryv1.MetricQuality_METRIC_QUALITY_STALE,
		domain.QualityUnavailable:  telemetryv1.MetricQuality_METRIC_QUALITY_UNAVAILABLE,
		domain.QualityUnsupported:  telemetryv1.MetricQuality_METRIC_QUALITY_UNSUPPORTED,
		domain.QualityEstimated:    telemetryv1.MetricQuality_METRIC_QUALITY_ESTIMATED,
		domain.QualityExperimental: telemetryv1.MetricQuality_METRIC_QUALITY_EXPERIMENTAL,
	}
	message := &telemetryv1.Envelope{
		SchemaVersion:        envelope.SchemaVersion,
		Codec:                envelope.Codec,
		AgentId:              envelope.AgentID,
		HostId:               envelope.HostID,
		BootId:               envelope.BootID,
		Sequence:             envelope.Sequence,
		ObservedUnixNano:     envelope.ObservedAt.UTC().UnixNano(),
		MonotonicElapsedNano: envelope.MonotonicElapsedNano,
		SampleCount:          envelope.SampleCount,
		MetricValueCount:     envelope.MetricValueCount,
		Samples:              make([]*telemetryv1.MetricSample, 0, len(envelope.Samples)),
		Containers:           make([]*telemetryv1.ContainerInventory, 0, len(envelope.Containers)),
	}
	for _, sample := range envelope.Samples {
		quality, exists := qualityMap[sample.Metric.Quality]
		if !exists {
			return nil, fmt.Errorf("unsupported metric quality %q", sample.Metric.Quality)
		}
		messageSample := &telemetryv1.MetricSample{
			DeviceId:         sample.DeviceID,
			MetricName:       sample.Metric.Name,
			Unit:             sample.Metric.Unit,
			Quality:          quality,
			Source:           sample.Metric.Source,
			Semantics:        sample.Metric.Semantics,
			ObservedUnixNano: sample.Metric.ObservedAt.UTC().UnixNano(),
		}
		if sample.Metric.Value != nil {
			messageSample.ScalarValue = &telemetryv1.MetricSample_NumberValue{NumberValue: *sample.Metric.Value}
		}
		message.Samples = append(message.Samples, messageSample)
	}
	for _, container := range envelope.Containers {
		message.Containers = append(message.Containers, &telemetryv1.ContainerInventory{
			ContainerId:         container.ContainerID,
			Name:                container.Name,
			Image:               container.Image,
			State:               container.State,
			Health:              container.Health,
			SelectedForAlerting: container.SelectedForAlerting,
		})
	}
	return message, nil
}

func verifyWireChecksum(message *telemetryv1.Envelope) error {
	if len(message.GetChecksumSha256()) != sha256.Size {
		return fmt.Errorf("telemetry protobuf checksum has invalid length")
	}
	expected := append([]byte(nil), message.GetChecksumSha256()...)
	message.ChecksumSha256 = nil
	canonical, err := proto.Marshal(message)
	message.ChecksumSha256 = expected
	if err != nil {
		return fmt.Errorf("marshal canonical checksum payload: %w", err)
	}
	digest := sha256.Sum256(canonical)
	if string(expected) != string(digest[:]) {
		return fmt.Errorf("telemetry protobuf checksum mismatch")
	}
	return nil
}
