#!/usr/bin/env bash
# Keep the local streaming proof narrow and prevent it from being represented as
# a production backend validation.
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"

harness=scripts/rehearse-inference-privacy-local.sh
for required in \
  'TestProxyStreamingRecordsNoPromptOrResponseContent' \
  'TestProxyMalformedStreamDoesNotPersistFrameContent' \
  'TestProxyFallsBackWhenPrimaryReturnsRetryableGatewayStatus' \
  'TestProxyDoesNotFollowBackendRedirects' \
  'metadata_only_usage":"locally validated' \
  'real_vllm_stream":"live environment gate' \
  'real_llama_cpp_stream":"live environment gate' \
  'real_lm_studio_stream":"live environment gate' \
  'No approved runtime is contacted'; do
  if ! grep -Fq "$required" "$harness"; then
    echo "inference privacy rehearsal must retain $required" >&2
    exit 1
  fi
done

echo "Inference privacy rehearsal contract passed."
