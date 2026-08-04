package telemetry

import (
	"fmt"

	"github.com/DataDog/sketches-go/ddsketch"
	"google.golang.org/protobuf/proto"
)

const (
	QuantileAlgorithmDDSketch = "ddsketch"
	QuantileFormatVersion     = 1
	DefaultRelativeAccuracy   = 0.01
	DefaultMaxBins            = 2048
)

// QuantileSketch is a versioned, bounded DDSketch wrapper. Its metadata is
// persisted with a rollup so readers can reject incompatible sketch formats
// rather than silently computing an invalid p95.
type QuantileSketch struct {
	Algorithm        string  `json:"algorithm"`
	FormatVersion    int     `json:"formatVersion"`
	RelativeAccuracy float64 `json:"relativeAccuracy"`
	MaxBins          int     `json:"maxBins"`
	sketch           *ddsketch.DDSketch
}

func NewQuantileSketch() (*QuantileSketch, error) {
	sketch, err := ddsketch.LogCollapsingHighestDenseDDSketch(DefaultRelativeAccuracy, DefaultMaxBins)
	if err != nil {
		return nil, fmt.Errorf("create DDSketch: %w", err)
	}
	return &QuantileSketch{
		Algorithm:        QuantileAlgorithmDDSketch,
		FormatVersion:    QuantileFormatVersion,
		RelativeAccuracy: DefaultRelativeAccuracy,
		MaxBins:          DefaultMaxBins,
		sketch:           sketch,
	}, nil
}

func (q *QuantileSketch) ValidateMetadata() error {
	if q == nil || q.sketch == nil {
		return fmt.Errorf("quantile sketch is not initialized")
	}
	if q.Algorithm != QuantileAlgorithmDDSketch {
		return fmt.Errorf("unsupported quantile algorithm %q", q.Algorithm)
	}
	if q.FormatVersion != QuantileFormatVersion {
		return fmt.Errorf("unsupported quantile format version %d", q.FormatVersion)
	}
	if q.RelativeAccuracy <= 0 || q.RelativeAccuracy > 0.1 {
		return fmt.Errorf("invalid relative accuracy %f", q.RelativeAccuracy)
	}
	if q.MaxBins < 64 {
		return fmt.Errorf("invalid max bins %d", q.MaxBins)
	}
	return nil
}

func (q *QuantileSketch) Add(value float64) error {
	if err := q.ValidateMetadata(); err != nil {
		return err
	}
	if err := q.sketch.Add(value); err != nil {
		return fmt.Errorf("add value to DDSketch: %w", err)
	}
	return nil
}

func (q *QuantileSketch) Count() float64 {
	if q == nil || q.sketch == nil {
		return 0
	}
	return q.sketch.GetCount()
}

func (q *QuantileSketch) P95() (float64, error) {
	if err := q.ValidateMetadata(); err != nil {
		return 0, err
	}
	if q.sketch.IsEmpty() {
		return 0, fmt.Errorf("cannot calculate p95 for an empty DDSketch")
	}
	value, err := q.sketch.GetValueAtQuantile(0.95)
	if err != nil {
		return 0, fmt.Errorf("calculate p95: %w", err)
	}
	return value, nil
}

func (q *QuantileSketch) Merge(other *QuantileSketch) error {
	if err := q.ValidateMetadata(); err != nil {
		return err
	}
	if err := other.ValidateMetadata(); err != nil {
		return err
	}
	if q.Algorithm != other.Algorithm || q.FormatVersion != other.FormatVersion || q.RelativeAccuracy != other.RelativeAccuracy || q.MaxBins != other.MaxBins {
		return fmt.Errorf("cannot merge incompatible DDSketch metadata")
	}
	if err := q.sketch.MergeWith(other.sketch); err != nil {
		return fmt.Errorf("merge DDSketch: %w", err)
	}
	return nil
}

// Marshal serializes the internal protobuf form. The format metadata remains
// outside this payload in the rollup row and is validated before unmarshal.
func (q *QuantileSketch) Marshal() ([]byte, error) {
	if err := q.ValidateMetadata(); err != nil {
		return nil, err
	}
	bytes, err := proto.Marshal(q.sketch.ToProto())
	if err != nil {
		return nil, fmt.Errorf("marshal DDSketch: %w", err)
	}
	return bytes, nil
}

func UnmarshalQuantileSketch(algorithm string, formatVersion int, relativeAccuracy float64, maxBins int, payload []byte) (*QuantileSketch, error) {
	candidate, err := NewQuantileSketch()
	if err != nil {
		return nil, err
	}
	candidate.Algorithm = algorithm
	candidate.FormatVersion = formatVersion
	candidate.RelativeAccuracy = relativeAccuracy
	candidate.MaxBins = maxBins
	if err := candidate.ValidateMetadata(); err != nil {
		return nil, err
	}
	message := candidate.sketch.ToProto()
	if err := proto.Unmarshal(payload, message); err != nil {
		return nil, fmt.Errorf("unmarshal DDSketch payload: %w", err)
	}
	decoded, err := ddsketch.FromProto(message)
	if err != nil {
		return nil, fmt.Errorf("decode DDSketch payload: %w", err)
	}
	candidate.sketch = decoded
	return candidate, nil
}
