package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrivateGuardRequiresTokenWhenEndpointIsNotOpenedByMask(t *testing.T) {
	oldPrivateMode := PrivateMode
	oldMask := EndpointsMask
	oldToken := AccessToken
	PrivateMode = true
	EndpointsMask = "01010"
	AccessToken = ""
	t.Cleanup(func() {
		PrivateMode = oldPrivateMode
		EndpointsMask = oldMask
		AccessToken = oldToken
	})

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rr := httptest.NewRecorder()
	PrivateGuard(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("guarded handler should not run without an explicit token or mask opening")
	}).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestPrivateGuardAllowsEndpointOpenedByMask(t *testing.T) {
	oldPrivateMode := PrivateMode
	oldMask := EndpointsMask
	oldToken := AccessToken
	PrivateMode = true
	EndpointsMask = "01010"
	AccessToken = ""
	t.Cleanup(func() {
		PrivateMode = oldPrivateMode
		EndpointsMask = oldMask
		AccessToken = oldToken
	})

	req := httptest.NewRequest(http.MethodGet, "/docker-compose", nil)
	rr := httptest.NewRecorder()
	PrivateGuard(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}
