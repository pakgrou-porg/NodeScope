package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/pki"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "init-root":
		initRoot(os.Args[2:])
	case "issue":
		issue(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}
func usage() { fmt.Fprintln(os.Stderr, "Usage: nodescope-pki init-root|issue [flags]") }
func initRoot(arguments []string) {
	flags := flag.NewFlagSet("init-root", flag.ExitOnError)
	var commonName, output string
	var years int
	flags.StringVar(&commonName, "common-name", "NodeScope Offline Root CA", "root certificate common name")
	flags.StringVar(&output, "output-directory", "", "offline root output directory")
	flags.IntVar(&years, "years", 10, "root validity in years")
	_ = flags.Parse(arguments)
	if output == "" || years < 1 || years > 20 {
		fmt.Fprintln(os.Stderr, "output-directory is required and years must be 1..20")
		os.Exit(2)
	}
	if err := pki.EnsureDirectory(output); err != nil {
		fail(err)
	}
	certificate, key, err := pki.InitializeRoot(commonName, time.Duration(years)*365*24*time.Hour)
	if err != nil {
		fail(err)
	}
	if err := pki.WritePublic(filepath.Join(output, "root-ca.pem"), certificate); err != nil {
		fail(err)
	}
	if err := pki.WritePrivate(filepath.Join(output, "root-ca-key.pem"), key); err != nil {
		fail(err)
	}
	fmt.Println("created offline root CA; move root-ca-key.pem to protected offline storage before issuing leaf certificates")
}
func issue(arguments []string) {
	flags := flag.NewFlagSet("issue", flag.ExitOnError)
	var kind, commonName, caCertificate, caKey, certificatePath, keyPath, dnsValues, ipValues string
	var days int
	flags.StringVar(&kind, "kind", "", "replica or agent")
	flags.StringVar(&commonName, "common-name", "", "leaf common name")
	flags.StringVar(&caCertificate, "ca-certificate", "", "offline root certificate PEM")
	flags.StringVar(&caKey, "ca-key", "", "offline root private key PEM")
	flags.StringVar(&certificatePath, "certificate-output", "", "leaf certificate output PEM")
	flags.StringVar(&keyPath, "key-output", "", "leaf private-key output PEM")
	flags.StringVar(&dnsValues, "dns-san", "", "comma-separated DNS SANs")
	flags.StringVar(&ipValues, "ip-san", "", "comma-separated IP SANs")
	flags.IntVar(&days, "days", 90, "leaf validity in days")
	_ = flags.Parse(arguments)
	if kind == "" || commonName == "" || caCertificate == "" || caKey == "" || certificatePath == "" || keyPath == "" || days < 1 || days > 397 {
		fmt.Fprintln(os.Stderr, "kind, common-name, CA inputs, outputs, and a 1..397-day validity are required")
		os.Exit(2)
	}
	certificateData, err := os.ReadFile(caCertificate)
	if err != nil {
		fail(err)
	}
	keyData, err := os.ReadFile(caKey)
	if err != nil {
		fail(err)
	}
	request := pki.IssueRequest{Kind: pki.Kind(kind), CommonName: commonName, DNSNames: split(dnsValues), IPAddresses: parseIPs(ipValues), ValidFor: time.Duration(days) * 24 * time.Hour}
	certificate, key, err := pki.Issue(certificateData, keyData, request)
	if err != nil {
		fail(err)
	}
	if err := pki.WritePublic(certificatePath, certificate); err != nil {
		fail(err)
	}
	if err := pki.WritePrivate(keyPath, key); err != nil {
		fail(err)
	}
	fmt.Println("issued", kind, "certificate", commonName)
}
func split(value string) []string {
	var result []string
	for _, part := range strings.Split(value, ",") {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}
func parseIPs(value string) []net.IP {
	var result []net.IP
	for _, part := range split(value) {
		if ip := net.ParseIP(part); ip != nil {
			result = append(result, ip)
		} else {
			fmt.Fprintln(os.Stderr, "invalid IP SAN:", part)
			os.Exit(2)
		}
	}
	return result
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
