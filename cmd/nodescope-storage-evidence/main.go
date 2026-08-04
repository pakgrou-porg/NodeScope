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

type evidence struct {
	HostSlug               string     `json:"host_slug"`
	FirstReceivedAt        *time.Time `json:"first_received_at"`
	LastReceivedAt         *time.Time `json:"last_received_at"`
	ReceivedBatchCount     int64      `json:"received_batch_count"`
	ExpectedBatchCount     int64      `json:"expected_batch_count"`
	CompletenessPercent    float64    `json:"completeness_percent"`
	MaxGapSeconds          float64    `json:"max_gap_seconds"`
	MedianCompressedBytes  *float64   `json:"median_compressed_bytes"`
	P95CompressedBytes     *float64   `json:"p95_compressed_bytes"`
	P99CompressedBytes     *float64   `json:"p99_compressed_bytes"`
	TotalCompressedBytes   int64      `json:"total_compressed_bytes"`
	MetricCardinality      int64      `json:"metric_cardinality"`
	TelemetryRelationBytes int64      `json:"telemetry_relation_bytes"`
	TelemetryIndexBytes    int64      `json:"telemetry_index_bytes"`
	RawSampleRelationBytes int64      `json:"raw_sample_relation_bytes"`
	RawSampleIndexBytes    int64      `json:"raw_sample_index_bytes"`
}

func main() {
	slug := flag.String("slug", "", "NodeScope host slug")
	since := flag.String("since", "", "RFC3339 UTC start time for the evidence window")
	flag.Parse()
	if strings.TrimSpace(*slug) == "" || strings.TrimSpace(*since) == "" {
		fmt.Fprintln(os.Stderr, "slug and since are required")
		os.Exit(2)
	}
	start, err := time.Parse(time.RFC3339, *since)
	if err != nil {
		fmt.Fprintln(os.Stderr, "since must use RFC3339 format:", err)
		os.Exit(2)
	}
	databaseURL := strings.TrimSpace(os.Getenv("NODESCOPE_STORAGE_AUDITOR_DATABASE_URL"))
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "NODESCOPE_STORAGE_AUDITOR_DATABASE_URL is required")
		os.Exit(2)
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open storage-auditor database connection:", err)
		os.Exit(1)
	}
	defer pool.Close()
	var value evidence
	err = pool.QueryRow(context.Background(), `select host_slug, first_received_at, last_received_at, received_batch_count,
		expected_batch_count, completeness_percent, max_gap_seconds, median_compressed_bytes, p95_compressed_bytes,
		p99_compressed_bytes, total_compressed_bytes, metric_cardinality, telemetry_relation_bytes, telemetry_index_bytes,
		raw_sample_relation_bytes, raw_sample_index_bytes from nodescope.storage_probe_evidence($1, $2)`, *slug, start).Scan(
		&value.HostSlug, &value.FirstReceivedAt, &value.LastReceivedAt, &value.ReceivedBatchCount,
		&value.ExpectedBatchCount, &value.CompletenessPercent, &value.MaxGapSeconds, &value.MedianCompressedBytes,
		&value.P95CompressedBytes, &value.P99CompressedBytes, &value.TotalCompressedBytes, &value.MetricCardinality,
		&value.TelemetryRelationBytes, &value.TelemetryIndexBytes, &value.RawSampleRelationBytes, &value.RawSampleIndexBytes,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read storage evidence:", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, "write storage evidence:", err)
		os.Exit(1)
	}
}
