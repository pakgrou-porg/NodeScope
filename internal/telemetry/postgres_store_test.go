package telemetry

import "testing"

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
