package agent

import "time"

type CapabilityState string

const (
	CapabilityAvailable   CapabilityState = "available"
	CapabilityUnavailable CapabilityState = "unavailable"
	CapabilityUnsupported CapabilityState = "unsupported"
)

type Capability struct {
	ID               string          `json:"id"`
	State            CapabilityState `json:"state"`
	Detail           string          `json:"detail"`
	DetectedPath     string          `json:"detectedPath,omitempty"`
	RemediationSteps []string        `json:"remediationSteps,omitempty"`
	Verification     string          `json:"verification,omitempty"`
	DocumentationURL string          `json:"documentationUrl,omitempty"`
}

type PreflightReport struct {
	GeneratedAt     time.Time    `json:"generatedAt"`
	OperatingSystem string       `json:"operatingSystem"`
	Architecture    string       `json:"architecture"`
	Capabilities    []Capability `json:"capabilities"`
}
