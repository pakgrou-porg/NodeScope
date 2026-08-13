//go:build linux

package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func TestInventoryProxyCollectorAcceptsBoundedFixedSchema(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", request.Method)
		}
		_, _ = writer.Write([]byte(`{"containers":[{"containerId":"abc123","name":"vllm","image":"vllm:latest","state":"running","health":"healthy"}]}`))
	}))
	defer server.Close()

	collector := &InventoryProxyCollector{
		client:            server.Client(),
		endpoint:          server.URL,
		alertedContainers: map[string]bool{"vllm": true},
	}
	samples, inventory, err := collector.CollectContainerInventory(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("collect inventory: %v", err)
	}
	if len(inventory) != 1 || inventory[0].Name != "vllm" || !inventory[0].SelectedForAlerting {
		t.Fatalf("unexpected inventory: %#v", inventory)
	}
	if len(samples) < 2 || samples[0].Metric.Quality != domain.QualityFresh || samples[0].Metric.Source != "inventory-proxy" {
		t.Fatalf("unexpected proxy samples: %#v", samples)
	}
}

func TestInventoryProxyCollectorRejectsUnknownSchemaFields(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"containers":[],"leakedDockerField":"must-not-pass"}`))
	}))
	defer server.Close()

	collector := &InventoryProxyCollector{client: server.Client(), endpoint: server.URL}
	if _, _, err := collector.CollectContainerInventory(context.Background(), time.Now().UTC()); err == nil {
		t.Fatal("expected unknown fixed-schema field to be rejected")
	}
}

func TestInventoryProxyCollectorReturnsUnavailableWhenProxyFails(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	collector := &InventoryProxyCollector{client: server.Client(), endpoint: server.URL}
	samples, inventory, err := collector.CollectContainerInventory(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("expected unavailable evidence rather than collector error: %v", err)
	}
	if len(inventory) != 0 || len(samples) != 1 || samples[0].Metric.Quality != domain.QualityUnavailable {
		t.Fatalf("unexpected unavailable outcome: samples=%#v inventory=%#v", samples, inventory)
	}
}

func TestInventoryProxyCollectorDoesNotFollowRedirects(t *testing.T) {
	redirectTargetCalled := false
	target := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirectTargetCalled = true
	}))
	defer target.Close()
	proxy := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusFound)
	}))
	defer proxy.Close()

	collector := &InventoryProxyCollector{client: proxy.Client(), endpoint: proxy.URL}
	samples, inventory, err := collector.CollectContainerInventory(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("expected redirect to produce unavailable evidence, got %v", err)
	}
	if redirectTargetCalled {
		t.Fatal("inventory proxy collector followed an unapproved redirect target")
	}
	if len(inventory) != 0 || len(samples) != 1 || samples[0].Metric.Quality != domain.QualityUnavailable {
		t.Fatalf("unexpected redirect outcome: samples=%#v inventory=%#v", samples, inventory)
	}
}
