type LogLevel = "info" | "warn" | "error";
type SafeFields = Record<string, string | number | boolean | null | undefined>;

const sensitiveKey = /(authorization|cookie|credential|endpoint|key|password|secret|token|url)/i;

function safeFields(fields: SafeFields = {}): SafeFields {
  const result: SafeFields = {};
  for (const [key, value] of Object.entries(fields)) {
    if (sensitiveKey.test(key)) continue;
    result[key] = typeof value === "string" ? value.slice(0, 128) : value;
  }
  return result;
}

function write(level: LogLevel, event: string, fields?: SafeFields) {
  const entry = JSON.stringify({ level, event, ...safeFields(fields) });
  // This is the sole controlled console boundary for server diagnostics. It
  // emits structured, bounded metadata rather than arbitrary error objects.
  console[level](entry);
}

export const logger = {
  info: (event: string, fields?: SafeFields) => write("info", event, fields),
  warn: (event: string, fields?: SafeFields) => write("warn", event, fields),
  error: (event: string, fields?: SafeFields) => write("error", event, fields),
};
