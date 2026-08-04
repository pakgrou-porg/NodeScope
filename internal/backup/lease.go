package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Lease struct {
	FencingToken int64
	ExpiresAt    time.Time
}
type Leaser interface {
	Acquire(context.Context, string, string, time.Duration) (Lease, error)
	Current(context.Context, string, string, int64) (bool, error)
}
type PostgresLeaser struct{ pool *pgxpool.Pool }

func NewPostgresLeaser(pool *pgxpool.Pool) *PostgresLeaser { return &PostgresLeaser{pool: pool} }
func (leaser *PostgresLeaser) Acquire(ctx context.Context, name, replicaID string, ttl time.Duration) (Lease, error) {
	var lease Lease
	if ttl < 10*time.Second || ttl > time.Hour {
		return Lease{}, fmt.Errorf("lease TTL must be from 10 seconds through one hour")
	}
	err := leaser.pool.QueryRow(ctx, `select fencing_token, expires_at from nodescope.acquire_maintenance_lease($1, $2, $3)`, name, replicaID, int(ttl.Seconds())).Scan(&lease.FencingToken, &lease.ExpiresAt)
	if err != nil {
		return Lease{}, err
	}
	return lease, nil
}
func (leaser *PostgresLeaser) Current(ctx context.Context, name, replicaID string, token int64) (bool, error) {
	var current bool
	err := leaser.pool.QueryRow(ctx, `select nodescope.lease_is_current($1, $2, $3)`, name, replicaID, token).Scan(&current)
	return current, err
}
