//go:build linux

package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func InspectPreflight(now func() time.Time) PreflightReport {
	if now == nil {
		now = time.Now
	}
	return PreflightReport{
		GeneratedAt:     now().UTC(),
		OperatingSystem: readOSRelease(),
		Architecture:    runtime.GOARCH,
		Capabilities: []Capability{
			commandCapability("amd_smi", "amd-smi", "AMD GPU telemetry", []string{
				"Framework AMD GPU telemetry on Fedora is experimental until NodeScope publishes an exact qualified Fedora, kernel, firmware, ROCm, and AMD SMI matrix.",
				"Do not install an unverified package name from this report; follow the qualified NodeScope compatibility appendix or vendor-supported path for the actual host.",
				"Keep this capability unavailable rather than inferring GPU values when the required toolchain is not qualified.",
			}, "amd-smi version", "https://rocm.docs.amd.com/projects/amdsmi/en/latest/install/install.html"),
			commandCapability("xrt_smi", "xrt-smi", "AMD XDNA NPU telemetry", []string{
				"Fedora AMD XDNA telemetry is experimental until NodeScope publishes an exact qualified Fedora, kernel, firmware, ROCm, XRT, and XDNA matrix.",
				"Do not install or substitute a package based on this report; follow the qualified NodeScope compatibility appendix or vendor-supported path for the actual host.",
			}, "xrt-smi examine -f JSON -o /tmp/nodescope-xrt-smi.json", "https://ryzenai.docs.amd.com/en/latest/xrt_smi.html"),
			commandCapability("nvidia_smi", "nvidia-smi", "NVIDIA/DGX GPU telemetry", []string{
				"Install or repair the NVIDIA driver tooling supplied by DGX OS for this host.",
				"NodeScope will not infer dedicated VRAM when the platform exposes unified-memory semantics.",
			}, "nvidia-smi --query-gpu=name,temperature.gpu --format=csv,noheader", "https://docs.nvidia.com/dgx/dgx-spark/dgx-dashboard.html"),
			fileCapability("procfs", "/proc/stat", "Linux CPU and memory telemetry", nil, "cat /proc/stat"),
			{
				ID:     "container_inventory_proxy",
				State:  CapabilityUnavailable,
				Detail: "Container inventory remains unavailable until an administrator configures the approved fixed-schema HTTPS proxy.",
				RemediationSteps: []string{
					"Do not add the NodeScope service account to the docker group or mount the root-equivalent Docker socket.",
					"Deploy an approved narrow read-only proxy or privileged helper with a fixed output schema.",
					"Set NODESCOPE_DOCKER_INVENTORY_ENABLED=true and NODESCOPE_CONTAINER_INVENTORY_PROXY_URL to the proxy HTTPS endpoint.",
				},
				Verification: "NodeScope agent preflight plus an authenticated proxy HTTPS request",
			},
		},
	}
}

func commandCapability(id, command, detail string, remediation []string, verification, documentationURL string) Capability {
	path, err := exec.LookPath(command)
	if err == nil {
		return Capability{ID: id, State: CapabilityAvailable, Detail: detail, DetectedPath: path, Verification: verification, DocumentationURL: documentationURL}
	}
	return Capability{ID: id, State: CapabilityUnavailable, Detail: detail + " command is not on PATH.", RemediationSteps: remediation, Verification: verification, DocumentationURL: documentationURL}
}

func fileCapability(id, path, detail string, remediation []string, verification string) Capability {
	if _, err := os.Stat(path); err == nil {
		return Capability{ID: id, State: CapabilityAvailable, Detail: detail, DetectedPath: path, Verification: verification}
	}
	return Capability{ID: id, State: CapabilityUnavailable, Detail: detail + " path is unavailable.", RemediationSteps: remediation, Verification: verification}
}

func readOSRelease() string {
	contents, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return runtime.GOOS
	}
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}
	return runtime.GOOS
}

func FindMountedPath(path string) (string, bool) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", false
	}
	return resolved, true
}
