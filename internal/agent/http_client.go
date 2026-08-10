package agent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

func newMTLSHTTPClient(config Config, timeout time.Duration) (*http.Client, error) {
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
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}
