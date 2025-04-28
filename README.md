# SecretVM Attest REST Server (Go)

SecretVM Attest REST Server is a lightweight REST server implemented in Go. It provides attestation reports for confidential virtual machines (VMs) over HTTPS. The server exposes multiple endpoints to return different attestation reports, including:

- **/status** – Returns the server status.
- **/attestation** – Executes an internal process (e.g., `attest_tool`) and returns a JSON attestation report.
- **/gpu** – Returns the NVIDIA confidential GPU attestation report.
- **/cpu** – Returns the Intel TDX attestation report.
- **/self** – Returns self attestation data (e.g., TDX measurement registers).
- **/docker_logs** – Return Docker container logs as plain text. **Requires** exactly one of:
  - `name` (container name), or  
  - `index` (zero-based container index).  
  Optional `lines` parameter (100, 500, 1000; default 1000).

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
- **Attestation Report Visualization**: Offers HTML endpoints (e.g., `/gpu.html`) that display attestation reports in a web page format. These pages are styled for readability and include an easy copy-to-clipboard option for the report content.
- **Docker Log Monitoring**: Provides an endpoint to fetch Docker container logs and a live web interface for real-time log viewing. The web interface (`/docker_logs.html`) features a dark theme, auto-scrolling, log length selection, and copy-to-clipboard functionality.

## Project Structure

```
secret-vm-attest-rest-server/
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
- **SECRETVM_REPORT_DIR**: Directory where attestation report files are stored (default: `reports`).
- **SECRETVM_REST_SERVER_IP**: The IP address on which the server listens (default: `0.0.0.0`).
- **SECRETVM_SECURE**: Boolean indicating whether to enable HTTPS (default: `true`).
- **SECRETVM_REST_SERVER_PORT**: Port for the server (default: `29343`).
- **SECRETVM_CERT_PATH**: Path to SSL certificate file (default: `cert/ssl_cert.pem`).
- **SECRETVM_KEY_PATH**: Path to SSL key file (default: `cert/ssl_key.pem`).

### Attestation Configuration
- **SECRETVM_ATTEST_TOOL**: Command name for the attestation tool (default: `attest_tool`).
- **SECRETVM_ATTEST_TIMEOUT_SEC**: Timeout in seconds for attestation command execution (default: `10`).

### Attestation File Names
- **SECRETVM_GPU_ATTESTATION_FILE**: Filename for GPU attestation reports (default: `gpu_attestation.txt`).
- **SECRETVM_CPU_ATTESTATION_FILE**: Filename for CPU (TDX) attestation reports (default: `tdx_attestation.txt`).
- **SECRETVM_SELF_ATTESTATION_FILE**: Filename for self attestation reports (default: `self_report.txt`).

For example, your `.env` file might look like this:

```
SECRETVM_REPORT_DIR=reports
SECRETVM_REST_SERVER_IP=0.0.0.0
SECRETVM_SECURE=true
SECRETVM_REST_SERVER_PORT=29343
SECRETVM_CERT_PATH=cert/ssl_cert.pem
SECRETVM_KEY_PATH=cert/ssl_key.pem
SECRETVM_ATTEST_TOOL=attest_tool
SECRETVM_ATTEST_TIMEOUT_SEC=10
```

## Installation and Running

1. **Clone the repository:**

   ```bash
   git clone https://github.com/scrtlabs/secret-vm-attest-rest-server.git
   cd secret-vm-attest-rest-server
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
   go build -o secret-vm-attest-rest-server cmd/main.go
   ./secret-vm-attest-rest-server --secure=true --port=29343 --ip=0.0.0.0
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

### `/gpu.html`, `/cpu.html`, `/self.html`

- **Method:** GET  
- **Description:** Renders the content of the corresponding attestation report file (GPU, CPU, or self) in a human-friendly HTML page. The page is styled for better readability and includes a copy-to-clipboard button to easily copy the report content.  
- **Error Handling:**  
  - If the attestation file is missing or cannot be read, an error message is displayed on the page (e.g., indicating the file could not be retrieved).

### `/docker_logs`

- **Method:** GET  
- **Description:** Retrieves the latest log entries from a Docker container (intended for debugging/monitoring). By default, this endpoint targets the container named `secret-vm-docker`. If that container is not found, it will fall back to the first running container.  
- **Query Parameters:**
  - `name` (string) – exact container name (takes priority over `index`)
  - `index` (integer) – zero-based index into the `docker ps` list
  - `lines` (optional) – number of lines to fetch (100, 500, 1000; default 1000)
- **Response:** Plain text output of the requested log lines.  
**Error Handling:**  
  - **400 Bad Request** if neither `name` nor `index` is provided.  
  - **404 Not Found** if no running container matches the specified `name` or `index`.  

### `/docker_logs.html`

- **Method:** GET  
**Description:**  
  - A radio selector lets you choose exactly one mode: **By Name** or **By Index**.  
  - Enter the container name or index, then click **Apply**.  
  - After **Apply**, the page will auto-refresh the logs every 2 seconds, **only** if the log view is scrolled to the bottom.  
  - The **Copy Logs** button copies the currently displayed logs to your clipboard.  
  - If the container isn’t found or the input field is left empty, the interface displays the full error message returned by the server.

