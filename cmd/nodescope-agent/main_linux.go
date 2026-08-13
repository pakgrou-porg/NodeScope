//go:build linux

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pakgrou-porg/nodescope/internal/agent"
)

func main() {
	preflightOnly := flag.Bool("preflight", false, "Print dependency and capability report without collecting telemetry")
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
		fmt.Fprintln(os.Stderr, "NodeScope agent configuration error:", err)
		os.Exit(2)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting NodeScope agent", "config", config.RedactedSummary())
	sender, err := agent.NewSender(config)
	if err != nil {
		logger.Error("create NodeScope ingestion sender", "error", err)
		os.Exit(1)
	}
	runContext, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
		logger.Error("open NodeScope agent state", "error", err)
		os.Exit(1)
	}
	collectors := []agent.Collector{agent.NewLinuxHostCollector(), agent.NewLinuxDRMCollector(), agent.NewNvidiaCollector(), agent.NewXDNACollector(), agent.NewSelectedProcessCollector(config.SelectedProcesses), agent.NewInferenceRuntimeProcessCollector(config.InferenceRuntimeProcesses), agent.NewInferenceRuntimeEndpointCollector(config.InferenceRuntimeEndpoints)}
	if config.ContainerInventoryEnabled {
		inventoryCollector, err := agent.NewInventoryProxyCollector(config)
		if err != nil {
			logger.Error("create NodeScope inventory proxy collector", "error", err)
			os.Exit(1)
		}
		collectors = append(collectors, inventoryCollector)
	}
	runner, err := agent.NewRunner(config, collectors, sender, state)
	if err != nil {
		logger.Error("create NodeScope agent runner", "error", err)
		os.Exit(1)
	}
	if *collectOnce {
		if err := runner.CollectOnce(runContext); err != nil {
			logger.Error("NodeScope agent collection failed", "error", err)
			os.Exit(1)
		}
		logger.Info("NodeScope agent collection succeeded")
		return
	}
	if err := runner.Run(runContext); err != nil && runContext.Err() == nil {
		logger.Error("NodeScope agent stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
