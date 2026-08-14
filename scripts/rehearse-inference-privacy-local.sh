#!/usr/bin/env bash
# Rehearse metadata-only inference proxy behavior with deterministic local
# backends. This does not validate any actual vLLM, llama.cpp, or LM Studio host.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

if ! command -v go >/dev/null 2>&1; then
  echo "required dependency is unavailable: go" >&2
  exit 2
fi

go test ./internal/proxy -run '^(TestProxyForwardsOpenAIRequestAndRecordsOnlyUsageMetadata|TestProxyStreamingRecordsNoPromptOrResponseContent|TestProxyMalformedStreamDoesNotPersistFrameContent|TestProxyBackendErrorDoesNotRecordResponseContent|TestProxyFallsBackWhenPrimaryReturnsRetryableGatewayStatus|TestProxyDoesNotFollowBackendRedirects|TestProxyBackendHeadersAreAllowlisted|TestProxyTransportErrorDoesNotReflectUnderlyingErrorContent)$' -count=1

cat <<'JSON'
{"schema_version":1,"scope":"local deterministic streaming proxy privacy rehearsal","result":"passed","controls":{"prompt_and_completion_relay":"locally validated","metadata_only_usage":"locally validated","streaming_content_nonretention":"locally validated","malformed_stream_nonretention":"locally validated","retryable_fallback_containment":"locally validated","redirect_containment":"locally validated","header_allowlist":"locally validated","real_vllm_stream":"live environment gate","real_llama_cpp_stream":"live environment gate","real_lm_studio_stream":"live environment gate"},"recovery":"No approved runtime is contacted. If a live validation exposes content or backend credentials, disable the route, revoke the client credential, remove runtime approval, rotate affected backend credentials, and preserve only redacted evidence."}
JSON
