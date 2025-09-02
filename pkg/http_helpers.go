// pkg/http_helpers.go
package pkg

import "net/http"

// PrivateGuard blocks handler when PrivateMode is enabled.
func PrivateGuard(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if PrivateMode {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("VM is running in private mode. Endpoint isn't available."))
			return
		}
		h.ServeHTTP(w, r)
	}
}
