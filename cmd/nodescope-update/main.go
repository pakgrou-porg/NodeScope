package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/update"
)

func main() {
	var version, archiveURL, checksumURL, stageDirectory, activeBinary, repository string
	var requireAttestation, rollback bool
	flag.StringVar(&version, "version", "", "pinned release version (for example v0.1.0)")
	flag.StringVar(&archiveURL, "archive-url", "", "HTTPS URL of the pinned native archive")
	flag.StringVar(&checksumURL, "checksum-url", "", "HTTPS URL of the matching SHA-256 checksum")
	flag.StringVar(&stageDirectory, "stage-dir", "/var/lib/nodescope/update", "root-owned staging directory")
	flag.StringVar(&activeBinary, "active-binary", "/usr/local/lib/nodescope/nodescope-agent", "active agent binary path")
	flag.StringVar(&repository, "repository", "pakgrou-porg/NodeScope", "GitHub repository used for provenance verification")
	flag.BoolVar(&requireAttestation, "require-attestation", true, "require gh attestation verification before activation")
	flag.BoolVar(&rollback, "rollback", false, "restore the previous verified agent binary")
	flag.Parse()

	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "nodescope-update must run as root")
		os.Exit(2)
	}
	if rollback {
		if err := update.Rollback(activeBinary); err != nil {
			fmt.Fprintln(os.Stderr, "rollback failed:", err)
			os.Exit(1)
		}
		fmt.Println("restored previous NodeScope agent binary")
		return
	}
	if version == "" || archiveURL == "" || checksumURL == "" {
		fmt.Fprintln(os.Stderr, "version, archive-url, and checksum-url are required")
		os.Exit(2)
	}
	if !strings.HasPrefix(version, "v") {
		fmt.Fprintln(os.Stderr, "version must be a pinned v-prefixed release tag")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	result, err := update.DownloadAndStage(ctx, nilSafeHTTPClient{}, update.Artifact{Version: version, ArchiveURL: archiveURL, ChecksumURL: checksumURL, BinaryName: "nodescope-agent"}, stageDirectory)
	if err != nil {
		fmt.Fprintln(os.Stderr, "verified staging failed:", err)
		os.Exit(1)
	}
	defer os.Remove(result.ArchivePath)
	if requireAttestation {
		if _, lookupErr := exec.LookPath("gh"); lookupErr != nil {
			fmt.Fprintln(os.Stderr, "GitHub CLI is required for attestation verification; install gh or pass --require-attestation=false only under an approved break-glass procedure")
			os.Exit(2)
		}
		command := exec.CommandContext(ctx, "gh", "attestation", "verify", result.ArchivePath, "-R", repository)
		command.Stdout, command.Stderr = os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "GitHub attestation verification failed:", err)
			os.Exit(1)
		}
	}
	if err := os.Chmod(result.StagedBinary, 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "set staged permissions:", err)
		os.Exit(1)
	}
	if _, err := update.Activate(result.StagedBinary, activeBinary); err != nil {
		fmt.Fprintln(os.Stderr, "activate verified release:", err)
		os.Exit(1)
	}
	fmt.Println("activated verified NodeScope agent", version, "at", filepath.Clean(activeBinary))
}

// nilSafeHTTPClient provides the default guarded HTTPS client without adding
// request redirects that could silently move a pinned release to another host.
type nilSafeHTTPClient struct{}

func (nilSafeHTTPClient) Do(request *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(request)
}
