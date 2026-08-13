package proxy

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRecorder struct {
	pool *pgxpool.Pool
}

func NewPostgresRecorder(pool *pgxpool.Pool) *PostgresRecorder { return &PostgresRecorder{pool: pool} }

func (recorder *PostgresRecorder) RecordUsage(ctx context.Context, event UsageEvent) error {
	_, err := recorder.pool.Exec(ctx, `
		insert into nodescope.inference_usage_events (
			occurred_at, route_id, model, client_id, backend_url, status_code, streaming,
			ttft_milliseconds, duration_milliseconds, prompt_tokens, output_tokens, total_tokens, outcome
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, event.OccurredAt, event.RouteID, event.Model, event.ClientID, event.BackendID, event.StatusCode, event.Streaming,
		event.TTFTMilliseconds, event.DurationMilliseconds, event.PromptTokens, event.OutputTokens, event.TotalTokens, event.Outcome)
	return err
}
