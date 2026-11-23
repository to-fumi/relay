package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"relay/internal/platform/httpx"
)

func TestSecurityHeaders(t *testing.T) {
	h := httpx.SecurityHeaders(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))
	for k, want := range map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	} {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("%s = %q; want %q", k, got, want)
		}
	}
}
