package alerts

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func alertFloat(value float64) *float64 { return &value }

func testMetric(quality domain.MetricQuality, value *float64) domain.MetricValue {
	return domain.MetricValue{
		Name:       "device.temperature_celsius",
		Unit:       "celsius",
		Value:      value,
		Quality:    quality,
		Source:     "hardware-probe",
		Semantics:  "sensor reading",
		ObservedAt: time.Date(2026, 8, 13, 22, 0, 0, 0, time.UTC),
	}
}

func TestEvaluateTriggersOnlyForFreshEvidence(t *testing.T) {
	rule := Rule{Metric: "device.temperature_celsius", Operator: OperatorGreaterThan, Threshold: 82}
	decision, err := Evaluate(rule, testMetric(domain.QualityFresh, alertFloat(83)))
	if err != nil {
		t.Fatalf("evaluate fresh evidence: %v", err)
	}
	if !decision.Triggered || decision.Suppressed || decision.Reason != "evaluated_fresh_evidence" {
		t.Fatalf("unexpected fresh decision: %#v", decision)
	}

	decision, err = Evaluate(rule, testMetric(domain.QualityFresh, alertFloat(82)))
	if err != nil {
		t.Fatalf("evaluate below-threshold fresh evidence: %v", err)
	}
	if decision.Triggered || decision.Suppressed {
		t.Fatalf("fresh value equal to a strict threshold must not trigger: %#v", decision)
	}
}

func TestEvaluateSuppressesAllNonFreshEvidence(t *testing.T) {
	rule := Rule{Metric: "device.temperature_celsius", Operator: OperatorGreaterThan, Threshold: 82}
	for _, quality := range []domain.MetricQuality{
		domain.QualityStale,
		domain.QualityEstimated,
		domain.QualityExperimental,
	} {
		decision, err := Evaluate(rule, testMetric(quality, alertFloat(100)))
		if err != nil {
			t.Fatalf("evaluate %q evidence: %v", quality, err)
		}
		if decision.Triggered || !decision.Suppressed || decision.Reason != "ineligible_quality:"+string(quality) {
			t.Fatalf("non-fresh evidence must be suppressed for %q: %#v", quality, decision)
		}
	}

	for _, quality := range []domain.MetricQuality{domain.QualityUnavailable, domain.QualityUnsupported} {
		decision, err := Evaluate(rule, testMetric(quality, nil))
		if err != nil {
			t.Fatalf("evaluate %q evidence: %v", quality, err)
		}
		if decision.Triggered || !decision.Suppressed {
			t.Fatalf("non-numeric %q evidence must be suppressed: %#v", quality, decision)
		}
	}
}

func TestEvaluateRejectsUnsafeRuleInputs(t *testing.T) {
	metric := testMetric(domain.QualityFresh, alertFloat(83))
	for _, rule := range []Rule{
		{Metric: "other.metric", Operator: OperatorGreaterThan, Threshold: 82},
		{Metric: metric.Name, Operator: "unsupported", Threshold: 82},
	} {
		if _, err := Evaluate(rule, metric); err == nil {
			t.Fatalf("expected invalid rule %#v to fail", rule)
		}
	}
	if _, err := Evaluate(Rule{Metric: metric.Name, Operator: OperatorGreaterThan, Threshold: math.Inf(1)}, metric); err == nil || !strings.Contains(err.Error(), "finite") {
		t.Fatalf("expected non-finite threshold to fail, got %v", err)
	}
}
