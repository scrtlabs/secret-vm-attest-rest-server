package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"secret-vm-attest-rest-server/pkg/html"
	"text/template"
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
