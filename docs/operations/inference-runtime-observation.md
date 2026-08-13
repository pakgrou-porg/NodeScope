# Inference Runtime Observation

NodeScope can report whether an administrator-approved local inference runtime process is running on Linux. This optional collector is designed for **availability evidence only**; it does not probe model APIs, read process command lines, inspect process environments, parse logs, retain prompts, or retain completions.

## Configure approved process names

Set `NODESCOPE_INFERENCE_RUNTIME_PROCESS_NAMES` in the root-owned agent environment file as a comma-separated list of exact Linux `/proc/<pid>/comm` names. Typical deployments may choose names such as `vllm`, `llama-server`, or `LM Studio`, but the administrator must confirm the exact kernel process name on each host. NodeScope does not infer a runtime from command-line arguments.

```ini
# Leave unset to keep runtime process observation disabled.
NODESCOPE_INFERENCE_RUNTIME_PROCESS_NAMES=vllm,llama-server,LM Studio
```

| Evidence state | Meaning |
|---|---|
| Fresh value `1` | At least one process with the configured exact `comm` name was observed. |
| Fresh value `0` | The procfs table was readable, but no process with that configured name was observed. |
| Unavailable | No runtime names were configured or procfs could not be read. This state is not replaced with zero. |

The collector emits `inference.runtime.running` with `procfs-comm` provenance and device IDs of the form `runtime:<configured-name>`. It is local discovery only; successful model endpoint health, loaded model identity, token rate, time to first token, and prompt processing remain separate runtime-specific validation and proxy telemetry concerns.

## Configure OpenAI-compatible endpoint availability

NodeScope can also make a **metadata-only** `GET /v1/models` availability request to a specifically configured local vLLM, llama.cpp, or LM Studio endpoint. All three runtimes expose OpenAI-compatible APIs; llama.cpp and LM Studio document `GET /v1/models` directly, while vLLM documents its OpenAI-compatible HTTP server.[1] [2] [3]

```ini
# Semicolon-separated: endpoint-id|runtime-kind|base-url
# Runtime kind: vllm, llama_cpp, or lm_studio
NODESCOPE_INFERENCE_RUNTIME_ENDPOINTS=framework-vllm|vllm|http://127.0.0.1:8000;msi-lmstudio|lm_studio|https://msi.example.lan:1234
```

The endpoint identifier is an opaque administrator-selected label. It may contain only letters, numbers, periods, underscores, and hyphens. NodeScope accepts clear-text HTTP only on `localhost`, `127.0.0.1`, or `::1`; all other endpoint checks require HTTPS. Endpoint URLs cannot include user information, queries, fragments, paths, or configured authorization credentials.

| Evidence state | Meaning |
|---|---|
| Fresh value `1` | The configured endpoint returned a successful status to `GET /v1/models`. |
| Unavailable | No endpoints were configured, the endpoint could not be reached, or it returned a non-success status. No zero estimate is substituted. |

The collector emits `inference.runtime.api_available` with `openai-compatible-v1-models` provenance and device IDs of the form `runtime-api:<endpoint-id>`. It sends no request body, reads no response body or model list, stores no HTTP headers, has no authentication-header configuration, and does **not** follow redirects. A redirect is therefore explicit unavailable evidence rather than permission to probe another location. Consequently, it is availability evidence only, not a model inventory, a benchmark, an endpoint authorization test, or an inference-content collection path.

## References

[1]: https://docs.vllm.ai/en/latest/serving/online_serving/openai_compatible_server/ "vLLM: OpenAI-Compatible Server"
[2]: https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md "llama.cpp HTTP server documentation"
[3]: https://lmstudio.ai/docs/developer/openai-compat "LM Studio OpenAI Compatibility Endpoints"
