package app

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// BuildServerTLSConfig builds the server-side policy without logging certificate
// contents. Agent mTLS is opt-in until the internal CA is deployed, then the
// replica can fail closed by setting NODESCOPE_REQUIRE_AGENT_MTLS=true.
func BuildServerTLSConfig(config Config) (*tls.Config, error) {
	if config.CertificatePath == "" {
		return nil, nil
	}
	policy := &tls.Config{MinVersion: tls.VersionTLS13}
	if !config.RequireAgentMTLS {
		return policy, nil
	}
	certificate, err := os.ReadFile(config.AgentClientCAPath)
	if err != nil {
		return nil, fmt.Errorf("read NodeScope agent client CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certificate) {
		return nil, fmt.Errorf("NodeScope agent client CA contains no PEM certificate")
	}
	policy.ClientCAs = roots
	policy.ClientAuth = tls.RequireAndVerifyClientCert
	return policy, nil
}
