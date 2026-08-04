// Package storage defines NodeScope retention and capacity safety controls.
package storage

import "fmt"

type Mode string

const (
	ModeNormal      Mode = "normal"
	ModeConstrained Mode = "constrained"
	ModeSummaryOnly Mode = "summary_only"
	ModeProtective  Mode = "protective"
)

type Policy struct {
	ConstrainedPercent float64
	SummaryOnlyPercent float64
	ProtectivePercent  float64
}

func DefaultPolicy() Policy {
	return Policy{
		ConstrainedPercent: 70,
		SummaryOnlyPercent: 85,
		ProtectivePercent:  95,
	}
}

func (policy Policy) Validate() error {
	if policy.ConstrainedPercent <= 0 || policy.ConstrainedPercent >= 100 {
		return fmt.Errorf("constrained percentage must be in (0, 100)")
	}
	if policy.SummaryOnlyPercent <= policy.ConstrainedPercent || policy.SummaryOnlyPercent >= 100 {
		return fmt.Errorf("summary-only percentage must exceed constrained percentage and be below 100")
	}
	if policy.ProtectivePercent <= policy.SummaryOnlyPercent || policy.ProtectivePercent >= 100 {
		return fmt.Errorf("protective percentage must exceed summary-only percentage and be below 100")
	}
	return nil
}

type Decision struct {
	Mode                 Mode
	AcceptRawBatches     bool
	AcceptSummaryRollups bool
	RecommendedAction    string
}

func (policy Policy) Decide(usedBytes, quotaBytes int64) (Decision, error) {
	if err := policy.Validate(); err != nil {
		return Decision{}, err
	}
	if quotaBytes <= 0 {
		return Decision{}, fmt.Errorf("quota bytes must be positive")
	}
	if usedBytes < 0 {
		return Decision{}, fmt.Errorf("used bytes cannot be negative")
	}

	usedPercent := float64(usedBytes) / float64(quotaBytes) * 100
	switch {
	case usedPercent >= policy.ProtectivePercent:
		return Decision{
			Mode:                 ModeProtective,
			AcceptRawBatches:     false,
			AcceptSummaryRollups: true,
			RecommendedAction:    "Block raw telemetry storage, retain latest state and summaries, and require operator capacity action.",
		}, nil
	case usedPercent >= policy.SummaryOnlyPercent:
		return Decision{
			Mode:                 ModeSummaryOnly,
			AcceptRawBatches:     false,
			AcceptSummaryRollups: true,
			RecommendedAction:    "Stop retaining raw batches and preserve only latest state and summary rollups.",
		}, nil
	case usedPercent >= policy.ConstrainedPercent:
		return Decision{
			Mode:                 ModeConstrained,
			AcceptRawBatches:     true,
			AcceptSummaryRollups: true,
			RecommendedAction:    "Continue raw telemetry with stricter compression and emit a capacity warning.",
		}, nil
	default:
		return Decision{
			Mode:                 ModeNormal,
			AcceptRawBatches:     true,
			AcceptSummaryRollups: true,
			RecommendedAction:    "Operate with configured raw and summary retention.",
		}, nil
	}
}
