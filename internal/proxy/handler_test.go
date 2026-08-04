package proxy

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProxyForwardsOpenAIRequestAndRecordsOnlyUsageMetadata(t *testing.T) {
	const secretPrompt = "do-not-persist-this-prompt"
	const secretResponse = "do-not-persist-this-response"
	backend := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected backend path %s", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"` + secretResponse + `"}}],"usage":{"prompt_tokens":7,"completion_tokens":3,"total_tokens":10}}`))
	}))
	defer backend.Close()
	recorder := &MemoryRecorder{}
	handler := &Handler{
		Registry:      NewMemoryRegistry([]BackendRoute{{ID: "route-a", Model: "local-model", PrimaryURL: backend.URL, Enabled: true}}),
		Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"client-secret": "agentzero"}},
		Recorder:      recorder,
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"local-model","messages":[{"role":"user","content":"`+secretPrompt+`"}]}`))
	request.Header.Set("Authorization", "Bearer client-secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), secretResponse) {
		t.Fatal("proxy did not relay backend response")
	}
	events := recorder.Events()
	if len(events) != 1 {
		t.Fatalf("expected one usage event, got %d", len(events))
	}
	event := events[0]
	if event.ClientID != "agentzero" || event.RouteID != "route-a" || event.PromptTokens == nil || *event.PromptTokens != 7 || event.OutputTokens == nil || *event.OutputTokens != 3 {
		t.Fatalf("unexpected metadata-only event %#v", event)
	}
	if strings.Contains(event.Model, secretPrompt) || strings.Contains(event.BackendURL, secretResponse) {
		t.Fatal("usage event must not retain request or response content")
	}
}

func TestProxyFallsBackToApprovedSecondaryBackend(t *testing.T) {
	secondary := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"usage":{"total_tokens":1}}`))
	}))
	defer secondary.Close()
	recorder := &MemoryRecorder{}
	handler := &Handler{Registry: NewMemoryRegistry([]BackendRoute{{ID: "route-b", Model: "fallback-model", PrimaryURL: "http://127.0.0.1:1", SecondaryURL: secondary.URL, Enabled: true}}), Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"token": "cli"}}, Recorder: recorder, HTTPClient: &http.Client{Timeout: 100 * time.Millisecond}}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"fallback-model"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected fallback response, got %d", response.Code)
	}
	if events := recorder.Events(); len(events) != 1 || events[0].BackendURL != secondary.URL {
		t.Fatalf("expected secondary metadata event, got %#v", events)
	}
}

func TestProxyRejectsInvalidClientCredentialWithoutForwarding(t *testing.T) {
	handler := &Handler{Registry: NewMemoryRegistry(nil), Authenticator: StaticClientAuthenticator{Tokens: map[string]string{}}, Recorder: &MemoryRecorder{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"x"}`))
	request.Header.Set("Authorization", "Bearer bad")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestProxyStreamingRecordsNoPromptOrResponseContent(t *testing.T) {
	const secretPrompt = "stream-secret-prompt"
	const secretResponse = "stream-secret-response"
	backend := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte("data: " + secretResponse + "\n\n"))
	}))
	defer backend.Close()
	recorder := &MemoryRecorder{}
	handler := &Handler{Registry: NewMemoryRegistry([]BackendRoute{{ID: "stream-route", Model: "stream-model", PrimaryURL: backend.URL, Enabled: true}}), Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"token": "agentzero"}}, Recorder: recorder}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"stream-model","stream":true,"messages":[{"content":"`+secretPrompt+`"}]}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), secretResponse) {
		t.Fatalf("stream was not relayed safely: %d %s", response.Code, response.Body.String())
	}
	events := recorder.Events()
	if len(events) != 1 || !events[0].Streaming || events[0].PromptTokens != nil || events[0].OutputTokens != nil {
		t.Fatalf("unexpected streaming event %#v", events)
	}
	serialized := fmt.Sprintf("%#v", events[0])
	if strings.Contains(serialized, secretPrompt) || strings.Contains(serialized, secretResponse) {
		t.Fatal("streaming usage event retained inference content")
	}
}

func TestProxyBackendErrorDoesNotRecordResponseContent(t *testing.T) {
	const secretPrompt = "error-secret-prompt"
	const secretResponse = "error-secret-response"
	backend := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte(secretResponse))
	}))
	defer backend.Close()
	recorder := &MemoryRecorder{}
	handler := &Handler{Registry: NewMemoryRegistry([]BackendRoute{{ID: "error-route", Model: "error-model", PrimaryURL: backend.URL, Enabled: true}}), Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"token": "agentzero"}}, Recorder: recorder}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"error-model","messages":[{"content":"`+secretPrompt+`"}]}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadGateway || !strings.Contains(response.Body.String(), secretResponse) {
		t.Fatalf("backend error was not relayed: %d %s", response.Code, response.Body.String())
	}
	events := recorder.Events()
	if len(events) != 1 || events[0].Outcome != "backend_error" {
		t.Fatalf("unexpected backend-error event %#v", events)
	}
	serialized := fmt.Sprintf("%#v", events[0])
	if strings.Contains(serialized, secretPrompt) || strings.Contains(serialized, secretResponse) {
		t.Fatal("backend-error usage event retained inference content")
	}
}
