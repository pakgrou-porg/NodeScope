package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseHealth interface {
	Ping(context.Context) error
	Close()
}

type RuntimeDatabase struct {
	pool *pgxpool.Pool
}

func OpenRuntimeDatabase(ctx context.Context, databaseURL string) (*RuntimeDatabase, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("runtime database URL is required")
	}
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse runtime database URL: %w", err)
	}
	config.MaxConns = 8
	config.MinConns = 0
	config.MaxConnIdleTime = 2 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		if _, err := connection.Exec(ctx, "set role nodescope_runtime"); err != nil {
			return fmt.Errorf("assume nodescope runtime role: %w", err)
		}
		if _, err := connection.Exec(ctx, "set search_path = nodescope, pg_catalog"); err != nil {
			return fmt.Errorf("set runtime search path: %w", err)
		}
		return nil
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open runtime database pool: %w", err)
	}
	return &RuntimeDatabase{pool: pool}, nil
}

func (database *RuntimeDatabase) Ping(ctx context.Context) error {
	return database.pool.Ping(ctx)
}

// Pool is exposed only to internal NodeScope packages. Every acquired
// connection has already assumed nodescope_runtime in AfterConnect.
func (database *RuntimeDatabase) Pool() *pgxpool.Pool {
	return database.pool
}

func (database *RuntimeDatabase) Close() {
	if database != nil && database.pool != nil {
		database.pool.Close()
	}
}
