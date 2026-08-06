//go:build linux

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type DockerCollector struct {
	client            *http.Client
	alertedContainers map[string]bool
}

func NewDockerCollector(alertedContainers []string) *DockerCollector {
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", "/var/run/docker.sock")
	}}
	selected := map[string]bool{}
	for _, value := range alertedContainers {
		selected[value] = true
	}
	return &DockerCollector{client: &http.Client{Transport: transport, Timeout: 5 * time.Second}, alertedContainers: selected}
}

func (collector *DockerCollector) Name() string { return "docker_inventory" }

func (collector *DockerCollector) Collect(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	samples, _, err := collector.CollectContainerInventory(ctx, observedAt)
	return samples, err
}

func (collector *DockerCollector) CollectContainerInventory(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, []telemetry.ContainerInventory, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/json?all=1", nil)
	if err != nil {
		return nil, nil, err
	}
	response, err := collector.client.Do(request)
	if err != nil {
		return []telemetry.Sample{unavailableSample("docker", "container.inventory.available", "state", "docker-api", "Docker socket is unavailable or unreadable", observedAt)}, nil, nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return []telemetry.Sample{unavailableSample("docker", "container.inventory.available", "state", "docker-api", "Docker API did not return an inventory", observedAt)}, nil, nil
	}
	var containers []dockerContainer
	if err := json.NewDecoder(http.MaxBytesReader(nil, response.Body, 2<<20)).Decode(&containers); err != nil {
		return nil, nil, fmt.Errorf("decode Docker container inventory: %w", err)
	}
	samples := []telemetry.Sample{numericSample("docker", "container.inventory.available", "state", 1, "docker-api", "Docker container inventory is available", observedAt)}
	inventory := make([]telemetry.ContainerInventory, 0, len(containers))
	for _, container := range containers {
		name := strings.TrimPrefix(firstString(container.Names, container.ID), "/")
		selected := collector.alertedContainers[container.ID] || collector.alertedContainers[name]
		health := "unreported"
		inventory = append(inventory, telemetry.ContainerInventory{
			ContainerID:         container.ID,
			Name:                name,
			Image:               container.Image,
			State:               container.State,
			Health:              health,
			SelectedForAlerting: selected,
		})
		running := 0.0
		if container.State == "running" {
			running = 1
		}
		samples = append(samples, numericSample("container:"+container.ID, "container.running", "state", running, "docker-api", "Docker container state for "+name, observedAt))
		samples = append(samples, unavailableSample("container:"+container.ID, "container.health", "state", "docker-api", "Docker health check status was not requested by this read-only inventory query", observedAt))
	}
	return samples, inventory, nil
}

type dockerContainer struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
	Image string   `json:"Image"`
	State string   `json:"State"`
}

func firstString(values []string, fallback string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return fallback
}
