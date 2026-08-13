//go:build windows

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/pakgrou-porg/nodescope/internal/agent"
)

func main() {
	preflightOnly := flag.Bool("preflight", false, "Print capability report without collecting telemetry")
	ingestionPreflight := flag.Bool("ingestion-preflight", false, "Authenticate against an ingestion replica without transmitting telemetry")
	collectOnce := flag.Bool("once", false, "Collect and transmit one telemetry envelope, then exit")
	flag.Parse()

	if *preflightOnly {
		report := agent.InspectPreflight(nil)
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "encode preflight report:", err)
			os.Exit(1)
		}
		return
	}

	config, err := agent.LoadConfig(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "NodeScope Windows agent configuration error:", err)
		os.Exit(2)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting NodeScope Windows baseline agent", "config", config.RedactedSummary())
	sender, err := agent.NewSender(config)
	if err != nil {
		logger.Error("create NodeScope ingestion sender", "error", err)
		os.Exit(1)
	}
	runContext, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if *ingestionPreflight {
		result, err := sender.Preflight(runContext)
		if err != nil {
			logger.Error("NodeScope ingestion preflight failed", "error", err)
			os.Exit(1)
		}
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			logger.Error("encode NodeScope ingestion preflight", "error", err)
			os.Exit(1)
		}
		return
	}
	state, err := agent.OpenSequenceStore(config.StateDirectory)
	if err != nil {
		logger.Error("open NodeScope Windows agent state", "error", err)
		os.Exit(1)
	}
	runner, err := agent.NewRunner(config, []agent.Collector{agent.NewWindowsBaselineCollector()}, sender, state)
	if err != nil {
		logger.Error("create NodeScope Windows agent runner", "error", err)
		os.Exit(1)
	}
	if *collectOnce {
		if err := runner.CollectOnce(runContext); err != nil {
			logger.Error("NodeScope Windows baseline collection failed", "error", err)
			os.Exit(1)
		}
		logger.Info("NodeScope Windows baseline collection succeeded")
		return
	}
	if err := runner.Run(runContext); err != nil && runContext.Err() == nil {
		logger.Error("NodeScope Windows agent stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
