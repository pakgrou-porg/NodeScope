// Package consoleclient loads metadata-only fleet status for the native
// NodeScope CLI and TUI. It never retrieves inference prompts or responses.
package consoleclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxCredentialBytes = 16 << 10
	maxResponseBytes   = 1 << 20
)

// HostStatus is the metadata-only shared status shape rendered by the native
// consoles. Fields not published by the HTTPS control API remain zero-valued
// rather than being estimated by the client.
type HostStatus struct {
	HostSlug                 string     `json:"host_slug"`
	DisplayName              string     `json:"display_name"`
	Platform                 string     `json:"platform"`
	EffectiveIntervalSeconds int        `json:"effective_interval_seconds"`
	FreshnessState           string     `json:"freshness_state"`
	LatestReceipt            *time.Time `json:"latest_receipt"`
	CurrentMetricCount       int64      `json:"current_metric_count"`
	UnavailableMetricCount   int64      `json:"unavailable_metric_count"`
	StaleMetricCount         int64      `json:"stale_metric_count"`
	ClockOffsetSeconds       *float64   `json:"clock_offset_seconds"`
	ClockOffsetQuality       *string    `json:"clock_offset_quality"`
}

type remoteFleetHost struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	Platform               string     `json:"platform"`
	Freshness              string     `json:"freshness"`
	LatestServerReceipt    *time.Time `json:"latest_server_receipt"`
	MetricCount            int64      `json:"metric_count"`
	UnavailableMetricCount int64      `json:"unavailable_metric_count"`
	StaleMetricCount       int64      `json:"stale_metric_count"`
}

func (host remoteFleetHost) status() HostStatus {
	return HostStatus{
		HostSlug:               host.ID,
		DisplayName:            host.Name,
		Platform:               host.Platform,
		FreshnessState:         host.Freshness,
		LatestReceipt:          host.LatestServerReceipt,
		CurrentMetricCount:     host.MetricCount,
		UnavailableMetricCount: host.UnavailableMetricCount,
		StaleMetricCount:       host.StaleMetricCount,
	}
}

// Config selects exactly one preferred read path. HTTPS takes precedence over
// the verifier database environment and SSH provides a local-machine relay
// without requiring a database credential on the invoking workstation.
type Config struct {
	DatabaseURL    string
	Endpoint       string
	CredentialFile string
	CAFile         string
	SSHTarget      string
	SSHCommand     string
	HTTPTimeout    time.Duration
}

func (config Config) normalized() Config {
	config.DatabaseURL = strings.TrimSpace(config.DatabaseURL)
	config.Endpoint = strings.TrimRight(strings.TrimSpace(config.Endpoint), "/")
	config.CredentialFile = strings.TrimSpace(config.CredentialFile)
	config.CAFile = strings.TrimSpace(config.CAFile)
	config.SSHTarget = strings.TrimSpace(config.SSHTarget)
	config.SSHCommand = strings.TrimSpace(config.SSHCommand)
	if config.SSHCommand == "" {
		config.SSHCommand = "nodescope-cli"
	}
	if config.HTTPTimeout <= 0 {
		config.HTTPTimeout = 10 * time.Second
	}
	return config
}

// Validate checks configuration before the caller opens a network connection.
func (config Config) Validate() error {
	config = config.normalized()
	if config.Endpoint != "" && config.SSHTarget != "" {
		return errors.New("endpoint and ssh-target cannot be used together")
	}
	if config.Endpoint == "" && config.SSHTarget == "" && config.DatabaseURL == "" {
		return errors.New("an HTTPS endpoint, SSH target, or verifier database URL is required")
	}
	if config.Endpoint != "" {
		endpoint, err := url.ParseRequestURI(config.Endpoint)
		if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
			return errors.New("endpoint must be a credential-free HTTPS base URL")
		}
		if config.CredentialFile == "" {
			return errors.New("credential-file is required for HTTPS endpoint mode")
		}
	}
	if config.SSHTarget != "" && strings.HasPrefix(config.SSHTarget, "-") {
		return errors.New("ssh-target must not begin with a dash")
	}
	return nil
}

// LoadFleet reads fleet metadata from a role-checked HTTPS API, an SSH relay
// running the read-only native CLI, or the legacy local verifier database.
func LoadFleet(ctx context.Context, config Config) ([]HostStatus, error) {
	config = config.normalized()
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if config.Endpoint != "" {
		return loadHTTPS(ctx, config)
	}
	if config.SSHTarget != "" {
		return loadSSH(ctx, config)
	}
	return loadDatabase(ctx, config.DatabaseURL)
}

