package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxInferenceRequestBytes int64 = 8 << 20
const maxNonStreamingResponseBytes int64 = 32 << 20

type ClientAuthenticator interface {
	AuthenticateClient(context.Context, string) (string, error)
}

type StaticClientAuthenticator struct {
	Tokens map[string]string
}

func (authenticator StaticClientAuthenticator) AuthenticateClient(_ context.Context, token string) (string, error) {
	clientID, ok := authenticator.Tokens[token]
	if !ok || strings.TrimSpace(clientID) == "" {
		return "", fmt.Errorf("client credential is invalid")
	}
	return clientID, nil
}

type Handler struct {
	Registry      RouteRegistry
	Authenticator ClientAuthenticator
	Recorder      UsageRecorder
	Observer      OperationalObserver
	HTTPClient    *http.Client
	Now           func() time.Time
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	safeWriter := &proxyResponseWriter{ResponseWriter: writer}
	defer recoverProxyPanic(safeWriter)
	writer = safeWriter
	if request.Method != http.MethodPost {
		writeProxyProblem(writer, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}
	if handler.Registry == nil || handler.Authenticator == nil || handler.Recorder == nil {
		writeProxyProblem(writer, http.StatusServiceUnavailable, "inference proxy is not configured")
		return
	}
	clientToken, ok := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(clientToken) == "" {
		writeProxyProblem(writer, http.StatusUnauthorized, "client credential is required")
		return
	}
	clientID, err := handler.Authenticator.AuthenticateClient(request.Context(), strings.TrimSpace(clientToken))
	if err != nil {
		writeProxyProblem(writer, http.StatusUnauthorized, "client credential is invalid")
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxInferenceRequestBytes)
	defer request.Body.Close()
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		writeProxyProblem(writer, http.StatusRequestEntityTooLarge, "inference request exceeds the configured size limit")
		return
	}
	var metadata struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(payload, &metadata); err != nil || strings.TrimSpace(metadata.Model) == "" {
		writeProxyProblem(writer, http.StatusBadRequest, "OpenAI request must include a model")
		return
	}
	route, err := handler.Registry.Resolve(request.Context(), metadata.Model)
	if err != nil {
		writeProxyProblem(writer, http.StatusNotFound, "approved backend route is unavailable")
		return
	}
	started := handler.clock()()
	response, backendID, err := handler.forward(request.Context(), request, payload, route)
	if err != nil {
		handler.record(request.Context(), UsageEvent{OccurredAt: handler.clock()(), RouteID: route.ID, Model: metadata.Model, ClientID: clientID, BackendID: backendID, Streaming: metadata.Stream, DurationMilliseconds: handler.clock()().Sub(started).Milliseconds(), Outcome: "transport_error"})
		writeProxyProblem(writer, http.StatusBadGateway, "approved backend route did not respond")
		return
	}
	defer response.Body.Close()
	event := UsageEvent{OccurredAt: handler.clock()(), RouteID: route.ID, Model: metadata.Model, ClientID: clientID, BackendID: backendID, StatusCode: response.StatusCode, Streaming: metadata.Stream, Outcome: map[bool]string{true: "completed", false: "backend_error"}[response.StatusCode < 400]}
	if response.StatusCode >= http.StatusBadRequest {
		event.DurationMilliseconds = handler.clock()().Sub(started).Milliseconds()
		handler.record(request.Context(), event)
		writeProxyProblem(writer, http.StatusBadGateway, "approved backend route returned an error")
		return
	}

	copySafeBackendHeaders(writer.Header(), response.Header)
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-NodeScope-Route", route.ID)
	writer.WriteHeader(response.StatusCode)

	if metadata.Stream {
		var streamErr error
		event.TTFTMilliseconds, streamErr = copyStreaming(writer, response.Body, started, handler.clock())
		if streamErr != nil {
			event.Outcome = "stream_error"
		}
		event.DurationMilliseconds = handler.clock()().Sub(started).Milliseconds()
		handler.record(request.Context(), event)
		return
	}
	responseBytes, err := io.ReadAll(io.LimitReader(response.Body, maxNonStreamingResponseBytes+1))
	if err == nil && int64(len(responseBytes)) <= maxNonStreamingResponseBytes {
		_, _ = writer.Write(responseBytes)
		prompt, output, total := extractUsage(responseBytes)
		event.PromptTokens, event.OutputTokens, event.TotalTokens = prompt, output, total
		ttft := handler.clock()().Sub(started).Milliseconds()
		event.TTFTMilliseconds = &ttft
	} else {
		// The initial bounded bytes are relayed before the remaining stream. They
		// are not retained in usage metadata and no response text is logged.
		_, _ = writer.Write(responseBytes)
		_, _ = io.Copy(writer, response.Body)
	}
	event.DurationMilliseconds = handler.clock()().Sub(started).Milliseconds()
	handler.record(request.Context(), event)
}

