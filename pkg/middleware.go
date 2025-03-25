package pkg

import (
	"log"
	"net/http"
)

// LoggingMiddleware logs the remote address and requested URL for each incoming HTTP request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request from %s -> %s", r.RemoteAddr, r.URL.String())
		next.ServeHTTP(w, r)
	})
}
