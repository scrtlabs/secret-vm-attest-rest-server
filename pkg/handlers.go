package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
		"time": time.Now().Format(time.RFC3339),
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
