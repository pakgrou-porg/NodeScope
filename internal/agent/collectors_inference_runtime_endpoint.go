package agent

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

// InferenceRuntimeEndpointCollector checks only the response status of the
// documented OpenAI-compatible GET /v1/models discovery endpoint. It never
// reads model lists, response bytes, headers, credentials, prompt content, or
// completion content.
type InferenceRuntimeEndpointCollector struct {
	endpoints []InferenceRuntimeEndpoint
	client    *http.Client
}

func NewInferenceRuntimeEndpointCollector(endpoints []InferenceRuntimeEndpoint) *InferenceRuntimeEndpointCollector {
	return newInferenceRuntimeEndpointCollector(endpoints, &http.Client{Timeout: 3 * time.Second})
}

func newInferenceRuntimeEndpointCollector(endpoints []InferenceRuntimeEndpoint, client *http.Client) *InferenceRuntimeEndpointCollector {
	return &InferenceRuntimeEndpointCollector{endpoints: append([]InferenceRuntimeEndpoint(nil), endpoints...), client: client}
}

func (collector *InferenceRuntimeEndpointCollector) Name() string {
	return "inference_runtime_endpoints"
}

func (collector *InferenceRuntimeEndpointCollector) Collect(ctx context.Context, observedAt time.Time) ([]telemetry.Sample, error) {
	if len(collector.endpoints) == 0 {
		return []telemetry.Sample{runtimeEndpointUnavailability("runtime-endpoint-discovery", "no inference runtime endpoints are configured", observedAt)}, nil
	}
	endpoints := append([]InferenceRuntimeEndpoint(nil), collector.endpoints...)
	sort.Slice(endpoints, func(left, right int) bool { return endpoints[left].ID < endpoints[right].ID })
	samples := make([]telemetry.Sample, 0, len(endpoints))
	for _, endpoint := range endpoints {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.BaseURL+"/v1/models", nil)
		if err != nil {
			samples = append(samples, runtimeEndpointUnavailable(endpoint, "configured endpoint URL could not be requested", observedAt))
			continue
		}
		request.Header.Set("Accept", "application/json")
		response, err := collector.client.Do(request)
		if err != nil {
			samples = append(samples, runtimeEndpointUnavailable(endpoint, "configured endpoint did not respond", observedAt))
			continue
		}
		response.Body.Close() // Deliberately discard the model-list body without reading it.
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			samples = append(samples, runtimeEndpointUnavailable(endpoint, "configured endpoint returned a non-success status", observedAt))
			continue
		}
		value := 1.0
		samples = append(samples, telemetry.Sample{DeviceID: "runtime-api:" + endpoint.ID, Metric: domain.MetricValue{Name: "inference.runtime.api_available", Unit: "state", Value: &value, Quality: domain.QualityFresh, Source: "openai-compatible-v1-models", Semantics: "configured " + endpoint.Kind + " endpoint responded to metadata-only availability check", ObservedAt: observedAt}})
	}
	return samples, nil
}

func runtimeEndpointUnavailable(endpoint InferenceRuntimeEndpoint, semantics string, observedAt time.Time) telemetry.Sample {
	return runtimeEndpointUnavailability("runtime-api:"+endpoint.ID, semantics, observedAt)
}

func runtimeEndpointUnavailability(deviceID, semantics string, observedAt time.Time) telemetry.Sample {
	return telemetry.Sample{DeviceID: deviceID, Metric: domain.MetricValue{Name: "inference.runtime.api_available", Unit: "state", Quality: domain.QualityUnavailable, Source: "openai-compatible-v1-models", Semantics: semantics, ObservedAt: observedAt}}
}
