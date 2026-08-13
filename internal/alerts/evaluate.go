// Package alerts contains deterministic, evidence-aware automatic alert policy.
package alerts

import (
	"fmt"
	"math"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

type Operator string

const (
	OperatorGreaterThan        Operator = "gt"
	OperatorGreaterThanOrEqual Operator = "gte"
	OperatorLessThan           Operator = "lt"
	OperatorLessThanOrEqual    Operator = "lte"
)

type Rule struct {
	Metric    string
	Operator  Operator
	Threshold float64
}

type Decision struct {
	Triggered  bool
	Suppressed bool
	Reason     string
}

// Evaluate applies one threshold rule to one metric without retaining or
// interpreting any content beyond the metric's scalar evidence. It suppresses
// every quality except fresh, even when the numeric value crosses a threshold.
func Evaluate(rule Rule, metric domain.MetricValue) (Decision, error) {
	if rule.Metric == "" || rule.Metric != metric.Name {
		return Decision{}, fmt.Errorf("alert rule metric %q does not match evidence %q", rule.Metric, metric.Name)
	}
	if math.IsNaN(rule.Threshold) || math.IsInf(rule.Threshold, 0) {
		return Decision{}, fmt.Errorf("alert threshold must be finite")
	}
	if err := metric.Validate(); err != nil {
		return Decision{}, fmt.Errorf("validate metric evidence: %w", err)
	}
	if !metric.Quality.EligibleForAutomaticAlerting() {
		return Decision{Suppressed: true, Reason: "ineligible_quality:" + string(metric.Quality)}, nil
	}
	if metric.Value == nil {
		return Decision{}, fmt.Errorf("fresh alert evidence requires a numeric value")
	}

	triggered, err := compare(*metric.Value, rule.Operator, rule.Threshold)
	if err != nil {
		return Decision{}, err
	}
	return Decision{Triggered: triggered, Reason: "evaluated_fresh_evidence"}, nil
}

func compare(value float64, operator Operator, threshold float64) (bool, error) {
	switch operator {
	case OperatorGreaterThan:
		return value > threshold, nil
	case OperatorGreaterThanOrEqual:
		return value >= threshold, nil
	case OperatorLessThan:
		return value < threshold, nil
	case OperatorLessThanOrEqual:
		return value <= threshold, nil
	default:
		return false, fmt.Errorf("unsupported alert operator %q", operator)
	}
}
