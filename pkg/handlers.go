package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// StatusHandler handles the /status endpoint and returns a simple JSON status message.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "server is alive"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AttestationHandler executes an external command (e.g., "attest_tool")
// and returns its JSON output. If the command fails or the output is not valid JSON,
// an error message is returned.
func AttestationHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("attest_tool")
	output, err := cmd.Output()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to generate attestation.",
		})
		return
	}

	// Validate that the output is valid JSON.
	var js map[string]interface{}
	if err := json.Unmarshal(output, &js); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid attestation format.",
		})
		return
	}

	// Return the valid JSON output.
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// MakeAttestationFileHandler returns an HTTP handler function that reads an attestation file.
// It is used for the /gpu, /cpu, and /self endpoints.
func MakeAttestationFileHandler(fileName, attestationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Construct the full file path.
		filePath := filepath.Join(ReportDir, fileName)

		// Check if the file exists.
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   fmt.Sprintf("%s attestation not available", attestationType),
				"details": fmt.Sprintf("The %s attestation data has not been generated or is not ready yet", attestationType),
			})
			return
		}

		// Read the file content.
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   fmt.Sprintf("Failed to retrieve %s attestation data", attestationType),
				"details": err.Error(),
			})
			return
		}

		// Return the file content as plain text.
		w.Header().Set("Content-Type", "text/plain")
		w.Write(content)
	}
}
