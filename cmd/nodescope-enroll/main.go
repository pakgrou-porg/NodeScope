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
	"github.com/pakgrou-porg/nodescope/internal/enrollment"
)

func main() {
	var request enrollment.Request
	var expiresInDays int
	flag.StringVar(&request.Slug, "slug", "", "stable lowercase NodeScope host slug")
	flag.StringVar(&request.DisplayName, "display-name", "", "human-readable host name")
	flag.StringVar(&request.Platform, "platform", "", "platform label such as fedora or dgx-os")
	flag.StringVar(&request.Address, "address", "", "static host IP address")
	flag.StringVar(&request.CredentialPath, "credential-output", "", "new root-protected credential file path; must not exist")
	flag.IntVar(&expiresInDays, "expires-in-days", 90, "credential lifetime from 1 to 365 days")
	flag.Parse()

	if expiresInDays < 1 || expiresInDays > 365 {
		fmt.Fprintln(os.Stderr, "expires-in-days must be from 1 to 365")
		os.Exit(2)
	}
	databaseURL := strings.TrimSpace(os.Getenv("NODESCOPE_ENROLLER_DATABASE_URL"))
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "NODESCOPE_ENROLLER_DATABASE_URL is required")
		os.Exit(2)
	}
	request.ExpiresAt = time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour)
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open enrollment database connection:", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "verify enrollment database connection:", err)
		os.Exit(1)
	}
	result, err := enrollment.Enroll(context.Background(), pool, request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "enroll NodeScope agent:", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, "write enrollment result:", err)
		os.Exit(1)
	}
}
