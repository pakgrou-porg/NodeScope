package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pakgrou-porg/nodescope/internal/backup"
)

func main() {
	var replicaID, outputDirectory, mode string
	var retentionDays int
	var host, port, database, user, passwordFile string
	flag.StringVar(&replicaID, "replica-id", "", "NodeScope replica identifier")
	flag.StringVar(&outputDirectory, "output-directory", "", "shared backup directory")
	flag.StringVar(&mode, "mode", "default", "default (configuration plus summaries) or full")
	flag.IntVar(&retentionDays, "retention-days", 10, "number of daily snapshots to retain")
	flag.StringVar(&host, "pg-host", "", "PostgreSQL host")
	flag.StringVar(&port, "pg-port", "5432", "PostgreSQL port")
	flag.StringVar(&database, "pg-database", "postgres", "PostgreSQL database")
	flag.StringVar(&user, "pg-user", "", "dedicated NodeScope backup user")
	flag.StringVar(&passwordFile, "pg-passfile", "", "private PostgreSQL password file")
	flag.Parse()
	if replicaID == "" || outputDirectory == "" || host == "" || user == "" || passwordFile == "" {
		fmt.Fprintln(os.Stderr, "replica-id, output-directory, pg-host, pg-user, and pg-passfile are required")
		os.Exit(2)
	}
	if mode != string(backup.ModeDefault) && mode != string(backup.ModeFull) {
		fmt.Fprintln(os.Stderr, "mode must be default or full")
		os.Exit(2)
	}
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "nodescope-backup must run as root to read its protected password file")
		os.Exit(2)
	}
	connectionString := (&url.URL{Scheme: "postgres", User: url.User(user), Host: net.JoinHostPort(host, port), Path: database}).String() + "?sslmode=verify-full&passfile=" + url.QueryEscape(passwordFile)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open backup lease database pool:", err)
		os.Exit(1)
	}
	defer pool.Close()
	runner := backup.Runner{Leaser: backup.NewPostgresLeaser(pool), Executor: backup.PGDumpExecutor{}}
	path, err := runner.Run(ctx, backup.Request{ReplicaID: replicaID, OutputDirectory: outputDirectory, Mode: backup.Mode(mode), RetentionDays: retentionDays, Connection: backup.Connection{Host: host, Port: port, Database: database, User: user, PasswordFile: passwordFile}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "NodeScope backup failed:", err)
		os.Exit(1)
	}
	fmt.Println(path)
}
