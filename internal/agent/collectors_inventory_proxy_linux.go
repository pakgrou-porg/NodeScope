//go:build linux

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

const (
	maxInventoryProxyResponse = 1 << 20
	maxProxyContainers        = 1024
)

// InventoryProxyCollector fetches an explicitly bounded, fixed-schema container
// inventory from a separately deployed least-privilege HTTPS helper. It never
// connects to the Docker socket and is constructed only when the administrator
// explicitly enables the inventory feature.
type InventoryProxyCollector struct {
	client            *http.Client
	endpoint          string
	alertedContainers map[string]bool
}

func NewInventoryProxyCollector(config Config) (*InventoryProxyCollector, error) {
	client, err := newMTLSHTTPClient(config, 5*time.Second)
	if err != nil {
		return nil, err
	}
	selected := map[string]bool{}
	for _, value := range config.AlertedContainers {
		selected[value] = true
	}
	return &InventoryProxyCollector{
		client:            client,
		endpoint:          config.ContainerInventoryProxyURL,
		alertedContainers: selected,
	}, nil
}

func (collector *InventoryProxyCollector) Name() string { return "container_inventory_proxy" }

func (collector *InventoryProxyCollector) Collect(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	samples, _, err := collector.CollectContainerInventory(ctx, observedAt)
	return samples, err
}

func (collector *InventoryProxyCollector) CollectContainerInventory(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, []telemetry.ContainerInventory, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, collector.endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	response, err := noRedirectHTTPClient(collector.client).Do(request)
	if err != nil {
		return []telemetry.Sample{unavailableSample("containers", "container.inventory.available", "state", "inventory-proxy", "approved inventory proxy is unavailable", observedAt)}, nil, nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return []telemetry.Sample{unavailableSample("containers", "container.inventory.available", "state", "inventory-proxy", "approved inventory proxy did not return an inventory", observedAt)}, nil, nil
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxInventoryProxyResponse+1))
	decoder.DisallowUnknownFields()
	var payload inventoryProxyPayload
	if err := decoder.Decode(&payload); err != nil {
		return nil, nil, fmt.Errorf("decode fixed-schema container inventory: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, nil, fmt.Errorf("container inventory proxy returned trailing JSON")
	}
	if len(payload.Containers) > maxProxyContainers {
		return nil, nil, fmt.Errorf("container inventory proxy exceeded %d container limit", maxProxyContainers)
	}

	inventory := make([]telemetry.ContainerInventory, 0, len(payload.Containers))
	seen := make(map[string]struct{}, len(payload.Containers))
	for index, container := range payload.Containers {
		if err := container.validate(); err != nil {
			return nil, nil, fmt.Errorf("container inventory record %d: %w", index, err)
		}
		if _, exists := seen[container.ContainerID]; exists {
			return nil, nil, fmt.Errorf("duplicate container inventory %q", container.ContainerID)
		}
		seen[container.ContainerID] = struct{}{}
		selected := collector.alertedContainers[container.ContainerID] || collector.alertedContainers[container.Name]
		inventory = append(inventory, telemetry.ContainerInventory{
			ContainerID:         container.ContainerID,
			Name:                container.Name,
			Image:               container.Image,
			State:               container.State,
			Health:              container.Health,
			SelectedForAlerting: selected,
		})
	}

	samples := []telemetry.Sample{numericSample("containers", "container.inventory.available", "state", 1, "inventory-proxy", "approved fixed-schema inventory proxy is available", observedAt)}
	for _, container := range inventory {
		running := 0.0
		if container.State == "running" {
			running = 1
		}
		samples = append(samples, numericSample("container:"+container.ContainerID, "container.running", "state", running, "inventory-proxy", "proxy-reported container state for "+container.Name, observedAt))
		if container.Health == "" || container.Health == "unreported" {
			samples = append(samples, unavailableSample("container:"+container.ContainerID, "container.health", "state", "inventory-proxy", "proxy did not report a container health check", observedAt))
		}
	}
	return samples, inventory, nil
}

type inventoryProxyPayload struct {
	Containers []inventoryProxyContainer `json:"containers"`
}

type inventoryProxyContainer struct {
	ContainerID string `json:"containerId"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	State       string `json:"state"`
	Health      string `json:"health"`
}

func (container inventoryProxyContainer) validate() error {
	for label, field := range map[string]struct {
		value   string
		maximum int
	}{
		"container ID": {container.ContainerID, 128},
		"name":         {container.Name, 256},
		"image":        {container.Image, 512},
		"state":        {container.State, 64},
		"health":       {container.Health, 64},
	} {
		trimmed := strings.TrimSpace(field.value)
		if label != "health" && trimmed == "" {
			return fmt.Errorf("%s is required", label)
		}
		if !utf8.ValidString(field.value) || len(field.value) > field.maximum || strings.ContainsRune(field.value, '\x00') {
			return fmt.Errorf("%s is unsafe or exceeds %d bytes", label, field.maximum)
		}
	}
	return nil
}
