package pkg

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// init loads the .env file so that environment variables are available during initialization.
func init() {
	// Load the .env file here so that the variables are available during initialization.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found in pkg")
	}

	// Now initialize configuration variables.
	ReportDir = GetEnv("SECRETAI_REPORT_DIR", "reports")
	RESTServerIP = GetEnv("SECRETAI_REST_SERVER_IP", "0.0.0.0")
	Secure = GetBool("SECRETAI_SECURE", true)
	Port = GetInt("SECRETAI_REST_SERVER_PORT", 29343)

	// Set names of the attestation report files.
	GPUAttestationFile = "gpu_attestation.txt"
	CPUAttestationFile = "tdx_attestation.txt"
	SelfAttestationFile = "self_report.txt"
}

// GetEnv returns the value of the environment variable if set; otherwise returns the fallback value.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetBool returns the boolean value of an environment variable if set; otherwise returns the fallback value.
func GetBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

// GetInt returns the integer value of an environment variable if set; otherwise returns the fallback value.
func GetInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

// Global configuration variables.
var (
	// ReportDir is the directory where attestation report files are stored.
	ReportDir string
	// RESTServerIP is the IP address on which the server should listen.
	RESTServerIP string
	// Secure indicates whether HTTPS should be enabled.
	Secure bool
	// Port is the port number on which the server should listen.
	Port int

	// Names of the attestation report files.
	GPUAttestationFile  string
	CPUAttestationFile  string
	SelfAttestationFile string
)
