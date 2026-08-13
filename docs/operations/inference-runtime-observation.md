# Inference Runtime Process Observation

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
