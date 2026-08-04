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
	BackendURL           string
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