func (handler *Handler) forward(ctx context.Context, original *http.Request, payload []byte, route BackendRoute) (*http.Response, string, error) {
	primary, err := forwardOnce(ctx, original, payload, route.PrimaryURL, handler.client())
	if err == nil {
		if shouldFailOverStatus(primary.StatusCode) && strings.TrimSpace(route.SecondaryURL) != "" {
			primary.Body.Close()
			secondary, secondaryErr := forwardOnce(ctx, original, payload, route.SecondaryURL, handler.client())
			if secondaryErr == nil {
				return secondary, route.ID + ":secondary", nil
			}
			// Preserve the primary status metadata if the fallback transport is
			// unavailable. Its body remains closed and is never relayed or logged.
			return primary, route.ID + ":primary", nil
		}
		return primary, route.ID + ":primary", nil
	}
	if strings.TrimSpace(route.SecondaryURL) == "" {
		return nil, route.ID + ":primary", err
	}
	secondary, secondaryErr := forwardOnce(ctx, original, payload, route.SecondaryURL, handler.client())
	if secondaryErr != nil {
		return nil, route.ID + ":secondary", secondaryErr
	}
	return secondary, route.ID + ":secondary", nil
}

func shouldFailOverStatus(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func forwardOnce(ctx context.Context, original *http.Request, payload []byte, base string, client *http.Client) (*http.Response, error) {
	baseURL, err := url.Parse(base)
	if err != nil || baseURL.Scheme != "http" && baseURL.Scheme != "https" || baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, fmt.Errorf("invalid approved backend URL")
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")
	if strings.HasSuffix(baseURL.Path, "/v1") {
		baseURL.Path += "/chat/completions"
	} else {
		baseURL.Path += "/v1/chat/completions"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"Accept", "Accept-Encoding", "User-Agent"} {
		for _, value := range original.Header.Values(key) {
			request.Header.Add(key, value)
		}
	}
	request.Header.Set("Content-Type", "application/json")
	return client.Do(request)
}

func copySafeBackendHeaders(destination, source http.Header) {
	for _, key := range []string{"Content-Type"} {
		for _, value := range source.Values(key) {
			destination.Add(key, value)
		}
	}
}

func copyStreaming(writer http.ResponseWriter, source io.Reader, started time.Time, now func() time.Time) (*int64, error) {
	buffer := make([]byte, 32*1024)
	first := true
	var ttft *int64
	for {
		count, err := source.Read(buffer)
		if count > 0 {
			if first {
				value := now().Sub(started).Milliseconds()
				ttft = &value
				first = false
			}
			_, _ = writer.Write(buffer[:count])
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				return ttft, nil
			}
			return ttft, err
		}
	}
}

func extractUsage(payload []byte) (*int64, *int64, *int64) {
	var response struct {
		Usage struct {
			PromptTokens     *int64 `json:"prompt_tokens"`
			CompletionTokens *int64 `json:"completion_tokens"`
			TotalTokens      *int64 `json:"total_tokens"`
		} `json:"usage"`
	}
	if json.Unmarshal(payload, &response) != nil {
		return nil, nil, nil
	}
	return response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens
}

func (handler *Handler) record(ctx context.Context, event UsageEvent) {
	_ = handler.Recorder.RecordUsage(ctx, event)
	if handler.Observer != nil {
		_ = handler.Observer.ObserveProxy(ctx, operationalEvent(event))
	}
}
func (handler *Handler) client() *http.Client {
	if handler.HTTPClient != nil {
		return handler.HTTPClient
	}
	return http.DefaultClient
}
func (handler *Handler) clock() func() time.Time {
	if handler.Now != nil {
		return handler.Now
	}
	return time.Now
}

func writeProxyProblem(writer http.ResponseWriter, status int, title string) {
	writer.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{"type": "https://nodescope.dev/problems/inference-proxy", "title": title, "status": status})
}

type proxyResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (writer *proxyResponseWriter) WriteHeader(status int) {
	if writer.wroteHeader {
		return
	}
	writer.wroteHeader = true
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *proxyResponseWriter) Write(payload []byte) (int, error) {
	if !writer.wroteHeader {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(payload)
}

func (writer *proxyResponseWriter) Flush() {
	if !writer.wroteHeader {
		writer.WriteHeader(http.StatusOK)
	}
	if flusher, ok := writer.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func recoverProxyPanic(writer *proxyResponseWriter) {
	if recover() != nil && !writer.wroteHeader {
		writeProxyProblem(writer, http.StatusInternalServerError, "inference proxy encountered an unexpected failure")
	}
}
