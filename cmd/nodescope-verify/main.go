package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type status struct {
	HostSlug               string     `json:"host_slug"`
	LatestReceipt          *time.Time `json:"latest_receipt"`
	CurrentMetricCount     int64      `json:"current_metric_count"`
	UnavailableMetricCount int64      `json:"unavailable_metric_count"`
	StaleMetricCount       int64      `json:"stale_metric_count"`
	ClockOffsetSeconds     *float64   `json:"clock_offset_seconds"`
	ClockOffsetQuality     *string    `json:"clock_offset_quality"`
	ClockOffsetSource      *string    `json:"clock_offset_source"`
	ClockObservedAt        *time.Time `json:"clock_observed_at"`
	ClockReceivedAt        *time.Time `json:"clock_received_at"`
}

func main() {
	slug := flag.String("slug", "", "NodeScope host slug")
	flag.Parse()
	if strings.TrimSpace(*slug) == "" {
		fmt.Fprintln(os.Stderr, "slug is required")
		os.Exit(2)
	}
	databaseURL := strings.TrimSpace(os.Getenv("NODESCOPE_VERIFIER_DATABASE_URL"))
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "NODESCOPE_VERIFIER_DATABASE_URL is required")
		os.Exit(2)
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open verifier database connection:", err)
		os.Exit(1)
	}
	defer pool.Close()
	var value status
	err = pool.QueryRow(context.Background(), `select host_slug, latest_receipt, current_metric_count, unavailable_metric_count,
		stale_metric_count, clock_offset_seconds, clock_offset_quality, clock_offset_source, clock_observed_at, clock_received_at
		from nodescope.host_ingestion_status($1)`, *slug).Scan(
		&value.HostSlug, &value.LatestReceipt, &value.CurrentMetricCount, &value.UnavailableMetricCount,
		&value.StaleMetricCount, &value.ClockOffsetSeconds, &value.ClockOffsetQuality, &value.ClockOffsetSource,
		&value.ClockObservedAt, &value.ClockReceivedAt,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read host ingestion status:", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, "write verifier result:", err)
		os.Exit(1)
	}
}
