package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresService struct{ pool *pgxpool.Pool }

func NewPostgresService(pool *pgxpool.Pool) *PostgresService { return &PostgresService{pool: pool} }

func (service *PostgresService) FleetStatus(ctx context.Context, _ Principal) ([]FleetHost, error) {
	rows, err := service.pool.Query(ctx, `select host_slug, display_name, platform, freshness_state, latest_receipt, current_metric_count, unavailable_metric_count, stale_metric_count, clock_offset_seconds, clock_offset_quality from nodescope.fleet_ingestion_status()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]FleetHost, 0)
	for rows.Next() {
		var item FleetHost
		if err := rows.Scan(&item.ID, &item.Name, &item.Platform, &item.Freshness, &item.LatestServerReceipt, &item.MetricCount, &item.UnavailableMetricCount, &item.StaleMetricCount, &item.ClockOffsetSeconds, &item.ClockOffsetQuality); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (service *PostgresService) AcknowledgeAlert(ctx context.Context, principal Principal, alertID, note string) error {
	var auditID string
	if err := service.pool.QueryRow(ctx, `select nodescope.mcp_acknowledge_alert($1, $2, $3)::text`, principal.ID, alertID, note).Scan(&auditID); err != nil {
		return fmt.Errorf("acknowledge alert: %w", err)
	}
	return nil
}

func (service *PostgresService) SetCollectionInterval(ctx context.Context, principal Principal, hostID string, seconds int) error {
	var auditID string
	if err := service.pool.QueryRow(ctx, `select nodescope.mcp_set_collection_interval($1, $2, $3)::text`, principal.ID, hostID, seconds).Scan(&auditID); err != nil {
		return fmt.Errorf("set collection interval: %w", err)
	}
	return nil
}

func (service *PostgresService) RefreshStorageBaseline(ctx context.Context, principal Principal, hostID string, _ bool) error {
	var operationID string
	if err := service.pool.QueryRow(ctx, `select nodescope.mcp_refresh_storage_baseline($1, $2)::text`, principal.ID, hostID).Scan(&operationID); err != nil {
		return fmt.Errorf("refresh storage baseline: %w", err)
	}
	return nil
}

// Compile-time guard that time is retained in the public FleetHost contract.
var _ = time.Time{}
