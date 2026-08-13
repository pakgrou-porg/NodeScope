export const metricQualities = ["fresh", "stale", "unavailable", "unsupported", "estimated", "experimental"] as const;

export type MetricQuality = (typeof metricQualities)[number];

const qualityLabels: Record<MetricQuality, string> = {
  fresh: "fresh evidence",
  stale: "stale evidence",
  unavailable: "unavailable evidence",
  unsupported: "not supported",
  estimated: "estimated evidence",
  experimental: "experimental evidence",
};

export function metricQualityLabel(quality: MetricQuality): string {
  return qualityLabels[quality];
}

export function metricQualityAriaLabel(quality: MetricQuality): string {
  return `Metric quality: ${metricQualityLabel(quality)}`;
}
