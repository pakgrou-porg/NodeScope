package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type hostStatus struct {
	DisplayName              string
	EffectiveIntervalSeconds int
	FreshnessState           string
	LatestReceipt            *time.Time
	CurrentMetricCount       int64
	UnavailableMetricCount   int64
	StaleMetricCount         int64
	ClockOffsetSeconds       *float64
}

type refreshMessage struct {
	rows []hostStatus
	err  error
}

func main() {
	refresh := flag.Duration("refresh", 5*time.Second, "automatic refresh interval from 1s to 60s")
	flag.Parse()
	if *refresh < time.Second || *refresh > 60*time.Second {
		fmt.Fprintln(os.Stderr, "refresh must be from 1s to 60s")
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

	commands := make(chan string, 1)
	go readCommands(commands)
	ticker := time.NewTicker(*refresh)
	defer ticker.Stop()
	var latest refreshMessage
	for {
		latest = readFleet(context.Background(), pool)
		render(latest, *refresh)
		select {
		case command, open := <-commands:
			if !open || command == "q" || command == "quit" || command == "exit" {
				return
			}
			if command == "r" || command == "refresh" || command == "" {
				continue
			}
		case <-ticker.C:
			continue
		}
	}
}

func readCommands(commands chan<- string) {
	defer close(commands)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commands <- strings.ToLower(strings.TrimSpace(scanner.Text()))
	}
}

func readFleet(ctx context.Context, pool *pgxpool.Pool) refreshMessage {
	query, err := pool.Query(ctx, `select display_name, effective_interval_seconds, freshness_state, latest_receipt,
		current_metric_count, unavailable_metric_count, stale_metric_count, clock_offset_seconds
		from nodescope.fleet_ingestion_status()`)
	if err != nil {
		return refreshMessage{err: err}
	}
	defer query.Close()
	result := make([]hostStatus, 0)
	for query.Next() {
		var item hostStatus
		if err := query.Scan(&item.DisplayName, &item.EffectiveIntervalSeconds, &item.FreshnessState, &item.LatestReceipt,
			&item.CurrentMetricCount, &item.UnavailableMetricCount, &item.StaleMetricCount, &item.ClockOffsetSeconds); err != nil {
			return refreshMessage{err: err}
		}
		result = append(result, item)
	}
	return refreshMessage{rows: result, err: query.Err()}
}

func render(message refreshMessage, refresh time.Duration) {
	fmt.Print("\033[H\033[2J")
	fmt.Println("NODESCOPE  ·  FLEET TERMINAL")
	fmt.Printf("Read-only verifier session · auto refresh %s · commands: r + Enter refresh, q + Enter quit\n\n", refresh)
	if message.err != nil {
		fmt.Println("FLEET STATUS UNAVAILABLE")
		fmt.Println("The verifier could not query the NodeScope status function. No substitute values are shown.")
		return
	}
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "HOST\tFRESHNESS\tLAST RECEIPT\tINTERVAL\tMETRICS\tUNAVAILABLE\tSTALE\tCLOCK")
	for _, item := range message.rows {
		age := "unavailable"
		if item.LatestReceipt != nil {
			age = time.Since(*item.LatestReceipt).Round(time.Second).String()
		}
		offset := "—"
		if item.ClockOffsetSeconds != nil {
			offset = fmt.Sprintf("%.1fs", *item.ClockOffsetSeconds)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%ds\t%d\t%d\t%d\t%s\n", item.DisplayName, item.FreshnessState, age, item.EffectiveIntervalSeconds, item.CurrentMetricCount, item.UnavailableMetricCount, item.StaleMetricCount, offset)
	}
	_ = writer.Flush()
	fmt.Println("\nFreshness uses server receipt time and each host's configured interval. Prompt and response content are never queried or rendered.")
}
