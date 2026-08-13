package telemetry

import (
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func TestRawRetentionAllowedFailsConservativeWithoutCapacityStatus(t *testing.T) {
	for name, input := range map[string]struct {
		statusPresent bool
		acceptRaw     bool
		want          bool
	}{
		"missing status":        {statusPresent: false, acceptRaw: true, want: false},
		"explicitly protective": {statusPresent: true, acceptRaw: false, want: false},
		"explicitly accepting":  {statusPresent: true, acceptRaw: true, want: true},
	} {
		t.Run(name, func(t *testing.T) {
			if got := rawRetentionAllowed(input.statusPresent, input.acceptRaw); got != input.want {
				t.Fatalf("rawRetentionAllowed(%t, %t) = %t, want %t", input.statusPresent, input.acceptRaw, got, input.want)
			}
		})
	}
}

func TestShouldReplaceLatestPrefersNewerThenHigherQualityEvidence(t *testing.T) {
	observedAt := time.Date(2026, 8, 13, 22, 0, 0, 0, time.UTC)
	for name, input := range map[string]struct {
		existingAt      time.Time
		existingQuality domain.MetricQuality
		incomingAt      time.Time
		incomingQuality domain.MetricQuality
		want            bool
	}{
		"newer unavailable state replaces prior fresh state": {
			existingAt: observedAt, existingQuality: domain.QualityFresh,
			incomingAt: observedAt.Add(time.Second), incomingQuality: domain.QualityUnavailable,
			want: true,
		},
		"older fresh state cannot replace newer stale state": {
			existingAt: observedAt.Add(time.Second), existingQuality: domain.QualityStale,
			incomingAt: observedAt, incomingQuality: domain.QualityFresh,
			want: false,
		},
		"equal timestamp fresh state resists stale replacement": {
			existingAt: observedAt, existingQuality: domain.QualityFresh,
			incomingAt: observedAt, incomingQuality: domain.QualityStale,
			want: false,
		},
		"equal timestamp fresh state resists unavailable replacement": {
			existingAt: observedAt, existingQuality: domain.QualityFresh,
			incomingAt: observedAt, incomingQuality: domain.QualityUnavailable,
			want: false,
		},
		"equal timestamp fresh state resists estimated replacement": {
			existingAt: observedAt, existingQuality: domain.QualityFresh,
			incomingAt: observedAt, incomingQuality: domain.QualityEstimated,
			want: false,
		},
		"equal timestamp fresh state resists experimental replacement": {
			existingAt: observedAt, existingQuality: domain.QualityFresh,
			incomingAt: observedAt, incomingQuality: domain.QualityExperimental,
			want: false,
		},
		"equal timestamp fresh evidence upgrades stale state": {
			existingAt: observedAt, existingQuality: domain.QualityStale,
			incomingAt: observedAt, incomingQuality: domain.QualityFresh,
			want: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := shouldReplaceLatest(input.existingAt, input.existingQuality, input.incomingQuality, input.incomingAt); got != input.want {
				t.Fatalf("shouldReplaceLatest(existing=%s/%s, incoming=%s/%s) = %t, want %t", input.existingAt, input.existingQuality, input.incomingAt, input.incomingQuality, got, input.want)
			}
		})
	}
}
