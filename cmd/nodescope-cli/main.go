package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/consoleclient"
)

func main() {
	format := flag.String("format", "table", "output format: table, json, or ndjson")
	endpoint := flag.String("endpoint", env("NODESCOPE_CONTROL_API_URL"), "HTTPS NodeScope control API base URL")
	credentialFile := flag.String("credential-file", env("NODESCOPE_CONTROL_API_CREDENTIAL_FILE"), "path to HTTPS API credential file")
	caFile := flag.String("ca-file", env("NODESCOPE_CONTROL_API_CA_FILE"), "path to PEM CA certificate for the HTTPS API")
	sshTarget := flag.String("ssh-target", env("NODESCOPE_SSH_TARGET"), "SSH target running a read-only nodescope-cli")
	sshCommand := flag.String("ssh-command", env("NODESCOPE_SSH_COMMAND"), "remote NodeScope CLI command; defaults to nodescope-cli")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTPS request timeout")
	flag.Parse()
	if *format != "table" && *format != "json" && *format != "ndjson" {
		fmt.Fprintln(os.Stderr, "format must be table, json, or ndjson")
		os.Exit(2)
	}
	hosts, err := consoleclient.LoadFleet(context.Background(), consoleclient.Config{
		DatabaseURL:    env("NODESCOPE_VERIFIER_DATABASE_URL"),
		Endpoint:       *endpoint,
		CredentialFile: *credentialFile,
		CAFile:         *caFile,
		SSHTarget:      *sshTarget,
		SSHCommand:     *sshCommand,
		HTTPTimeout:    *timeout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := writeOutput(os.Stdout, *format, hosts); err != nil {
		fmt.Fprintln(os.Stderr, "write fleet status:", err)
		os.Exit(1)
	}
}

func env(name string) string { return strings.TrimSpace(os.Getenv(name)) }

func writeOutput(writer io.Writer, format string, rows []consoleclient.HostStatus) error {
	switch format {
	case "json":
		return json.NewEncoder(writer).Encode(rows)
	case "ndjson":
		encoder := json.NewEncoder(writer)
		for _, item := range rows {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
		return nil
	case "table":
		renderTable(writer, rows)
		return nil
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func renderTable(output io.Writer, rows []consoleclient.HostStatus) {
	writer := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "HOST\tFRESHNESS\tAGE\tINTERVAL\tMETRICS\tUNAVAILABLE\tSTALE\tCLOCK OFFSET")
	for _, row := range rows {
		age := "unavailable"
		if row.LatestReceipt != nil {
			age = time.Since(*row.LatestReceipt).Round(time.Second).String()
		}
		interval := "n/a"
		if row.EffectiveIntervalSeconds > 0 {
			interval = fmt.Sprintf("%ds", row.EffectiveIntervalSeconds)
		}
		offset := "—"
		if row.ClockOffsetSeconds != nil {
			offset = fmt.Sprintf("%.1fs", *row.ClockOffsetSeconds)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\n", row.DisplayName, row.FreshnessState, age, interval, row.CurrentMetricCount, row.UnavailableMetricCount, row.StaleMetricCount, offset)
	}
	_ = writer.Flush()
}
