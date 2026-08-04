package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type hostStatus struct {
	HostSlug                 string     `json:"host_slug"`
	DisplayName              string     `json:"display_name"`
	Platform                 string     `json:"platform"`
	EffectiveIntervalSeconds int        `json:"effective_interval_seconds"`
	FreshnessState           string     `json:"freshness_state"`
	LatestReceipt            *time.Time `json:"latest_receipt"`
	CurrentMetricCount       int64      `json:"current_metric_count"`
	UnavailableMetricCount   int64      `json:"unavailable_metric_count"`
	StaleMetricCount         int64      `json:"stale_metric_count"`
	ClockOffsetSeconds       *float64   `json:"clock_offset_seconds"`
	ClockOffsetQuality       *string    `json:"clock_offset_quality"`
}

func main() {
	format := flag.String("format", "table", "output format: table, json, or ndjson")
	flag.Parse()
	if *format != "table" && *format != "json" && *format != "ndjson" {
		fmt.Fprintln(os.Stderr, "format must be table, json, or ndjson")
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
	rows, err := pool.Query(context.Background(), `select host_slug, display_name, platform, effective_interval_seconds, freshness_state,
		latest_receipt, current_metric_count, unavailable_metric_count, stale_metric_count, clock_offset_seconds, clock_offset_quality
		from nodescope.fleet_ingestion_status()`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "query fleet status:", err)
		os.Exit(1)
	}
	defer rows.Close()
	result := make([]hostStatus, 0)
	for rows.Next() {
		var item hostStatus
		if err := rows.Scan(&item.HostSlug, &item.DisplayName, &item.Platform, &item.EffectiveIntervalSeconds, &item.FreshnessState,
			&item.LatestReceipt, &item.CurrentMetricCount, &item.UnavailableMetricCount, &item.StaleMetricCount, &item.ClockOffsetSeconds, &item.ClockOffsetQuality); err != nil {
			fmt.Fprintln(os.Stderr, "scan fleet status:", err)
			os.Exit(1)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "read fleet status:", err)
		os.Exit(1)
	}
	switch *format {
	case "json":
		_ = json.NewEncoder(os.Stdout).Encode(result)
	case "ndjson":
		encoder := json.NewEncoder(os.Stdout)
		for _, item := range result {
			if err := encoder.Encode(item); err != nil {
				os.Exit(1)
			}
		}
	default:
		renderTable(result)
	}
}

func renderTable(rows []hostStatus) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "HOST\tFRESHNESS\tAGE\tINTERVAL\tMETRICS\tUNAVAILABLE\tSTALE\tCLOCK OFFSET")
	for _, row := range rows {
		age := "unavailable"
		if row.LatestReceipt != nil {
			age = time.Since(*row.LatestReceipt).Round(time.Second).String()
		}
		offset := "—"
		if row.ClockOffsetSeconds != nil {
			offset = fmt.Sprintf("%.1fs", *row.ClockOffsetSeconds)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%ds\t%d\t%d\t%d\t%s\n", row.DisplayName, row.FreshnessState, age, row.EffectiveIntervalSeconds, row.CurrentMetricCount, row.UnavailableMetricCount, row.StaleMetricCount, offset)
	}
	_ = writer.Flush()
}
