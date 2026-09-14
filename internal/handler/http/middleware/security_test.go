package middleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/middleware"
)

func TestSecurityHeadersApplyBaseline(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil))

	expected := map[string]string{
		"X-Content-Type-Options":        "nosniff",
		"X-Frame-Options":               "DENY",
		"Referrer-Policy":               "strict-origin-when-cross-origin",
		"Cross-Origin-Resource-Policy":  "cross-origin",
		"Content-Security-Policy":       middleware.DefaultContentSecurityPolicy,
		"Permissions-Policy":            "geolocation=(self), camera=(self), microphone=()",
	}

	for header, want := range expected {
		if got := recorder.Header().Get(header); got != want {
			t.Errorf("header %s = %q, want %q", header, got, want)
		}
	}

	if got := recorder.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS must not be sent over plain HTTP, got %q", got)
	}
}

func TestSecurityHeadersSetHSTSOnTLSConnections(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil)
	request.TLS = &tls.ConnectionState{}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Strict-Transport-Security"); got != "max-age=31536000; includeSubDomains" {
		t.Errorf("got HSTS %q, want %q", got, "max-age=31536000; includeSubDomains")
	}
}

func TestSecurityHeadersSetHSTSBehindTLSTerminatingProxy(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil)
	request.Header.Set("X-Forwarded-Proto", "HTTPS")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Strict-Transport-Security"); got == "" {
		t.Error("expected HSTS to be set when the proxy terminated TLS")
	}
}

func TestSecurityHeadersWithCustomPolicy(t *testing.T) {
	const policy = "default-src 'none'"

	handler := middleware.SecurityHeadersWithCSP(policy)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil))

	if got := recorder.Header().Get("Content-Security-Policy"); got != policy {
		t.Errorf("got CSP %q, want %q", got, policy)
	}
	if got := recorder.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("custom CSP must not drop the other headers, got X-Frame-Options=%q", got)
	}
}

func TestSecurityHeadersWithoutCSP(t *testing.T) {
	handler := middleware.SecurityHeadersWithCSP("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil))

	if got := recorder.Header().Get("Content-Security-Policy"); got != "" {
		t.Errorf("empty policy must skip the header, got %q", got)
	}
}

func TestSecurityHeadersDoNotOverrideHandlerValues(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/plans", nil))

	// The middleware sets headers before calling the handler, so a handler that
	// deliberately wants a looser policy for an embeddable endpoint still wins.
	if got := recorder.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("got X-Frame-Options %q, want %q", got, "SAMEORIGIN")
	}
}