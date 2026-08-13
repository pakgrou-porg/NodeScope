// Package telemetry validates bounded agent submissions before persistence.
package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

const (
	CurrentSchemaVersion  = 1
	CodecProtoZstd        = "protobuf+zstd"
	MaxCompressedBytes    = 1 << 20 // 1 MiB
	MaxUncompressedBytes  = 8 << 20 // 8 MiB
	MaxMetricValues       = 10_000
	MaxContainerInventory = 2_000
	// MaxFutureObservationSkew is aligned with the server clock-offset evidence
	// threshold. Larger positive skew would let an agent advance latest-state
	// ordering beyond later valid receipts.
	MaxFutureObservationSkew = time.Minute
)

// Sample groups one measured value with the device from which it originated.
type Sample struct {
	DeviceID string             `json:"deviceId"`
	Metric   domain.MetricValue `json:"metric"`
}

// Envelope is the logical, decoded form of a signed agent submission. Actual
// wire compression is handled at the transport boundary; these fields make the
// declared limits auditable and testable.
type ContainerInventory struct {
	ContainerID         string `json:"containerId"`
	Name                string `json:"name"`
	Image               string `json:"image"`
	State               string `json:"state"`
	Health              string `json:"health"`
	SelectedForAlerting bool   `json:"selectedForAlerting"`
}

type Envelope struct {
	SchemaVersion        uint32               `json:"schemaVersion"`
	Codec                string               `json:"codec"`
	AgentID              string               `json:"agentId"`
	HostID               string               `json:"hostId"`
	BootID               string               `json:"bootId"`
	Sequence             uint64               `json:"sequence"`
	ObservedAt           time.Time            `json:"observedAt"`
	MonotonicElapsedNano uint64               `json:"monotonicElapsedNano"`
	SampleCount          uint32               `json:"sampleCount"`
	MetricValueCount     uint32               `json:"metricValueCount"`
	UncompressedBytes    uint32               `json:"uncompressedBytes"`
	CompressedBytes      uint32               `json:"compressedBytes"`
	ChecksumSHA256       string               `json:"checksumSha256"`
	Samples              []Sample             `json:"samples"`
	Containers           []ContainerInventory `json:"containers,omitempty"`
}

func (e Envelope) IdempotencyKey() string {
	return strings.Join([]string{e.AgentID, e.BootID, fmt.Sprintf("%d", e.Sequence)}, ":")
}

func (e Envelope) Validate() error {
	if e.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported telemetry schema version %d", e.SchemaVersion)
	}
	if e.Codec != CodecProtoZstd {
		return fmt.Errorf("unsupported telemetry codec %q", e.Codec)
	}
	for label, value := range map[string]string{
		"agent ID": e.AgentID,
		"host ID":  e.HostID,
		"boot ID":  e.BootID,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", label)
		}
	}
	if e.Sequence == 0 {
		return fmt.Errorf("sequence must be greater than zero")
	}
	if e.ObservedAt.IsZero() {
		return fmt.Errorf("observed time is required")
	}
	if e.CompressedBytes == 0 || e.CompressedBytes > MaxCompressedBytes {
		return fmt.Errorf("compressed size must be between 1 and %d bytes", MaxCompressedBytes)
	}
	if e.UncompressedBytes == 0 || e.UncompressedBytes > MaxUncompressedBytes {
		return fmt.Errorf("uncompressed size must be between 1 and %d bytes", MaxUncompressedBytes)
	}
	if e.MetricValueCount == 0 || e.MetricValueCount > MaxMetricValues {
		return fmt.Errorf("metric value count must be between 1 and %d", MaxMetricValues)
	}
	if int(e.SampleCount) != len(e.Samples) || int(e.MetricValueCount) != len(e.Samples) {
		return fmt.Errorf("declared sample counts do not match decoded samples")
	}
	if len(e.ChecksumSHA256) != sha256.Size*2 {
		return fmt.Errorf("checksum must be a SHA-256 hexadecimal digest")
	}
	if _, err := hex.DecodeString(e.ChecksumSHA256); err != nil {
		return fmt.Errorf("checksum must be hexadecimal: %w", err)
	}

	if len(e.Containers) > MaxContainerInventory {
		return fmt.Errorf("container inventory must not exceed %d records", MaxContainerInventory)
	}
	seenContainers := make(map[string]struct{}, len(e.Containers))
	for index, container := range e.Containers {
		for label, value := range map[string]string{
			"container ID":    container.ContainerID,
			"container name":  container.Name,
			"container image": container.Image,
			"container state": container.State,
		} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("container %d requires %s", index, label)
			}
		}
		if _, exists := seenContainers[container.ContainerID]; exists {
			return fmt.Errorf("duplicate container inventory %q", container.ContainerID)
		}
		seenContainers[container.ContainerID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(e.Samples))
	for index, sample := range e.Samples {
		if strings.TrimSpace(sample.DeviceID) == "" {
			return fmt.Errorf("sample %d requires a device ID", index)
		}
		if err := sample.Metric.Validate(); err != nil {
			return fmt.Errorf("sample %d invalid: %w", index, err)
		}
		key := strings.Join([]string{sample.DeviceID, sample.Metric.Name, sample.Metric.ObservedAt.UTC().Format(time.RFC3339Nano)}, ":")
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate metric sample %q", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// ValidateReceiptTimes rejects observations materially ahead of server receipt
// time. A bounded tolerance allows normal network and clock jitter, while
// preventing a future-dated agent from blocking later latest-state updates.
func (e Envelope) ValidateReceiptTimes(receivedAt time.Time) error {
	upperBound := receivedAt.UTC().Add(MaxFutureObservationSkew)
	if e.ObservedAt.After(upperBound) {
		return fmt.Errorf("envelope observation time is too far ahead of server receipt time")
	}
	for index, sample := range e.Samples {
		if sample.Metric.ObservedAt.After(upperBound) {
			return fmt.Errorf("sample %d observation time is too far ahead of server receipt time", index)
		}
	}
	return nil
}
