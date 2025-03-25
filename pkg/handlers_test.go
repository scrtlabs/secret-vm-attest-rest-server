package pkg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(StatusHandler)

	// Call the handler directly and record the response
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the content type
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}

	// Check the response body
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}

	// Verify the status field exists
	if status, exists := response["status"]; !exists || status != "server is alive" {
		t.Errorf("Response missing or incorrect status field: %v", response)
	}

	// Verify the time field exists
	if _, exists := response["time"]; !exists {
		t.Errorf("Response missing time field: %v", response)
	}
}

func TestStatusHandlerInvalidMethod(t *testing.T) {
	// Create a request with an invalid method
	req, err := http.NewRequest("POST", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(StatusHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}

	// Check the content type
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}

	// Check the response body
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}

	// Verify the error field exists
	if errorMsg, exists := response["error"]; !exists || errorMsg != "Method not allowed" {
		t.Errorf("Response missing or incorrect error field: %v", response)
	}
}