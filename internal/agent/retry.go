package agent

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

// RetryClassified is implemented by delivery errors that explicitly describe
// whether retrying the same idempotent envelope can recover.
type RetryClassified interface {
	error
	Retryable() bool
}

// DeliveryError deliberately contains only a short public reason. It never
// retains endpoint credentials, request bodies, or upstream response content.
type DeliveryError struct {
	Reason     string
	CanRetry   bool
	StatusCode int
}

func (err *DeliveryError) Error() string {
	if err.StatusCode > 0 {
		return fmt.Sprintf("telemetry delivery failed: %s (status %d)", err.Reason, err.StatusCode)
	}
	return "telemetry delivery failed: " + err.Reason
}

func (err *DeliveryError) Retryable() bool { return err.CanRetry }

func isRetryable(err error) bool {
	var classified RetryClassified
	return errors.As(err, &classified) && classified.Retryable()
}

type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	JitterFraction float64
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: time.Second,
		MaxBackoff:     30 * time.Second,
		JitterFraction: 0.20,
	}
}

func (policy RetryPolicy) Validate() error {
	if policy.MaxAttempts < 1 || policy.MaxAttempts > 10 {
		return fmt.Errorf("retry max attempts must be from 1 to 10")
	}
	if policy.InitialBackoff <= 0 || policy.MaxBackoff < policy.InitialBackoff {
		return fmt.Errorf("retry backoff bounds are invalid")
	}
	if policy.JitterFraction < 0 || policy.JitterFraction > 1 {
		return fmt.Errorf("retry jitter fraction must be from 0 to 1")
	}
	return nil
}

// Delay returns full-jitter-bounded exponential backoff for a zero-based retry
// number. Jitter keeps independently recovering agents from stampeding replicas.
func (policy RetryPolicy) Delay(retryNumber int, random func() float64) time.Duration {
	if retryNumber < 0 {
		retryNumber = 0
	}
	base := float64(policy.InitialBackoff) * math.Pow(2, float64(retryNumber))
	if base > float64(policy.MaxBackoff) {
		base = float64(policy.MaxBackoff)
	}
	if random == nil {
		random = rand.Float64
	}
	jitter := (random()*2 - 1) * policy.JitterFraction
	return time.Duration(base * (1 + jitter))
}
