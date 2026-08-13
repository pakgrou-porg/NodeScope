package agent

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
)

func TestInferenceRuntimeEndpointCollectorRecordsAvailabilityWithoutReadingModelList(t *testing.T) {
	const modelCanary = "do-not-retain-model-name"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/v1/models" || request.Header.Get("Authorization") != "" {
			t.Fatalf("unsafe runtime health request: %s %s %#v", request.Method, request.URL.Path, request.Header)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"data":[{"id":"` + modelCanary + `"}]}`))
	}))
	defer server.Close()
	collector := newInferenceRuntimeEndpointCollector([]InferenceRuntimeEndpoint{{ID: "framework-vllm", Kind: "vllm", BaseURL: server.URL}}, server.Client())
	samples, err := collector.Collect(context.Background(), time.Unix(100, 0))
	if err != nil || len(samples) != 1 || samples[0].Metric.Quality != domain.QualityFresh || samples[0].Metric.Value == nil || *samples[0].Metric.Value != 1 || strings.Contains(samples[0].Metric.Semantics, modelCanary) {
		t.Fatalf("unexpected runtime endpoint output: samples=%#v err=%v", samples, err)
	}
}

func TestInferenceRuntimeEndpointCollectorReportsExplicitUnavailable(t *testing.T) {
	collector := NewInferenceRuntimeEndpointCollector(nil)
	samples, err := collector.Collect(context.Background(), time.Unix(100, 0))
	if err != nil || len(samples) != 1 || samples[0].Metric.Quality != domain.QualityUnavailable || samples[0].Metric.Value != nil {
		t.Fatalf("unexpected unavailable runtime output: samples=%#v err=%v", samples, err)
	}
}

func TestInferenceRuntimeEndpointCollectorClosesWithoutReadingResponseBody(t *testing.T) {
	body := &trackingReadCloser{Reader: bytes.NewBufferString(`{"data":[{"id":"model-metadata-canary"}]}`)}
	client := &http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Path != "/v1/models" || request.Body != nil {
			t.Fatalf("unexpected metadata-only request: %s %s body=%v", request.Method, request.URL.Path, request.Body)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header), Request: request}, nil
	})}
	collector := newInferenceRuntimeEndpointCollector([]InferenceRuntimeEndpoint{{ID: "local-vllm", Kind: "vllm", BaseURL: "http://127.0.0.1:8000"}}, client)
	if _, err := collector.Collect(context.Background(), time.Unix(100, 0)); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if body.reads != 0 || !body.closed {
		t.Fatalf("response body must be closed without reads: reads=%d closed=%t", body.reads, body.closed)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type trackingReadCloser struct {
	io.Reader
	reads  int
	closed bool
}

func (reader *trackingReadCloser) Read(payload []byte) (int, error) {
	reader.reads++
	return reader.Reader.Read(payload)
}

func (reader *trackingReadCloser) Close() error {
	reader.closed = true
	return nil
}
