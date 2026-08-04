// Package enrollment provides administrator-workstation utilities for issuing and
// rotating NodeScope agent credentials without exposing raw credentials to SQL,
// logs, or ordinary server processes.
package enrollment

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Request struct {
	Slug           string
	DisplayName    string
	Platform       string
	Address        string
	CredentialPath string
	ExpiresAt      time.Time
}

type Result struct {
	HostID          string
	AgentID         string
	RotationVersion int
	CredentialPath  string
	CredentialHint  string
	ExpiresAt       time.Time
}

func (request Request) Validate() error {
	for label, value := range map[string]string{
		"host slug":         strings.TrimSpace(request.Slug),
		"display name":      strings.TrimSpace(request.DisplayName),
		"platform":          strings.TrimSpace(request.Platform),
		"host address":      strings.TrimSpace(request.Address),
		"credential output": strings.TrimSpace(request.CredentialPath),
	} {
		if value == "" {
			return fmt.Errorf("%s is required", label)
		}
	}
	if request.ExpiresAt.IsZero() || !request.ExpiresAt.After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("credential expiry must be at least one minute in the future")
	}
	return nil
}

// Enroll creates a random credential locally, stores only its SHA-256 digest via
// the schema-scoped SECURITY DEFINER function, then writes the raw token once to
// a new root-protected file. The plaintext is never returned, logged, or sent to
// Postgres.
func Enroll(ctx context.Context, pool *pgxpool.Pool, request Request) (Result, error) {
	if pool == nil {
		return Result{}, fmt.Errorf("enrollment database pool is required")
	}
	if err := request.Validate(); err != nil {
		return Result{}, err
	}
	credential, hint, digest, err := newCredential()
	if err != nil {
		return Result{}, err
	}
	var result Result
	result.CredentialPath = request.CredentialPath
	result.CredentialHint = hint
	result.ExpiresAt = request.ExpiresAt.UTC()
	if err := writeCredentialFile(request.CredentialPath, credential); err != nil {
		return Result{}, err
	}
	persisted := false
	defer func() {
		if !persisted {
			_ = os.Remove(request.CredentialPath)
		}
	}()

	row := pool.QueryRow(ctx, `select host_id::text, agent_id::text, rotation_version
		from nodescope.enroll_or_rotate_agent($1, $2, $3, $4::inet, $5, $6, $7)`,
		request.Slug,
		request.DisplayName,
		request.Platform,
		request.Address,
		digest[:],
		hint,
		result.ExpiresAt,
	)
	if err := row.Scan(&result.HostID, &result.AgentID, &result.RotationVersion); err != nil {
		return Result{}, fmt.Errorf("enroll NodeScope agent: %w", err)
	}
	persisted = true
	return result, nil
}

func newCredential() (credential string, hint string, digest [sha256.Size]byte, err error) {
	bytes := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, bytes); err != nil {
		return "", "", digest, fmt.Errorf("generate agent credential: %w", err)
	}
	credential = hex.EncodeToString(bytes)
	digest = sha256.Sum256([]byte(credential))
	hint = hex.EncodeToString(digest[:4])
	return credential, hint, digest, nil
}

func writeCredentialFile(path string, credential string) error {
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create credential output safely: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(credential + "\n"); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("write agent credential: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync agent credential: %w", err)
	}
	return nil
}
