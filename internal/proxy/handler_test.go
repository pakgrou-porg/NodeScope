package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
		if request.Header.Get("Authorization") != "" || request.Header.Get("X-NodeScope-Route") != "" {
			t.Fatalf("NodeScope credentials or internal headers reached backend: %#v", request.Header)
		}
		if request.Header.Get("Accept") != "application/json" {
			t.Fatalf("safe Accept header was not forwarded: %#v", request.Header)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"` + secretResponse + `"}}],"usage":{"prompt_tokens":7,"completion_tokens":3,"total_tokens":10}}`))
	}))
	defer backend.Close()
	recorder := &MemoryRecorder{}
	requestLog := &memorySink{}
	tracer := &memorySink{}
	auditor := &memorySink{}
	supportExport := &memorySink{}
	handler := &Handler{
		Registry:      NewMemoryRegistry([]BackendRoute{{ID: "route-a", Model: "local-model", PrimaryURL: backend.URL, Enabled: true}}),
		Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"client-secret": "agentzero"}},
		Recorder:      recorder,
		Observer:      MetadataOnlyFanout{RequestLog: requestLog, Tracer: tracer, Auditor: auditor, SupportExport: supportExport},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"local-model","messages":[{"role":"user","content":"`+secretPrompt+`"}]}`))
	request.Header.Set("Authorization", "Bearer client-secret")
	request.Header.Set("X-NodeScope-Route", "internal-canary")
	request.Header.Set("Accept", "application/json")
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
	if strings.Contains(event.Model, secretPrompt) || strings.Contains(event.BackendID, secretResponse) || event.BackendID != "route-a:primary" {
		t.Fatal("usage event must not retain request or response content")
	}
	for name, sink := range map[string]*memorySink{"request log": requestLog, "tracer": tracer, "auditor": auditor, "support export": supportExport} {
		if len(sink.events) != 1 || strings.Contains(fmt.Sprintf("%#v", sink.events[0]), secretPrompt) || strings.Contains(fmt.Sprintf("%#v", sink.events[0]), secretResponse) || sink.events[0].BackendID != "route-a:primary" {
			t.Fatalf("%s retained unsafe content: %#v", name, sink.events)
		}
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
	if events := recorder.Events(); len(events) != 1 || events[0].BackendID != "route-b:secondary" {
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
	if response.Code != http.StatusBadGateway || strings.Contains(response.Body.String(), secretResponse) || !strings.Contains(response.Body.String(), "approved backend route returned an error") {
		t.Fatalf("backend error was not normalized safely: %d %s", response.Code, response.Body.String())
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

func TestProxyBackendHeadersAreAllowlisted(t *testing.T) {
	const secretHeader = "backend-header-content-canary"
	backend := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Backend-Diagnostic", secretHeader)
		writer.Header().Set("Set-Cookie", secretHeader)
		_, _ = writer.Write([]byte(`{"usage":{"total_tokens":1}}`))
	}))
	defer backend.Close()
	recorder := &MemoryRecorder{}
	handler := &Handler{Registry: NewMemoryRegistry([]BackendRoute{{ID: "header-route", Model: "header-model", PrimaryURL: backend.URL, Enabled: true}}), Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"token": "agentzero"}}, Recorder: recorder}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"header-model"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("X-Backend-Diagnostic") != "" || response.Header().Get("Set-Cookie") != "" || strings.Contains(response.Header().Values("Content-Type")[0], secretHeader) {
		t.Fatalf("unsafe backend headers were relayed: %#v", response.Result().Header)
	}
	if len(recorder.Events()) != 1 || strings.Contains(fmt.Sprintf("%#v", recorder.Events()[0]), secretHeader) {
		t.Fatalf("event retained backend-header content: %#v", recorder.Events())
	}
}

func TestProxyTransportErrorDoesNotReflectUnderlyingErrorContent(t *testing.T) {
	const secretPrompt = "transport-error-prompt-canary"
	const secretError = "transport-error-detail-canary"
	recorder := &MemoryRecorder{}
	handler := &Handler{
		Registry:      NewMemoryRegistry([]BackendRoute{{ID: "transport-route", Model: "transport-model", PrimaryURL: "https://backend.test", Enabled: true}}),
		Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"token": "agentzero"}},
		Recorder:      recorder,
		HTTPClient:    &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New(secretError) })},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"transport-model","messages":[{"content":"`+secretPrompt+`"}]}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadGateway || strings.Contains(response.Body.String(), secretPrompt) || strings.Contains(response.Body.String(), secretError) {
		t.Fatalf("transport error exposed content: %d %s", response.Code, response.Body.String())
	}
	events := recorder.Events()
	if len(events) != 1 || events[0].Outcome != "transport_error" || strings.Contains(fmt.Sprintf("%#v", events[0]), secretPrompt) || strings.Contains(fmt.Sprintf("%#v", events[0]), secretError) {
		t.Fatalf("transport event retained content: %#v", events)
	}
}

func TestProxyMalformedStreamDoesNotPersistFrameContent(t *testing.T) {
	const secretPrompt = "malformed-stream-prompt-canary"
	const secretFrame = "malformed-stream-frame-canary"
	recorder := &MemoryRecorder{}
	handler := &Handler{
		Registry:      NewMemoryRegistry([]BackendRoute{{ID: "stream-error-route", Model: "stream-error-model", PrimaryURL: "https://backend.test", Enabled: true}}),
		Authenticator: StaticClientAuthenticator{Tokens: map[string]string{"token": "agentzero"}},
		Recorder:      recorder,
		HTTPClient: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: &malformedBody{payload: []byte("data: " + secretFrame + "\n\n")}}, nil
		})},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"stream-error-model","stream":true,"messages":[{"content":"`+secretPrompt+`"}]}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), secretFrame) {
		t.Fatalf("stream was not relayed: %d %s", response.Code, response.Body.String())
	}
	events := recorder.Events()
	if len(events) != 1 || events[0].Outcome != "stream_error" || strings.Contains(fmt.Sprintf("%#v", events[0]), secretPrompt) || strings.Contains(fmt.Sprintf("%#v", events[0]), secretFrame) {
		t.Fatalf("malformed stream event retained content: %#v", events)
	}
}

func TestProxyPanicDoesNotReflectPanicContent(t *testing.T) {
	const secretPanic = "panic-content-canary"
	handler := &Handler{Registry: NewMemoryRegistry(nil), Authenticator: panicAuthenticator{value: secretPanic}, Recorder: &MemoryRecorder{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"panic-model"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), secretPanic) || !strings.Contains(response.Body.String(), "unexpected failure") {
		t.Fatalf("panic response exposed diagnostic content: %d %s", response.Code, response.Body.String())
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type malformedBody struct {
	payload []byte
	read    bool
}

func (body *malformedBody) Read(buffer []byte) (int, error) {
	if body.read {
		return 0, io.EOF
	}
	body.read = true
	return copy(buffer, body.payload), errors.New("malformed stream")
}

func (*malformedBody) Close() error { return nil }

type panicAuthenticator struct{ value string }

func (authenticator panicAuthenticator) AuthenticateClient(context.Context, string) (string, error) {
	panic(authenticator.value)
}