func loadDatabase(ctx context.Context, databaseURL string) ([]HostStatus, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open verifier database connection: %w", err)
	}
	defer pool.Close()
	rows, err := pool.Query(ctx, `select host_slug, display_name, platform, effective_interval_seconds, freshness_state,
		latest_receipt, current_metric_count, unavailable_metric_count, stale_metric_count, clock_offset_seconds, clock_offset_quality
		from nodescope.fleet_ingestion_status()`)
	if err != nil {
		return nil, fmt.Errorf("query fleet status: %w", err)
	}
	defer rows.Close()
	result := make([]HostStatus, 0)
	for rows.Next() {
		var item HostStatus
		if err := rows.Scan(&item.HostSlug, &item.DisplayName, &item.Platform, &item.EffectiveIntervalSeconds, &item.FreshnessState,
			&item.LatestReceipt, &item.CurrentMetricCount, &item.UnavailableMetricCount, &item.StaleMetricCount, &item.ClockOffsetSeconds, &item.ClockOffsetQuality); err != nil {
			return nil, fmt.Errorf("scan fleet status: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read fleet status: %w", err)
	}
	return result, nil
}

func loadHTTPS(ctx context.Context, config Config) ([]HostStatus, error) {
	credential, err := readCredential(config.CredentialFile)
	if err != nil {
		return nil, err
	}
	client, err := httpsClient(config.CAFile, config.HTTPTimeout)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, config.Endpoint+"/api/v1/fleet", nil)
	if err != nil {
		return nil, fmt.Errorf("build control API request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+credential)
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch remote fleet status: %w", err)
	}
	defer response.Body.Close()
	body, err := boundedRead(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch remote fleet status: control API returned %s", response.Status)
	}
	var payload struct {
		Hosts []remoteFleetHost `json:"hosts"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode remote fleet status: %w", err)
	}
	hosts := make([]HostStatus, 0, len(payload.Hosts))
	for _, host := range payload.Hosts {
		hosts = append(hosts, host.status())
	}
	return hosts, nil
}

func httpsClient(caFile string, timeout time.Duration) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS13}
	if caFile != "" {
		certificate, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read control API CA file: %w", err)
		}
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(certificate) {
			return nil, errors.New("control API CA file contains no certificate")
		}
		transport.TLSClientConfig.RootCAs = roots
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("control API redirects are not permitted")
		},
	}, nil
}

func readCredential(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open control API credential file: %w", err)
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxCredentialBytes+1))
	if err != nil {
		return "", fmt.Errorf("read control API credential file: %w", err)
	}
	if len(content) > maxCredentialBytes {
		return "", errors.New("control API credential file exceeds the maximum size")
	}
	credential := strings.TrimSpace(string(content))
	if credential == "" {
		return "", errors.New("control API credential file is empty")
	}
	return credential, nil
}

func boundedRead(reader io.Reader) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(reader, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read remote fleet status: %w", err)
	}
	if len(content) > maxResponseBytes {
		return nil, errors.New("remote fleet status exceeds the maximum response size")
	}
	return content, nil
}

var executeSSH = func(ctx context.Context, target, command string) ([]byte, error) {
	process := exec.CommandContext(ctx, "ssh", target, command)
	stdout, err := process.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare remote NodeScope CLI: %w", err)
	}
	if err := process.Start(); err != nil {
		return nil, fmt.Errorf("start remote NodeScope CLI: %w", err)
	}
	output, readErr := io.ReadAll(io.LimitReader(stdout, maxResponseBytes+1))
	if readErr != nil {
		_ = process.Wait()
		return nil, fmt.Errorf("read remote NodeScope CLI: %w", readErr)
	}
	err = process.Wait()
	if err != nil {
		return nil, fmt.Errorf("execute remote NodeScope CLI: %w", err)
	}
	if len(output) > maxResponseBytes {
		return nil, errors.New("remote NodeScope CLI response exceeds the maximum size")
	}
	return output, nil
}

func loadSSH(ctx context.Context, config Config) ([]HostStatus, error) {
	output, err := executeSSH(ctx, config.SSHTarget, config.SSHCommand+" --format json")
	if err != nil {
		return nil, err
	}
	var hosts []HostStatus
	if err := json.Unmarshal(output, &hosts); err != nil {
		return nil, fmt.Errorf("decode SSH fleet status: %w", err)
	}
	return hosts, nil
}
