package storage

import (
	"math"
	"testing"
)

func TestCapacityGovernorModes(t *testing.T) {
	policy := DefaultPolicy()
	tests := []struct {
		name      string
		usedBytes int64
		wantMode  Mode
		wantRaw   bool
	}{
		{name: "normal", usedBytes: 600, wantMode: ModeNormal, wantRaw: true},
		{name: "constrained", usedBytes: 700, wantMode: ModeConstrained, wantRaw: true},
		{name: "summary only", usedBytes: 850, wantMode: ModeSummaryOnly, wantRaw: false},
		{name: "protective", usedBytes: 950, wantMode: ModeProtective, wantRaw: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision, err := policy.Decide(test.usedBytes, 1000)
			if err != nil {
				t.Fatalf("decide: %v", err)
			}
			if decision.Mode != test.wantMode || decision.AcceptRawBatches != test.wantRaw {
				t.Fatalf("unexpected decision: %#v", decision)
			}
		})
	}
}

func TestCapacityGovernorRejectsInvalidInput(t *testing.T) {
	policy := DefaultPolicy()
	if _, err := policy.Decide(-1, 100); err == nil {
		t.Fatal("negative used bytes must be rejected")
	}
	if _, err := policy.Decide(1, 0); err == nil {
		t.Fatal("zero quota must be rejected")
	}
}

func TestCapacityGovernorRejectsNonFiniteThresholds(t *testing.T) {
	for name, policy := range map[string]Policy{
		"constrained NaN":              {ConstrainedPercent: math.NaN(), SummaryOnlyPercent: 85, ProtectivePercent: 95},
		"summary positive infinity":    {ConstrainedPercent: 70, SummaryOnlyPercent: math.Inf(1), ProtectivePercent: 95},
		"protective negative infinity": {ConstrainedPercent: 70, SummaryOnlyPercent: 85, ProtectivePercent: math.Inf(-1)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := policy.Decide(900, 1000); err == nil {
				t.Fatal("non-finite threshold must fail closed")
			}
		})
	}
}
