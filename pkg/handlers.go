package pkg

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	htmlpkg "secret-vm-attest-rest-server/pkg/html"
	"sort"
	"strconv"
	"strings"
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

	env := EnvValue
	if env == "" {
		env = "unknown"
	}

	status, err := getStatus()
	if err != nil {
		status = "server_error"
	}

	response := map[string]string{
		"status": status,
		"time":   time.Now().Format(time.RFC3339),
		"env":    env,
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

		// Determine Title and Description per type
		var titleText, descText string
		if attestationType == "Self" {
			titleText = fmt.Sprintf("%s Attestation Report", attestationType)
			descText = fmt.Sprintf("Below is the %s attestation report. Click the copy button to copy it to your clipboard.", attestationType)
		} else if attestationType == "Proof of Cloud" {
			titleText = "Proof of Cloud JWT"
			descText = "Below is the Proof of Cloud JWT. Click the copy button to copy it to your clipboard."
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
		tmpl, err := template.New("attestationHtml").Parse(htmlpkg.HtmlTemplate)
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

// MakePublicKeyHandler serves a public key file as plain text
func MakePublicKeyHandler(filePath, keyType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			respondWithError(w, http.StatusNotFound,
				fmt.Sprintf("%s public key not available", keyType), err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

func MakePublicKeyHTMLHandler(filePath, keyType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			respondWithError(w, http.StatusNotFound,
				fmt.Sprintf("%s public key not available", keyType), err.Error())
			return
		}

		data := struct {
			Title       string
			Description string
			Quote       string
			ShowVerify  bool
		}{
			Title:       fmt.Sprintf("%s Public Key", keyType),
			Description: "Below is the public key. Click the copy button to copy it.",
			Quote:       string(content),
			ShowVerify:  false,
		}

		tmpl, err := template.New("publicKey").Parse(htmlpkg.HtmlTemplate)
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

		// Serve as HTML with pre tag to preserve trailing newlines
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><style>body{background:#1e1e1e;color:#e0e0e0;margin:0;padding:10px;font-family:monospace}pre{margin:0;white-space:pre}</style></head><body><pre>%s&#8203;</pre><script>document.oncopy=function(e){var s=window.getSelection().toString();if(s.endsWith('\u200B')){e.preventDefault();e.clipboardData.setData('text/plain',s.slice(0,-1))}}</script></body></html>`, html.EscapeString(string(content)))
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
		tmpl, err := template.New("dockerCompose").Parse(htmlpkg.HtmlTemplate)
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
		du, _ := disk.Usage(FsMountPath) // Total and Used in bytes :contentReference[oaicite:5]{index=5}

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

// MakeVMUpdatesHandler serves raw JSON at /vm_updates
func MakeVMUpdatesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}

		sid := ServiceIDValue
		if sid == "" {
			respondWithJSON(w, http.StatusOK, map[string]string{
				"message": "VM is not upgradeable",
			})
			return
		}

		// Determine the query command based on the hardware environment.
		// If /dev/sev-guest exists, we are running on AMD SEV-SNP.
		queryCmd := "list_image_filters" // Default to TDX
		if _, err := os.Stat("/dev/sev-guest"); err == nil {
			queryCmd = "list_image_filters_amd"
		}

		// Call external kms-query binary
		cmd := exec.Command("kms-query", queryCmd, sid)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			respondWithError(w, http.StatusInternalServerError,
				"failed to query contract", err.Error()+": "+out.String())
			return
		}

		// Validate JSON
		var js json.RawMessage
		if err := json.Unmarshal(out.Bytes(), &js); err != nil {
			respondWithError(w, http.StatusInternalServerError,
				"invalid JSON from kms-query", err.Error())
			return
		}

		// Return as-is
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(out.Bytes())
	}
}

// MakeVMUpdatesHTMLHandler serves the page at /vm_updates.html
func MakeVMUpdatesHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed,
				"Method not allowed", "Only GET requests are supported")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := htmlpkg.VMUpdatesTmpl.ExecuteTemplate(w, "vm_updates.html", nil); err != nil {
			log.Printf("template execute error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// MakeVMLogsHandler implements /logs endpoint with sorting, filtering, and service selection.
//
// - ?service=secretvm → only system logs (journalctl)
// - ?service={container} → only docker logs for that container
// - no service param → system logs + all docker logs
//
// All lines are sorted by timestamp. Output format matches journalctl (no year).
func MakeVMLogsHandler(secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		// Default number of log lines for docker containers
		lines := 1000
		if l := r.URL.Query().Get("lines"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 {
				lines = v
			}
		}

		service := r.URL.Query().Get("service")

		var collected []LogLine
		hostName, _ := os.Hostname()

		// Collect system logs if service is not specified or equals "secretvm"
		if service == "" || service == "secretvm" {
			if sysLogs, err := fetchServicesLogs(); err == nil {
				collected = append(collected, parseJournalOutput(sysLogs)...)
			}
			if service == "secretvm" {
				sort.Slice(collected, func(i, j int) bool { return collected[i].Timestamp.Before(collected[j].Timestamp) })
				writeLogsResponse(w, collected)
				return
			}
		}

		// Collect docker logs (only in secure mode)
		if secure {
			if service != "" && service != "secretvm" {
				// Logs for a single container
				if ll, err := fetchDockerLogsForContainer(service, lines, hostName); err == nil {
					sort.Slice(ll, func(i, j int) bool { return ll[i].Timestamp.Before(ll[j].Timestamp) })
					writeLogsResponse(w, ll)
					return
				}
			} else {
				// Logs for all containers
				if ll, err := fetchDockerLogsAll(lines, hostName); err == nil {
					collected = append(collected, ll...)
				}
			}
		}

		// Sort combined logs and return
		sort.Slice(collected, func(i, j int) bool {
			return collected[i].Timestamp.Before(collected[j].Timestamp)
		})
		writeLogsResponse(w, collected)
	}
}

// MakeServicesHandler implements /services endpoint.
// It returns a JSON array containing "secretvm" plus all docker containers (running and stopped).
func MakeServicesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}
		names, err := listAllDockerContainerNames()
		if err != nil {
			respondWithJSON(w, http.StatusOK, []string{"secretvm"})
			return
		}
		services := append([]string{"secretvm"}, names...)
		respondWithJSON(w, http.StatusOK, services)
	}
}

