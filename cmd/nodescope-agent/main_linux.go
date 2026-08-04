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
	state, err := agent.OpenSequenceStore(config.StateDirectory)
	if err != nil {
		logger.Error("open NodeScope agent state", "error", err)
		os.Exit(1)
	}
	sender, err := agent.NewSender(config)
	if err != nil {
		logger.Error("create NodeScope ingestion sender", "error", err)
		os.Exit(1)
	}
	collectors := []agent.Collector{agent.NewLinuxHostCollector(), agent.NewLinuxDRMCollector(), agent.NewNvidiaCollector(), agent.NewXDNACollector(), agent.NewSelectedProcessCollector(config.SelectedProcesses)}
	if config.ContainerInventoryEnabled {
		collectors = append(collectors, agent.NewDockerCollector(config.AlertedContainers))
	}
	runner, err := agent.NewRunner(config, collectors, sender, state)
	if err != nil {
		logger.Error("create NodeScope agent runner", "error", err)
		os.Exit(1)
	}

	context, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if *collectOnce {
		if err := runner.CollectOnce(context); err != nil {
			logger.Error("NodeScope agent collection failed", "error", err)
			os.Exit(1)
		}
		logger.Info("NodeScope agent collection succeeded")
		return
	}
	if err := runner.Run(context); err != nil && context.Err() == nil {
		logger.Error("NodeScope agent stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
