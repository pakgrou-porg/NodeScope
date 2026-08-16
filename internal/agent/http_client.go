package agent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"runtime"
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
		if err := requirePrivateTLSKey(config.ClientPrivateKeyPath); err != nil {
			return nil, err
		}
		certificate, err := tls.LoadX509KeyPair(config.ClientCertificatePath, config.ClientPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load NodeScope agent client certificate: %w", err)
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{certificate}
	}
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}

func requirePrivateTLSKey(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("lstat NodeScope agent client private key: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("NodeScope agent client private key must be a direct regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("NodeScope agent client private key must not be group- or world-accessible")
	}
	return nil
}

func noRedirectHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	safeClient := *client
	safeClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &safeClient
}
