package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type PGDumpExecutor struct{ Binary string }

func (executor PGDumpExecutor) Dump(ctx context.Context, connection Connection, scope []string, destination string) error {
	if connection.Host == "" || connection.Port == "" || connection.Database == "" || connection.User == "" || connection.PasswordFile == "" {
		return fmt.Errorf("backup connection host, port, database, user, and password file are required")
	}
	if info, err := os.Stat(connection.PasswordFile); err != nil || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("backup password file must exist and be private")
	}
	binary := executor.Binary
	if binary == "" {
		binary = "pg_dump"
	}
	arguments := []string{"--host", connection.Host, "--port", connection.Port, "--username", connection.User, "--format", "custom", "--no-owner", "--no-privileges", "--file", destination}
	arguments = append(arguments, scope...)
	arguments = append(arguments, connection.Database)
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Env = append(os.Environ(), "PGPASSFILE="+connection.PasswordFile)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %w: %s", err, string(output))
	}
	return nil
}
