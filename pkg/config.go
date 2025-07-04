package pkg

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// init loads the .env file so that environment variables are available during initialization.
func init() {
	// Load the .env file here so that the variables are available during initialization.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found in pkg")
	}

	// Now initialize configuration variables.
	ReportDir = GetEnv("SECRETVM_REPORT_DIR", "reports")
	RESTServerIP = GetEnv("SECRETVM_REST_SERVER_IP", "0.0.0.0")
	Secure = GetBool("SECRETVM_SECURE", true)
	Port = GetInt("SECRETVM_REST_SERVER_PORT", 29343)

	// Certificate paths
	CertPath = GetEnv("SECRETVM_CERT_PATH", "cert/ssl_cert.pem")
	KeyPath = GetEnv("SECRETVM_KEY_PATH", "cert/ssl_key.pem")

	// Attestation tool configuration
	AttestTool = GetEnv("SECRETVM_ATTEST_TOOL", "attest_tool")
	AttestTimeout = time.Duration(GetInt("SECRETVM_ATTEST_TIMEOUT_SEC", 10)) * time.Second

	// Set names of the attestation report files - can be configured via env vars
	GPUAttestationFile = GetEnv("SECRETVM_GPU_ATTESTATION_FILE", "gpu_attestation.txt")
	CPUAttestationFile = GetEnv("SECRETVM_CPU_ATTESTATION_FILE", "tdx_attestation.txt")
	SelfAttestationFile = GetEnv("SECRETVM_SELF_ATTESTATION_FILE", "self_report.txt")

	// Path to docker-compose file (must be set in env).
	DockerComposePath = GetEnv("SECRETVM_DOCKER_COMPOSE_PATH", "docker_compose.yaml")
	
	// Path to vm config file (must be set in env).
	VmConfigPath = GetEnv("SECRETVM_CONFIG_PATH", "/mnt/config/secret-vm.json")

	// Path to filesystem mount point
	FsMountPath = GetEnv("SECRETVM_FS_MOUNT_PATH", "/mnt/secure")

	// Create report directory if it doesn't exist
	if err := os.MkdirAll(ReportDir, 0755); err != nil {
		log.Printf("Warning: Failed to create report directory %s: %v", ReportDir, err)
	}
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
	// Server configuration
	ReportDir    string // Directory where attestation report files are stored
	RESTServerIP string // IP address on which the server should listen
	Secure       bool   // Whether HTTPS should be enabled
	Port         int    // Port number on which the server should listen
	CertPath     string // Path to SSL certificate file
	KeyPath      string // Path to SSL key file

	// Attestation tool configuration
	AttestTool    string        // Command name for the attestation tool
	AttestTimeout time.Duration // Timeout for attestation command execution

	// Names of the attestation report files
	GPUAttestationFile  string
	CPUAttestationFile  string
	SelfAttestationFile string

	// Path to docker-compose file
	DockerComposePath string
	VmConfigPath      string

	// Filesystem mount path
	FsMountPath string
)
