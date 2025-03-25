package pkg

import (
	"log"
	"net/http"
	"time"
)

// ResponseWriter is a wrapper for http.ResponseWriter that captures the status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and passes it to the original ResponseWriter
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs the remote address, requested URL, HTTP method, and response status code
// for each incoming HTTP request, along with the time taken to process the request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()
		
		// Create a wrapper for the response writer to capture the status code
		wrapped := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process the request
		next.ServeHTTP(wrapped, r)
		
		// Calculate request processing time
		duration := time.Since(start)
		
		// Log the request details including status code and duration
		log.Printf("%s | %s %s | %d | %s | %v", 
			r.RemoteAddr, 
			r.Method, 
			r.URL.String(), 
			wrapped.statusCode,
			http.StatusText(wrapped.statusCode),
			duration)
	})
}

// SecurityHeadersMiddleware adds security-related HTTP headers to all responses.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers to responses.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
