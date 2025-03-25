# SecretAI Attest REST Server (Go)

SecretAI Attest REST Server is a lightweight REST server implemented in Go. It provides attestation reports for confidential virtual machines (VMs) over HTTPS. The server exposes multiple endpoints to return different attestation reports, including:

- **/status** – Returns the server status.
- **/attestation** – Executes an internal process (e.g., `attest_tool`) and returns a JSON attestation report.
- **/gpu** – Returns the NVIDIA confidential GPU attestation report.
- **/cpu** – Returns the Intel TDX attestation report.
- **/self** – Returns self attestation data (e.g., TDX measurement registers).

## Features

- **Secure Communication:** Supports HTTPS with TLS certificates.
- **Configuration via Environment Variables:** Defaults are defined in a configuration package and can be overridden by a `.env` file.
- **Modular Structure:** Uses a command-line interface and a dedicated package (`pkg`) for configuration, handlers, and middleware.
- **Command-Line Flags:** Allows overriding defaults (secure mode, port, and IP address) using flags.
- **Enhanced Security Headers:** Implements best practice security headers for all responses.
- **CORS Support:** Built-in CORS middleware for cross-origin requests.
- **Graceful Shutdown:** Handles in-flight requests properly during server shutdown.
- **Improved Logging:** Detailed logging including request methods, status codes, and response times.
- **Context Support:** Uses Go contexts for timeout and cancellation management.
- **Method Validation:** All endpoints validate HTTP methods to ensure proper usage.
- **Standardized Error Responses:** Consistent JSON error responses across all endpoints.

## Project Structure

```
secretai-attest-rest/
├── .env                   # Environment variables file.
├── go.mod                 # Go module definition.
├── cmd/
│   └── main.go            # Main entry point for the server.
└── pkg/
    ├── config.go          # Configuration: loads .env and sets global variables.
    ├── handlers.go        # HTTP handlers for endpoints (/status, /attestation, etc.).
    └── middleware.go      # Logging middleware.
```

## Configuration

The server configuration is managed in the `pkg/config.go` file. It uses [godotenv](https://github.com/joho/godotenv) to load environment variables from a `.env` file. Key configuration variables include:

### Server Configuration
- **SECRETAI_REPORT_DIR**: Directory where attestation report files are stored (default: `reports`).
- **SECRETAI_REST_SERVER_IP**: The IP address on which the server listens (default: `0.0.0.0`).
- **SECRETAI_SECURE**: Boolean indicating whether to enable HTTPS (default: `true`).
- **SECRETAI_REST_SERVER_PORT**: Port for the server (default: `29343`).
- **SECRETAI_CERT_PATH**: Path to SSL certificate file (default: `cert/ssl_cert.pem`).
- **SECRETAI_KEY_PATH**: Path to SSL key file (default: `cert/ssl_key.pem`).

### Attestation Configuration
- **SECRETAI_ATTEST_TOOL**: Command name for the attestation tool (default: `attest_tool`).
- **SECRETAI_ATTEST_TIMEOUT_SEC**: Timeout in seconds for attestation command execution (default: `10`).

### Attestation File Names
- **SECRETAI_GPU_ATTESTATION_FILE**: Filename for GPU attestation reports (default: `gpu_attestation.txt`).
- **SECRETAI_CPU_ATTESTATION_FILE**: Filename for CPU (TDX) attestation reports (default: `tdx_attestation.txt`).
- **SECRETAI_SELF_ATTESTATION_FILE**: Filename for self attestation reports (default: `self_report.txt`).

For example, your `.env` file might look like this:

```
SECRETAI_REPORT_DIR=reports
SECRETAI_REST_SERVER_IP=0.0.0.0
SECRETAI_SECURE=true
SECRETAI_REST_SERVER_PORT=29343
SECRETAI_CERT_PATH=cert/ssl_cert.pem
SECRETAI_KEY_PATH=cert/ssl_key.pem
SECRETAI_ATTEST_TOOL=attest_tool
SECRETAI_ATTEST_TIMEOUT_SEC=10
```

## Installation and Running

1. **Clone the repository:**

   ```bash
   git clone https://github.com/scrtlabs/secretai-attest-rest.git
   cd secretai-attest-rest
   ```

2. **Set up your environment:**

   Make sure you have a valid `.env` file in the project root with your desired settings.

3. **Build and run the server:**

   To run using the Go command-line tool, execute:

   ```bash
   go run cmd/main.go
   ```

   Alternatively, build the binary:

   ```bash
   go build -o secretai-attest-rest cmd/main.go
   ./secretai-attest-rest --secure=true --port=29343 --ip=0.0.0.0
   ```

4. **Run tests:**

   To run all tests:

   ```bash
   go test ./...
   ```

   To run a specific test:

   ```bash
   go test -run TestStatusHandler ./pkg
   ```

   With verbose output:

   ```bash
   go test -v ./pkg
   ```

## API Endpoints

### `/status`
- **Method:** GET  
- **Description:** Returns a JSON object indicating that the server is alive.
- **Response Example:**

  ```json
  {
    "status": "server is alive"
  }
  ```

### `/gpu`, `/cpu`, `/self`
- **Method:** GET  
- **Description:** Reads the corresponding attestation file from the configured report directory and returns its content as plain text.
- **Error Handling:**
  - Returns a JSON error if the file is missing or cannot be read.
