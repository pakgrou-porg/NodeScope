// Package proxy implements NodeScope's OpenAI-compatible telemetry proxy. It
// deliberately models only route, client, timing, status, and usage metadata;
// prompts and response content are never included in a persisted event.
package proxy

import (
	"context"
	"time"
)

type BackendRoute struct {
	ID           string `json:"id"`
	Model        string `json:"model"`
	PrimaryURL   string `json:"primary_url"`
	SecondaryURL string `json:"secondary_url,omitempty"`
	Enabled      bool   `json:"enabled"`
}

type RouteRegistry interface {
	Resolve(context.Context, string) (BackendRoute, error)
}

type UsageEvent struct {
	OccurredAt           time.Time
	RouteID              string
	Model                string
	ClientID             string
	BackendID            string
	StatusCode           int
	Streaming            bool
	TTFTMilliseconds     *int64
	DurationMilliseconds int64
	PromptTokens         *int64
	OutputTokens         *int64
	TotalTokens          *int64
	Outcome              string
}

type UsageRecorder interface {
	RecordUsage(context.Context, UsageEvent) error
}

// OperationalEvent is the only event shape that the proxy exposes to logging,
// tracing, audit, and support-export adapters. It intentionally has no field
// for inference text, headers, body bytes, tool arguments, or credentials.
type OperationalEvent struct {
	OccurredAt           time.Time
	RouteID              string
	Model                string
	ClientID             string
	BackendID            string
	StatusCode           int
	Streaming            bool
	TTFTMilliseconds     *int64
	DurationMilliseconds int64
	PromptTokens         *int64
	OutputTokens         *int64
	TotalTokens          *int64
	Outcome              string
}

type OperationalObserver interface {
	ObserveProxy(context.Context, OperationalEvent) error
}

func operationalEvent(event UsageEvent) OperationalEvent {
	return OperationalEvent{
		OccurredAt:           event.OccurredAt,
		RouteID:              event.RouteID,
		Model:                event.Model,
		ClientID:             event.ClientID,
		BackendID:            event.BackendID,
		StatusCode:           event.StatusCode,
		Streaming:            event.Streaming,
		TTFTMilliseconds:     event.TTFTMilliseconds,
		DurationMilliseconds: event.DurationMilliseconds,
		PromptTokens:         event.PromptTokens,
		OutputTokens:         event.OutputTokens,
		TotalTokens:          event.TotalTokens,
		Outcome:              event.Outcome,
	}
}
