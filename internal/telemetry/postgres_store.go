package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pakgrou-porg/nodescope/internal/domain"
)

type AgentIdentity struct {
	AgentID string
	HostID  string
}

type PersistResult struct {
	Inserted    bool
	RawRetained bool
}

// PostgresStore persists only NodeScope schema data. It expects its pool to be
// configured by app.OpenRuntimeDatabase, which assumes nodescope_runtime on
// every connection before this code sees it.
type PostgresStore struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool, now: time.Now}
}

func (store *PostgresStore) AuthenticateAgent(ctx context.Context, bearerToken string) (AgentIdentity, error) {
	if bearerToken == "" {
		return AgentIdentity{}, fmt.Errorf("agent token is required")
	}
	digest := sha256.Sum256([]byte(bearerToken))
	var identity AgentIdentity
	err := store.pool.QueryRow(ctx, `
		select id::text, host_id::text
		from nodescope.agents
		where credential_digest = $1
		  and revoked_at is null
	`, digest[:]).Scan(&identity.AgentID, &identity.HostID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AgentIdentity{}, fmt.Errorf("unknown or revoked agent credential")
		}
		return AgentIdentity{}, fmt.Errorf("lookup agent credential: %w", err)
	}
	return identity, nil
}

func (store *PostgresStore) PersistEnvelope(ctx context.Context, identity AgentIdentity, envelope Envelope) (PersistResult, error) {
	if identity.AgentID == "" || identity.HostID == "" {
		return PersistResult{}, fmt.Errorf("agent identity is incomplete")
	}
	if envelope.AgentID != identity.AgentID || envelope.HostID != identity.HostID {
		return PersistResult{}, fmt.Errorf("envelope identity does not match authenticated agent")
	}
	if err := envelope.Validate(); err != nil {
		return PersistResult{}, err
	}

	now := store.now().UTC()
	expiresAt := now.Add(48 * time.Hour)
	transaction, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PersistResult{}, fmt.Errorf("begin telemetry transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	// Serialize a replay across raw and compact receipt tables. A capacity mode
	// change must never allow the same idempotency key to be accepted twice.
	if _, err := transaction.Exec(ctx, "select pg_advisory_xact_lock(hashtextextended($1, 0))", envelope.IdempotencyKey()); err != nil {
		return PersistResult{}, fmt.Errorf("lock telemetry idempotency key: %w", err)
	}
	var alreadyReceived bool
	if err := transaction.QueryRow(ctx, `
		select exists(select 1 from nodescope.telemetry_batches where idempotency_key = $1)
		    or exists(select 1 from nodescope.ingest_receipts where idempotency_key = $1)
	`, envelope.IdempotencyKey()).Scan(&alreadyReceived); err != nil {
		return PersistResult{}, fmt.Errorf("check telemetry idempotency: %w", err)
	}
	if alreadyReceived {
		return PersistResult{Inserted: false}, nil
	}

	rawRetained, err := store.acceptRawBatches(ctx, transaction)
	if err != nil {
		return PersistResult{}, err
	}
	if rawRetained {
		if err := store.insertRawBatch(ctx, transaction, identity, envelope, expiresAt); err != nil {
			return PersistResult{}, err
		}
	} else {
		if _, err := transaction.Exec(ctx, `
			insert into nodescope.ingest_receipts (idempotency_key, agent_id, host_id, expires_at, raw_retained)
			values ($1, $2, $3::uuid, $4, false)
		`, envelope.IdempotencyKey(), identity.AgentID, identity.HostID, expiresAt); err != nil {
			return PersistResult{}, fmt.Errorf("insert compact ingest receipt: %w", err)
		}
	}

	for _, sample := range envelope.Samples {
		if rawRetained {
			if err := store.insertRawSample(ctx, transaction, identity, envelope, sample, expiresAt); err != nil {
				return PersistResult{}, err
			}
		}
		if err := store.upsertLatest(ctx, transaction, identity, sample); err != nil {
			return PersistResult{}, err
		}
	}
	for _, container := range envelope.Containers {
		if err := store.upsertContainer(ctx, transaction, identity, container, envelope.ObservedAt.UTC()); err != nil {
			return PersistResult{}, err
		}
	}
	if err := store.upsertLatest(ctx, transaction, identity, clockOffsetSample(envelope.ObservedAt.UTC(), now)); err != nil {
		return PersistResult{}, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return PersistResult{}, fmt.Errorf("commit telemetry transaction: %w", err)
	}
	return PersistResult{Inserted: true, RawRetained: rawRetained}, nil
}

func clockOffsetSample(observedAt, receivedAt time.Time) Sample {
	offset := receivedAt.Sub(observedAt).Seconds()
	quality := domain.QualityFresh
	if math.Abs(offset) > 60 {
		quality = domain.QualityStale
	}
	return Sample{DeviceID: "agent-clock", Metric: domain.MetricValue{
		Name:       "agent.clock_offset_seconds",
		Unit:       "seconds",
		Value:      &offset,
		Quality:    quality,
		Source:     "nodescope-server",
		Semantics:  "server receipt time minus agent observation time; stale when absolute offset exceeds 60 seconds",
		ObservedAt: receivedAt,
	}}
}

func (store *PostgresStore) acceptRawBatches(ctx context.Context, transaction pgx.Tx) (bool, error) {
	var acceptRaw bool
	err := transaction.QueryRow(ctx, `select accept_raw_batches from nodescope.capacity_status where singleton = true`).Scan(&acceptRaw)
	if err == pgx.ErrNoRows {
		return rawRetentionAllowed(false, false), nil
	}
	if err != nil {
		return false, fmt.Errorf("read capacity state: %w", err)
	}
	return rawRetentionAllowed(true, acceptRaw), nil
}

// rawRetentionAllowed keeps the admission default fail-conservative: the
// explicit singleton capacity record must exist and grant raw retention.
// Latest-state and compact idempotency receipts remain available otherwise.
func rawRetentionAllowed(statusPresent, acceptRaw bool) bool {
	return statusPresent && acceptRaw
}

func (store *PostgresStore) insertRawBatch(ctx context.Context, transaction pgx.Tx, identity AgentIdentity, envelope Envelope, expiresAt time.Time) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode telemetry payload: %w", err)
	}
	_, err = transaction.Exec(ctx, `
		insert into nodescope.telemetry_batches (
			idempotency_key, agent_id, host_id, boot_id, sequence, observed_at,
			expires_at, compressed_bytes, metric_value_count, payload
		) values ($1, $2, $3::uuid, $4, $5, $6, $7, $8, $9, $10::jsonb)
	`, envelope.IdempotencyKey(), identity.AgentID, identity.HostID, envelope.BootID,
		envelope.Sequence, envelope.ObservedAt.UTC(), expiresAt,
		envelope.CompressedBytes, envelope.MetricValueCount, payload)
	if err != nil {
		return fmt.Errorf("insert telemetry batch: %w", err)
	}
	return nil
}

