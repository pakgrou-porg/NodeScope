package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type Sender struct {
	client     *http.Client
	endpoints  []string
	credential string
}

func NewSender(config Config) (*Sender, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS13}
	if config.CACertificatePath != "" {
		certificate, err := os.ReadFile(config.CACertificatePath)
		if err != nil {
			return nil, fmt.Errorf("read NodeScope CA certificate: %w", err)
		}
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(certificate) {
			return nil, fmt.Errorf("NodeScope CA certificate contains no PEM certificate")
		}
		transport.TLSClientConfig.RootCAs = roots
	}
	if config.ClientCertificatePath != "" {
		certificate, err := tls.LoadX509KeyPair(config.ClientCertificatePath, config.ClientPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load NodeScope agent client certificate: %w", err)
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{certificate}
	}
	return &Sender{
		client:     &http.Client{Transport: transport, Timeout: 15 * time.Second},
		endpoints:  []string{strings.TrimRight(config.PreferredEndpoint, "/"), strings.TrimRight(config.SecondaryEndpoint, "/")},
		credential: config.Credential,
	}, nil
}

// Send always tries the preferred endpoint first. A request body is rebuilt for
// each endpoint to preserve deterministic idempotent payloads. Client and
// authorization failures fail closed; only transient failures can be retried.
func (sender *Sender) Send(ctx context.Context, envelope telemetry.Envelope) error {
	payload, err := telemetry.EncodeCompressedEnvelope(envelope)
	if err != nil {
		return &DeliveryError{Reason: "encode envelope", CanRetry: false}
	}
	var failures []string
	for _, endpoint := range sender.endpoints {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/api/v1/ingest", bytes.NewReader(payload))
		if err != nil {
			return &DeliveryError{Reason: "create ingestion request", CanRetry: false}
		}
		request.Header.Set("Content-Type", "application/x-protobuf")
		request.Header.Set("Content-Encoding", "zstd")
		request.Header.Set("Authorization", "Bearer "+sender.credential)
		response, requestErr := sender.client.Do(request)
		if requestErr != nil {
			failures = append(failures, endpoint+": transport failure")
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return nil
		}
		if response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != http.StatusTooManyRequests {
			return &DeliveryError{Reason: "ingestion rejected", CanRetry: false, StatusCode: response.StatusCode}
		}
		failures = append(failures, fmt.Sprintf("%s: status %d", endpoint, response.StatusCode))
	}
	return &DeliveryError{Reason: "all endpoints unavailable: " + strings.Join(failures, "; "), CanRetry: true}
}
