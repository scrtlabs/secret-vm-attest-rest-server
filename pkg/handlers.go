package pkg

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"secret-vm-attest-rest-server/pkg/html"
	"strconv"
	"text/template"
	"time"

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
)

// StatusHandler handles the /status endpoint and returns a simple JSON status message.
// Only GET requests are accepted.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
		return
	}

	response := map[string]string{
		"status": "server is alive",
		"time":   time.Now().Format(time.RFC3339),
	}
	respondWithJSON(w, http.StatusOK, response)
}

// MakeAttestationFileHandler returns an HTTP handler function that reads an attestation file.
// It is used for the /gpu, /cpu, and /self endpoints. Only GET requests are accepted.
func MakeAttestationFileHandler(fileName, attestationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		// Construct the full file path.
		filePath := filepath.Join(ReportDir, fileName)

		// Check if the file exists.
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("%s attestation file not found: %s", attestationType, filePath)
			respondWithError(w, http.StatusNotFound,
				fmt.Sprintf("%s attestation not available", attestationType),
				fmt.Sprintf("The %s attestation data has not been generated or is not ready yet", attestationType))
			return
		}

		// Read the file content.
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading %s attestation file: %v", attestationType, err)
			respondWithError(w, http.StatusInternalServerError,
				fmt.Sprintf("Failed to retrieve %s attestation data", attestationType),
				err.Error())
			return
		}

		// Return the file content as plain text with proper headers.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// MakeAttestationHTMLHandler returns an HTTP handler that serves an HTML page for the attestation quote.
