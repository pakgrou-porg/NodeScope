package pki

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"
)

func TestReplicaAndAgentCertificatesHaveSeparatedTLSUsage(t *testing.T) {
	rootCert, rootKey, err := InitializeRoot("NodeScope Offline Root", 3650*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	replicaCert, _, err := Issue(rootCert, rootKey, IssueRequest{Kind: Replica, CommonName: "framework", DNSNames: []string{"framework.nodescope.lan"}, IPAddresses: []net.IP{net.ParseIP("10.116.2.145")}, ValidFor: 90 * 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	agentCert, _, err := Issue(rootCert, rootKey, IssueRequest{Kind: Agent, CommonName: "agent-framework", ValidFor: 90 * 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	replica := parseCertificate(t, replicaCert)
	agent := parseCertificate(t, agentCert)
	if len(replica.ExtKeyUsage) != 1 || replica.ExtKeyUsage[0] != x509.ExtKeyUsageServerAuth || !replica.IPAddresses[0].Equal(net.ParseIP("10.116.2.145")) {
		t.Fatalf("invalid replica certificate %#v", replica)
	}
	if len(agent.ExtKeyUsage) != 1 || agent.ExtKeyUsage[0] != x509.ExtKeyUsageClientAuth {
		t.Fatalf("invalid agent certificate %#v", agent)
	}
}
func parseCertificate(t *testing.T, data []byte) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("certificate is not PEM")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return certificate
}
