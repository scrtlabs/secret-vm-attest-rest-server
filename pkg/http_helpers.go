// pkg/http_helpers.go
package pkg

import (
	"net/http"
	"strings"
)

// groups: one bit controls both plain and html variants
var endpointBits = map[string]int{
	"/logs":                0,
	"/docker-compose":      1,
	"/docker-compose.html": 1,
	"/services":            2,
	"/vm_upgrades":         3,
	"/vm_upgrades.html":    3,
	"/resources":           4,
	"/resources.html":      4,
}

// leftmost char -> bit 0
func bitIsOpen(mask string, idx int) bool {
	if mask == "" || idx < 0 {
		return false
	}
	if idx >= len(mask) {
		return false
	}
	return mask[idx] == '1'
}

// Authorization: Bearer <token> | X-Dev-Token: <token> | ?token=<token>
func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		low := strings.ToLower(auth)
		if strings.HasPrefix(low, "bearer ") && len(auth) > 7 {
			return strings.TrimSpace(auth[7:])
		}
	}
	if h := r.Header.Get("X-Dev-Token"); h != "" {
		return h
	}
	if q := r.URL.Query().Get("token"); q != "" {
		return q
	}
	return ""
}

func PrivateGuard(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !PrivateMode {
			h.ServeHTTP(w, r)
			return
		}
		idx, ok := endpointBits[r.URL.Path]
		if !ok {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized: provide Bearer token, X-Dev-Token, or ?token"))
			return
		}

		if bitIsOpen(EndpointsMask, idx) {
			h.ServeHTTP(w, r)
			return
		}

		token := extractToken(r)
		if AccessToken != "" && token == AccessToken {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized: invalid or missing Bearer/X-Dev-Token/?token"))
	}
}
