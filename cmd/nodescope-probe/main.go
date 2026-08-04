package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: nodescope-probe <health-url>")
		os.Exit(2)
	}
	client := &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- local container health endpoint only.
		},
	}
	response, err := client.Get(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "NodeScope health probe failed:", err)
		os.Exit(1)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Fprintln(os.Stderr, "NodeScope health probe returned:", response.Status)
		os.Exit(1)
	}
}
