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
	maxReceiptAgeSeconds := flag.Int("max-receipt-age-seconds", 15, "maximum server receipt age considered fresh")
	outputDir := flag.String("output-dir", "", "optional directory for an atomically written verification report")
	requireFresh := flag.Bool("require-fresh", false, "exit non-zero unless server receipt and metric state are fresh")
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
	reportValue := verificationReport{
		status:           value,
		GeneratedAt:      time.Now().UTC(),
		MaxReceiptAgeSec: *maxReceiptAgeSeconds,
		Assessment:       assessStatus(value, time.Now().UTC(), *maxReceiptAgeSeconds),
	}
	path, err := writeVerificationReportAtomic(*outputDir, reportValue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "write verifier report:", err)
		os.Exit(1)
	}
	reportValue.ReportPath = path
	if err := json.NewEncoder(os.Stdout).Encode(reportValue); err != nil {
		fmt.Fprintln(os.Stderr, "write verifier result:", err)
		os.Exit(1)
	}
	if *requireFresh && !reportValue.Assessment.Fresh {
		fmt.Fprintln(os.Stderr, "host verification is not fresh:", strings.Join(reportValue.Assessment.Reasons, ","))
		os.Exit(3)
	}
}