// // MakeVMLiveLogsHandler serves the live-updating HTML page
// // that polls /logs with a selectable line count.
// func MakeVMLiveLogsHandler() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet {
// 			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
// 			return
// 		}
// 		data := struct{ Title string }{Title: "Live Docker Container Logs"}
// 		tmpl, err := template.New("liveLogs").Parse(html.DockerLiveLogsTemplate)
// 		if err != nil {
// 			log.Printf("Error parsing live logs template: %v", err)
// 			http.Error(w, "Internal server error", http.StatusInternalServerError)
// 			return
// 		}
// 		w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 		if err := tmpl.Execute(w, data); err != nil {
// 			log.Printf("Error executing live logs template: %v", err)
// 			http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		}
// 	}
// }

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

type ItaTokenResponse struct {
	KeyName string `json:"key_name"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

func fetchItaJwt() ([]ItaTokenResponse, error, int) {
	if len(ItaKeys) == 0 {
		return nil, fmt.Errorf("no ITA API keys configured"), http.StatusInternalServerError
	}
	if len(ItaKeys) > 3 {
		return nil, fmt.Errorf("too many ITA API keys configured (max 3)"), http.StatusBadRequest
	}

	quoteFilePath := filepath.Join(ReportDir, CPUAttestationFile)
	rawQuote, err := os.ReadFile(quoteFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read raw quote: %v", err), http.StatusInternalServerError
	}

	hexStr := strings.TrimSpace(string(rawQuote))
	quoteBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex quote: %v", err), http.StatusInternalServerError
	}
	b64Quote := base64.StdEncoding.EncodeToString(quoteBytes)

	var results []ItaTokenResponse
	client := &http.Client{Timeout: 10 * time.Second}

	for keyName, keyInfo := range ItaKeys {
		if keyInfo.ApiKey == "" || keyInfo.PolicyId == "" {
			results = append(results, ItaTokenResponse{KeyName: keyName, Error: "ITA API Key or Policy ID is empty"})
			continue
		}

		payload := map[string]interface{}{
			"quote":      b64Quote,
			"policy_ids": []string{keyInfo.PolicyId},
		}
		payloadBytes, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", ItaApiUrl, bytes.NewBuffer(payloadBytes))
		if err != nil {
			results = append(results, ItaTokenResponse{KeyName: keyName, Error: fmt.Sprintf("failed to create request: %v", err)})
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("x-api-key", keyInfo.ApiKey)

		resp, err := client.Do(req)
		if err != nil {
			results = append(results, ItaTokenResponse{KeyName: keyName, Error: fmt.Sprintf("failed to contact ITA API: %v", err)})
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			results = append(results, ItaTokenResponse{KeyName: keyName, Error: fmt.Sprintf("failed to read ITA response: %v", readErr)})
			continue
		}

		if resp.StatusCode != http.StatusOK {
			results = append(results, ItaTokenResponse{KeyName: keyName, Error: fmt.Sprintf("ITA API returned status %d: %s", resp.StatusCode, string(respBody))})
			continue
		}

		var itaResp struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(respBody, &itaResp); err != nil || itaResp.Token == "" {
			results = append(results, ItaTokenResponse{KeyName: keyName, Error: "invalid ITA response format or empty token"})
			continue
		}

		results = append(results, ItaTokenResponse{KeyName: keyName, Token: itaResp.Token})
	}

	return results, nil, http.StatusOK
}

// MakeItaJwtHandler dynamically fetches the ITA JWT token(s)
func MakeItaJwtHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		tokens, err, code := fetchItaJwt()
		if err != nil {
			respondWithError(w, code, "Failed to fetch ITA JWT(s)", err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, tokens)
	}
}

// MakeItaJwtHTMLHandler dynamically fetches and renders the ITA JWT token(s).
func MakeItaJwtHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		tokens, err, code := fetchItaJwt()
		if err != nil {
			respondWithError(w, code, "Failed to fetch ITA JWT", err.Error())
			return
		}

		type JwtItem struct {
			Title   string
			Content string
		}
		var items []JwtItem
		for _, t := range tokens {
			content := t.Token
			if content == "" && t.Error != "" {
				content = "Error: " + t.Error
			}
			items = append(items, JwtItem{
				Title:   t.KeyName,
				Content: content,
			})
		}

		data := struct {
			Title       string
			Description string
			Items       []JwtItem
		}{
			Title:       "Intel Trust Authority JWTs",
			Description: "Below are the dynamically fetched Intel Trust Authority JWTs across all configured policies.",
			Items:       items,
		}

		tmpl, err := template.New("itaJwtHtml").Parse(htmlpkg.MultiItemHtmlTemplate)
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

func fetchPocJwt() (string, error, int) {
	quoteFilePath := filepath.Join(ReportDir, CPUAttestationFile)
	if _, err := os.Stat(quoteFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("quote file not found: %s", quoteFilePath), http.StatusInternalServerError
	}

	cmd := exec.Command("get_poc_token.sh", quoteFilePath)
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute get_poc_token.sh: %v. Stderr: %s", err, errBuffer.String()), http.StatusInternalServerError
	}

	var result struct {
		Jwt string `json:"jwt"`
	}
	if err := json.Unmarshal(outBuffer.Bytes(), &result); err != nil {
		return "", fmt.Errorf("invalid json from get_poc_token.sh output: %v. Output: %s", err, outBuffer.String()), http.StatusInternalServerError
	}
	if result.Jwt == "" {
		return "", fmt.Errorf("empty jwt returned by get_poc_token.sh"), http.StatusInternalServerError
	}

	return result.Jwt, nil, http.StatusOK
}

// MakePocJwtHandler dynamically fetches the PoC JWT token.
func MakePocJwtHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		token, err, code := fetchPocJwt()
		if err != nil {
			respondWithError(w, code, "Failed to fetch Proof of Cloud JWT", err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(token))
	}
}

// MakePocJwtHTMLHandler dynamically fetches and renders the PoC JWT token.
func MakePocJwtHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET requests are supported")
			return
		}

		token, err, code := fetchPocJwt()
		if err != nil {
			respondWithError(w, code, "Failed to fetch Proof of Cloud JWT", err.Error())
			return
		}

		data := struct {
			Title       string
			Description string
			Quote       string
			ShowVerify  bool
		}{
			Title:       "Proof of Cloud JWT",
			Description: "Below is the dynamically fetched Proof of Cloud JWT. Click the copy button to copy it.",
			Quote:       token,
			ShowVerify:  false,
		}

		tmpl, err := template.New("pocJwtHtml").Parse(htmlpkg.HtmlTemplate)
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
