// Package ingest applies NodeScope telemetry safety controls before persistence.
package ingest

import (
	"fmt"
	"sync"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type Outcome string

const (
	OutcomeAccepted  Outcome = "accepted"
	OutcomeDuplicate Outcome = "duplicate"
	OutcomeThrottled Outcome = "throttled"
)

type Receipt struct {
	Outcome        Outcome
	IdempotencyKey string
	AcceptedAt     time.Time
}

type Policy struct {
	MaxRequestsPerMinute int
	Burst                int
	DeduplicationTTL     time.Duration
}

func DefaultPolicy() Policy {
	return Policy{
		MaxRequestsPerMinute: 120,
		Burst:                20,
		DeduplicationTTL:     48 * time.Hour,
	}
}

func (policy Policy) Validate() error {
	if policy.MaxRequestsPerMinute < 1 {
		return fmt.Errorf("max requests per minute must be positive")
	}
	if policy.Burst < 1 {
		return fmt.Errorf("burst must be positive")
	}
	if policy.DeduplicationTTL <= 0 {
		return fmt.Errorf("deduplication TTL must be positive")
	}
	return nil
}

type clock func() time.Time

type agentWindow struct {
	startedAt time.Time
	count     int
}

// Service is intentionally storage-agnostic. The first version uses the
// bounded in-memory maps below; production persistence supplies the same
// reservation semantics using a unique idempotency key in Supabase.
type Service struct {
	policy Policy
	now    clock
	mu     sync.Mutex
	seen   map[string]time.Time
	rates  map[string]agentWindow
}

func NewService(policy Policy) (*Service, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		policy: policy,
		now:    time.Now,
		seen:   make(map[string]time.Time),
		rates:  make(map[string]agentWindow),
	}, nil
}

func (service *Service) Accept(authenticatedAgentID string, envelope telemetry.Envelope) (Receipt, error) {
	if envelope.AgentID != authenticatedAgentID {
		return Receipt{}, fmt.Errorf("authenticated agent does not match envelope agent")
	}
	if err := envelope.Validate(); err != nil {
		return Receipt{}, err
	}

	now := service.now().UTC()
	key := envelope.IdempotencyKey()
	service.mu.Lock()
	defer service.mu.Unlock()
	service.prune(now)

	if _, duplicate := service.seen[key]; duplicate {
		return Receipt{Outcome: OutcomeDuplicate, IdempotencyKey: key, AcceptedAt: now}, nil
	}

	window := service.rates[authenticatedAgentID]
	if window.startedAt.IsZero() || now.Sub(window.startedAt) >= time.Minute {
		window = agentWindow{startedAt: now}
	}
	if window.count >= service.policy.MaxRequestsPerMinute+service.policy.Burst {
		service.rates[authenticatedAgentID] = window
		return Receipt{Outcome: OutcomeThrottled, IdempotencyKey: key, AcceptedAt: now}, nil
	}
	window.count++
	service.rates[authenticatedAgentID] = window
	service.seen[key] = now.Add(service.policy.DeduplicationTTL)
	return Receipt{Outcome: OutcomeAccepted, IdempotencyKey: key, AcceptedAt: now}, nil
}

func (service *Service) prune(now time.Time) {
	for key, expiresAt := range service.seen {
		if !expiresAt.After(now) {
			delete(service.seen, key)
		}
	}
}
