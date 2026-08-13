package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
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

func TestIssueRejectsInvalidOrOutlivedCertificateAuthorities(t *testing.T) {
	validRoot, validRootKey, err := InitializeRoot("short-lived root", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	expiredRoot, expiredRootKey := testCA(t, time.Now().UTC().Add(-time.Hour), x509.KeyUsageCertSign)
	nonSigningRoot, nonSigningRootKey := testCA(t, time.Now().UTC().Add(24*time.Hour), x509.KeyUsageDigitalSignature)
	request := IssueRequest{Kind: Agent, CommonName: "agent-framework", ValidFor: 48 * time.Hour}
	for _, testCase := range []struct {
		name string
		cert []byte
		key  []byte
		want string
	}{
		{name: "outlives issuer", cert: validRoot, key: validRootKey, want: "exceeds CA"},
		{name: "expired issuer", cert: expiredRoot, key: expiredRootKey, want: "expired"},
		{name: "non signing issuer", cert: nonSigningRoot, key: nonSigningRootKey, want: "certificate-signing"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, issueErr := Issue(testCase.cert, testCase.key, request)
			if issueErr == nil || !strings.Contains(issueErr.Error(), testCase.want) {
				t.Fatalf("expected %q rejection, got %v", testCase.want, issueErr)
			}
		})
	}
}

func TestCertificatePublicationReplacesSymlinkWithoutFollowingIt(t *testing.T) {
	directory := t.TempDir()
	for _, testCase := range []struct {
		name  string
		write func(string, []byte) error
		mode  os.FileMode
	}{
		{name: "private", write: WritePrivate, mode: 0o600},
		{name: "public", write: WritePublic, mode: 0o644},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			external := filepath.Join(directory, testCase.name+"-external.pem")
			output := filepath.Join(directory, testCase.name+".pem")
			if err := os.WriteFile(external, []byte("outside-canary"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, output); err != nil {
				t.Fatal(err)
			}
			if err := testCase.write(output, []byte("published-certificate")); err != nil {
				t.Fatal(err)
			}
			outside, err := os.ReadFile(external)
			if err != nil || string(outside) != "outside-canary" {
				t.Fatalf("publication followed existing symlink: outside=%q err=%v", outside, err)
			}
			info, err := os.Lstat(output)
			if err != nil || info.Mode()&os.ModeSymlink != 0 {
				t.Fatalf("output must replace symlink with regular file: info=%#v err=%v", info, err)
			}
			if info.Mode().Perm() != testCase.mode {
				t.Fatalf("output mode=%#o want=%#o", info.Mode().Perm(), testCase.mode)
			}
			published, err := os.ReadFile(output)
			if err != nil || string(published) != "published-certificate" {
				t.Fatalf("unexpected published output=%q err=%v", published, err)
			}
		})
	}
}

func testCA(t *testing.T, notAfter time.Time, keyUsage x509.KeyUsage) ([]byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	template := &x509.Certificate{SerialNumber: big.NewInt(41), Subject: pkix.Name{CommonName: "test CA"}, NotBefore: now.Add(-2 * time.Hour), NotAfter: notAfter, IsCA: true, BasicConstraintsValid: true, KeyUsage: keyUsage}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	encodedKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encodedKey})
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