func (store *PostgresStore) insertRawSample(ctx context.Context, transaction pgx.Tx, identity AgentIdentity, envelope Envelope, sample Sample, expiresAt time.Time) error {
	metric := sample.Metric
	_, err := transaction.Exec(ctx, `
		insert into nodescope.metric_samples (
			batch_id, host_id, device_id, metric_name, observed_at, numeric_value,
			quality, source, semantics, expires_at
		) values (
			(select id from nodescope.telemetry_batches where idempotency_key = $1),
			$2::uuid, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`, envelope.IdempotencyKey(), identity.HostID, sample.DeviceID, metric.Name, metric.ObservedAt.UTC(), metric.Value,
		string(metric.Quality), metric.Source, metric.Semantics, expiresAt)
	if err != nil {
		return fmt.Errorf("insert raw metric sample %s: %w", metric.Name, err)
	}
	return nil
}

func (store *PostgresStore) upsertContainer(ctx context.Context, transaction pgx.Tx, identity AgentIdentity, container ContainerInventory, observedAt time.Time) error {
	_, err := transaction.Exec(ctx, `
		insert into nodescope.container_inventory (
			host_id, container_id, name, image, state, health, selected_for_alerting, observed_at
		) values ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		on conflict (host_id, container_id) do update
		set name = excluded.name,
			image = excluded.image,
			state = excluded.state,
			health = excluded.health,
			selected_for_alerting = excluded.selected_for_alerting,
			observed_at = excluded.observed_at,
			updated_at = now()
		where excluded.observed_at >= nodescope.container_inventory.observed_at
	`, identity.HostID, container.ContainerID, container.Name, container.Image, container.State, container.Health, container.SelectedForAlerting, observedAt)
	if err != nil {
		return fmt.Errorf("upsert container inventory: %w", err)
	}
	return nil
}

func (store *PostgresStore) upsertLatest(ctx context.Context, transaction pgx.Tx, identity AgentIdentity, sample Sample) error {
	metric := sample.Metric
	_, err := transaction.Exec(ctx, `
		insert into nodescope.metric_latest (
			host_id, device_id, metric_name, observed_at, numeric_value, quality, source, semantics
		) values ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		on conflict (host_id, device_id, metric_name) do update
		set observed_at = excluded.observed_at,
			received_at = now(),
			numeric_value = excluded.numeric_value,
			text_value = null,
			quality = excluded.quality,
			source = excluded.source,
			semantics = excluded.semantics
		where excluded.observed_at >= nodescope.metric_latest.observed_at
	`, identity.HostID, sample.DeviceID, metric.Name, metric.ObservedAt.UTC(), metric.Value,
		string(metric.Quality), metric.Source, metric.Semantics)
	if err != nil {
		return fmt.Errorf("upsert latest metric %s: %w", metric.Name, err)
	}
	return nil
}
