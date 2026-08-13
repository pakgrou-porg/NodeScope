package proxy

import (
	"context"
	"fmt"
	"sync"
)

type MemoryRegistry struct {
	mu     sync.RWMutex
	routes map[string]BackendRoute
}

func NewMemoryRegistry(routes []BackendRoute) *MemoryRegistry {
	registry := &MemoryRegistry{routes: make(map[string]BackendRoute, len(routes))}
	for _, route := range routes {
		registry.routes[route.Model] = route
	}
	return registry
}

func (registry *MemoryRegistry) Resolve(_ context.Context, model string) (BackendRoute, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	route, ok := registry.routes[model]
	if !ok || !route.Enabled {
		return BackendRoute{}, fmt.Errorf("approved route for model %q is unavailable", model)
	}
	return route, nil
}

type MemoryRecorder struct {
	mu     sync.Mutex
	events []UsageEvent
}

func (recorder *MemoryRecorder) RecordUsage(_ context.Context, event UsageEvent) error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.events = append(recorder.events, event)
	return nil
}

func (recorder *MemoryRecorder) Events() []UsageEvent {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return append([]UsageEvent(nil), recorder.events...)
}

// MemoryOperationalObserver supports adversarial tests for logging, tracing,
// audit, and support-export adapters that consume the safe event contract.
type MemoryOperationalObserver struct {
	mu     sync.Mutex
	events []OperationalEvent
}

func (observer *MemoryOperationalObserver) ObserveProxy(_ context.Context, event OperationalEvent) error {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.events = append(observer.events, event)
	return nil
}

func (observer *MemoryOperationalObserver) Events() []OperationalEvent {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	return append([]OperationalEvent(nil), observer.events...)
}
