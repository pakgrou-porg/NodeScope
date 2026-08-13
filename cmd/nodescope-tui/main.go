package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/consoleclient"
)

type refreshMessage struct {
	rows []consoleclient.HostStatus
	err  error
}

func main() {
	refresh := flag.Duration("refresh", 5*time.Second, "automatic refresh interval from 1s to 60s")
	endpoint := flag.String("endpoint", env("NODESCOPE_CONTROL_API_URL"), "HTTPS NodeScope control API base URL")
	credentialFile := flag.String("credential-file", env("NODESCOPE_CONTROL_API_CREDENTIAL_FILE"), "path to HTTPS API credential file")
	caFile := flag.String("ca-file", env("NODESCOPE_CONTROL_API_CA_FILE"), "path to PEM CA certificate for the HTTPS API")
	sshTarget := flag.String("ssh-target", env("NODESCOPE_SSH_TARGET"), "SSH target running a read-only nodescope-cli")
	sshCommand := flag.String("ssh-command", env("NODESCOPE_SSH_COMMAND"), "remote NodeScope CLI command; defaults to nodescope-cli")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTPS request timeout")
	flag.Parse()
	if *refresh < time.Second || *refresh > 60*time.Second {
		fmt.Fprintln(os.Stderr, "refresh must be from 1s to 60s")
		os.Exit(2)
	}
	config := consoleclient.Config{
		DatabaseURL:    env("NODESCOPE_VERIFIER_DATABASE_URL"),
		Endpoint:       *endpoint,
		CredentialFile: *credentialFile,
		CAFile:         *caFile,
		SSHTarget:      *sshTarget,
		SSHCommand:     *sshCommand,
		HTTPTimeout:    *timeout,
	}
	if err := config.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	commands := make(chan string, 1)
	go readCommands(commands)
	ticker := time.NewTicker(*refresh)
	defer ticker.Stop()
	var latest refreshMessage
	for {
		latest = readFleet(context.Background(), config)
		render(os.Stdout, latest, *refresh)
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

func env(name string) string { return strings.TrimSpace(os.Getenv(name)) }

func readCommands(commands chan<- string) {
	defer close(commands)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commands <- strings.ToLower(strings.TrimSpace(scanner.Text()))
	}
}

func readFleet(ctx context.Context, config consoleclient.Config) refreshMessage {
	rows, err := consoleclient.LoadFleet(ctx, config)
	return refreshMessage{rows: rows, err: err}
}

func render(output io.Writer, message refreshMessage, refresh time.Duration) {
	fmt.Fprint(output, "\033[H\033[2J")
	fmt.Fprintln(output, "NODESCOPE  ·  FLEET TERMINAL")
	fmt.Fprintf(output, "Read-only verifier session · auto refresh %s · commands: r + Enter refresh, q + Enter quit\n\n", refresh)
	if message.err != nil {
		fmt.Fprintln(output, "FLEET STATUS UNAVAILABLE")
		fmt.Fprintln(output, "The configured read path could not query NodeScope status. No substitute values are shown.")
		return
	}
	writer := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "HOST\tFRESHNESS\tLAST RECEIPT\tINTERVAL\tMETRICS\tUNAVAILABLE\tSTALE\tCLOCK")
	for _, item := range message.rows {
		age := "unavailable"
		if item.LatestReceipt != nil {
			age = time.Since(*item.LatestReceipt).Round(time.Second).String()
		}
		interval := "n/a"
		if item.EffectiveIntervalSeconds > 0 {
			interval = fmt.Sprintf("%ds", item.EffectiveIntervalSeconds)
		}
		offset := "—"
		if item.ClockOffsetSeconds != nil {
			offset = fmt.Sprintf("%.1fs", *item.ClockOffsetSeconds)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\n", item.DisplayName, item.FreshnessState, age, interval, item.CurrentMetricCount, item.UnavailableMetricCount, item.StaleMetricCount, offset)
	}
	_ = writer.Flush()
	fmt.Fprintln(output, "\nFreshness uses server receipt time. Prompt and response content are never queried or rendered.")
}
