package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Kind string

const (
	Replica Kind = "replica"
	Agent   Kind = "agent"
)

type IssueRequest struct {
	Kind        Kind
	CommonName  string
	DNSNames    []string
	IPAddresses []net.IP
	ValidFor    time.Duration
}

func InitializeRoot(commonName string, validFor time.Duration) ([]byte, []byte, error) {
	if strings.TrimSpace(commonName) == "" {
		return nil, nil, fmt.Errorf("root common name is required")
	}
	if validFor < 24*time.Hour {
		return nil, nil, fmt.Errorf("root validity must be at least one day")
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := serialNumber()
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	template := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: commonName}, NotBefore: now.Add(-5 * time.Minute), NotAfter: now.Add(validFor), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: mustMarshalKey(key)}), nil
}

func Issue(caCertificatePEM, caKeyPEM []byte, request IssueRequest) ([]byte, []byte, error) {
	if request.Kind != Replica && request.Kind != Agent {
		return nil, nil, fmt.Errorf("certificate kind must be replica or agent")
	}
	if strings.TrimSpace(request.CommonName) == "" {
		return nil, nil, fmt.Errorf("leaf common name is required")
	}
	if request.ValidFor < time.Hour || request.ValidFor > 397*24*time.Hour {
		return nil, nil, fmt.Errorf("leaf validity must be one hour through 397 days")
	}
	ca, caKey, err := parseCA(caCertificatePEM, caKeyPEM)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	if !ca.NotAfter.After(now) {
		return nil, nil, fmt.Errorf("CA certificate is expired")
	}
	if ca.KeyUsage&x509.KeyUsageCertSign == 0 {
		return nil, nil, fmt.Errorf("CA certificate lacks certificate-signing usage")
	}
	if now.Add(request.ValidFor).After(ca.NotAfter) {
		return nil, nil, fmt.Errorf("leaf validity exceeds CA certificate validity")
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := serialNumber()
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: request.CommonName}, NotBefore: now.Add(-5 * time.Minute), NotAfter: now.Add(request.ValidFor), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment, DNSNames: request.DNSNames, IPAddresses: request.IPAddresses, BasicConstraintsValid: true}
	if request.Kind == Replica {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	} else {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: mustMarshalKey(key)}), nil
}

func WritePrivate(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}
func WritePublic(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return os.Chmod(path, 0o644)
}
func EnsureDirectory(path string) error { return os.MkdirAll(filepath.Clean(path), 0o700) }
func serialNumber() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, limit)
}
func mustMarshalKey(key *ecdsa.PrivateKey) []byte {
	data, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic(err)
	}
	return data
}
func parseCA(certificatePEM, keyPEM []byte) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	certificateBlock, _ := pem.Decode(certificatePEM)
	if certificateBlock == nil {
		return nil, nil, fmt.Errorf("CA certificate is not PEM")
	}
	certificate, err := x509.ParseCertificate(certificateBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	if !certificate.IsCA {
		return nil, nil, fmt.Errorf("certificate is not a CA")
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("CA key is not PEM")
	}
	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return certificate, key, nil
}