// It reads the attestation file from the reports directory, parses the HTML template,
// and renders the page with a dynamic title, description, and the attestation quote.
func MakeAttestationHTMLHandler(fileName, attestationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow only GET requests.
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		// Construct the full path to the attestation file.
		filePath := filepath.Join(ReportDir, fileName)

		// Check if the file exists.
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("%s attestation file not found: %s", attestationType, filePath)
			respondWithError(w, http.StatusNotFound,
				fmt.Sprintf("%s attestation not available", attestationType),
				fmt.Sprintf("The %s attestation data has not been generated or is not ready yet", attestationType))
			return
		}

		// Read the attestation file content.
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading %s attestation file: %v", attestationType, err)
			respondWithError(w, http.StatusInternalServerError,
				fmt.Sprintf("Failed to retrieve %s attestation data", attestationType),
				err.Error())
			return
		}

		// Determine Title and Description per type (Self uses "Report")
		var titleText, descText string
		if attestationType == "Self" {
			titleText = fmt.Sprintf("%s Attestation Report", attestationType)
			descText = fmt.Sprintf("Below is the %s attestation report. Click the copy button to copy it to your clipboard.", attestationType)
		} else {
			titleText = fmt.Sprintf("%s Attestation Quote", attestationType)
			descText = fmt.Sprintf("Below is the %s attestation quote. Click the copy button to copy it to your clipboard.", attestationType)
		}

		// Prepare template data
		data := struct {
			Title       string
			Description string
			Quote       string
			ShowVerify  bool
		}{
			Title:       titleText,
			Description: descText,
			Quote:       string(content),
			ShowVerify:  attestationType == "CPU", // only CPU shows verification link
		}

		// Parse and execute HTML template
		tmpl, err := template.New("attestationHtml").Parse(html.HtmlTemplate)
		if err != nil {
			log.Printf("Error parsing HTML template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing HTML template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// MakeDockerComposeFileHandler returns a handler that serves the raw docker-compose file.
func MakeDockerComposeFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}

		path := DockerComposePath
		// Ensure the path is configured
		if path == "" {
			respondWithError(w, http.StatusInternalServerError,
				"Configuration error", "SECRETVM_DOCKER_COMPOSE_PATH is not set")
			return
		}

		// Read the file from disk
		content, err := os.ReadFile(path)
		if err != nil {
			respondWithError(w, http.StatusNotFound,
				"File not found", fmt.Sprintf("Could not read file %s: %v", path, err))
			return
		}

		// Serve as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// MakeDockerComposeHTMLHandler returns a handler that renders the docker-compose file in HTML.
func MakeDockerComposeHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}

		path := DockerComposePath
		// Read docker-compose content
		content, err := os.ReadFile(path)
		if err != nil {
			respondWithError(w, http.StatusNotFound,
				"File not found", fmt.Sprintf("Could not read file %s: %v", path, err))
			return
		}

		// Prepare data for the template
		data := struct {
			Title       string
			Description string
			Quote       string
			ShowVerify  bool
		}{
			Title:       "Docker Compose File",
			Description: "Below is the docker-compose configuration. Click the copy button to copy it.",
			Quote:       string(content),
			ShowVerify:  false, // no verification link needed
		}

		// Parse and execute shared HTML template
		tmpl, err := template.New("dockerCompose").Parse(html.HtmlTemplate)
		if err != nil {
			log.Printf("Error parsing HTML template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing HTML template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// ResourceStats holds both raw percentages and GB values as numbers.
type ResourceStats struct {
	MemoryUsedGB  float64 `json:"memory_used_gb"`
	MemoryTotalGB float64 `json:"memory_total_gb"`
	DiskUsedGB    float64 `json:"disk_used_gb"`
	DiskTotalGB   float64 `json:"disk_total_gb"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	CPUPercent    float64 `json:"cpu_percent"`
}

func MakeResourcesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}

		// Memory stats
		vm, _ := mem.VirtualMemory() // Total and Used in bytes :contentReference[oaicite:4]{index=4}

		// Disk stats for root
		du, _ := disk.Usage("/") // Total and Used in bytes :contentReference[oaicite:5]{index=5}

		// CPU percent
		cpus, _ := cpu.Percent(2000*time.Millisecond, false)
		cpuPct := cpus[0] // aggregate CPU usage :contentReference[oaicite:6]{index=6}

		// Convert bytes → GB and round to 3 decimal places
		toGB := func(b uint64) float64 {
			gb := float64(b) / (1024 * 1024 * 1024)
			return math.Round(gb*1000) / 1000
		}

		stats := ResourceStats{
			MemoryUsedGB:  toGB(vm.Used),
			MemoryTotalGB: toGB(vm.Total),
			DiskUsedGB:    toGB(du.Used),
			DiskTotalGB:   toGB(du.Total),
			MemoryPercent: vm.UsedPercent,
			DiskPercent:   du.UsedPercent,
			CPUPercent:    cpuPct,
		}

		respondWithJSON(w, http.StatusOK, stats)
	}
}

//go:embed html/resources.html
var resourcesFS embed.FS

var resourcesTmpl = template.Must(
	template.New("resources").
		ParseFS(resourcesFS, "html/resources.html"),
)

// MakeResourcesHTMLHandler serves the embedded HTML template.
func MakeResourcesHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := resourcesTmpl.ExecuteTemplate(w, "resources.html", nil); err != nil {
			log.Printf("template execute error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// MakeVMLogsHandler serves plain-text VM logs (including docker logs),
// requiring the client to specify either a container name or an index.
func MakeVMLogsHandler(secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only GET allowed
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		// Parse desired number of lines (default 1000)
		lines := 1000
		if l := r.URL.Query().Get("lines"); l != "" {
			if v, err := strconv.Atoi(l); err == nil {
				lines = v
			}
		}

		// Extract name and index parameters
		name := r.URL.Query().Get("name")
		idxStr := r.URL.Query().Get("index")
		useIndex := false
		index := 0
		if idxStr != "" {
			if i, err := strconv.Atoi(idxStr); err == nil {
				index = i
				useIndex = true
			}
		}

		// Fetch logs, honoring error if neither selector nor valid container
		logs, err := fetchServicesLogs()
		if err != nil {
			log.Printf("Error fetching system logs: %v", err)
		}
		if secure {
			// fetch docker logs only in https mode, because during http mode
			// the docker is still configuring
			out, err := fetchDockerLogsWithSelector(name, index, useIndex, lines)
			if err != nil {
				log.Printf("Error fetching Docker logs: %v", err)
			}
			logs += out
		}

		// Return logs as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(logs))
	}
}

// MakeVMLiveLogsHandler serves the live-updating HTML page
// that polls /logs with a selectable line count.
func MakeVMLiveLogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}
		data := struct{ Title string }{Title: "Live Docker Container Logs"}
		tmpl, err := template.New("liveLogs").Parse(html.DockerLiveLogsTemplate)
		if err != nil {
			log.Printf("Error parsing live logs template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing live logs template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// Helper function to respond with a JSON error message.
func respondWithError(w http.ResponseWriter, code int, error string, details string) {
	respondWithJSON(w, code, map[string]string{
		"error":   error,
		"details": details,
	})
}

// Helper function to respond with JSON data.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Convert payload to JSON
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error","details":"Failed to generate response"}`))
		return
	}

	// Set response headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(response)
}
