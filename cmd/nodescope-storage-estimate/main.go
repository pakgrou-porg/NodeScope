package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/storage"
)

type observationFile struct {
	HostID              string `json:"hostId"`
	ObservedDurationSec int64  `json:"observedDurationSeconds"`
	BatchCount          int64  `json:"batchCount"`
	CompressedByteCount int64  `json:"compressedByteCount"`
	MetricValueCount    int64  `json:"metricValueCount"`
}

func main() {
	inputPath := flag.String("input", "", "Path to a real NodeScope probe-summary JSON file")
	flag.Parse()
	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: nodescope-storage-estimate --input probe-summary.json")
		os.Exit(2)
	}

	contents, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read probe summary:", err)
		os.Exit(1)
	}
	var input observationFile
	if err := json.Unmarshal(contents, &input); err != nil {
		fmt.Fprintln(os.Stderr, "decode probe summary:", err)
		os.Exit(1)
	}
	estimate, err := storage.EstimateRetention(storage.ProbeObservation{
		HostID:              input.HostID,
		ObservedDuration:    time.Duration(input.ObservedDurationSec) * time.Second,
		BatchCount:          input.BatchCount,
		CompressedByteCount: input.CompressedByteCount,
		MetricValueCount:    input.MetricValueCount,
	}, storage.DefaultRetentionPlan())
	if err != nil {
		fmt.Fprintln(os.Stderr, "estimate retention:", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(estimate); err != nil {
		fmt.Fprintln(os.Stderr, "encode estimate:", err)
		os.Exit(1)
	}
}
