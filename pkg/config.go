package pkg

import "os"

// GetEnv returns the value of the environment variable if set; otherwise returns the fallback value.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Global configuration variables.
var (
	// ReportDir is the directory where attestation report files are stored.
	ReportDir = GetEnv("SECRETAI_REPORT_DIR", "reports")
	// Names of the attestation report files.
	GPUAttestationFile  = "gpu_attestation.txt"
	CPUAttestationFile  = "tdx_attestation.txt"
	SelfAttestationFile = "self_report.txt"
)
