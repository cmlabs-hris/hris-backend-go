package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/middleware"
	"github.com/go-chi/chi/v5"
)

func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func newThrottledRouter(limiter *middleware.RateLimiter) http.Handler {
	router := chi.NewRouter()
	router.Use(limiter.Middleware())
	router.Get("/ping", okHandler())
	return router
}

func serve(t *testing.T, handler http.Handler, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestRateLimiterAllowsRequestsUpToBurst(t *testing.T) {
	limiter := middleware.NewRateLimiter(1, 3, time.Minute)
	router := newThrottledRouter(limiter)

	for i := 1; i <= 3; i++ {
		recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("request %d: got status %d, want %d", i, recorder.Code, http.StatusOK)
		}
	}
}

func TestRateLimiterRejectsRequestsBeyondBurst(t *testing.T) {
	limiter := middleware.NewRateLimiter(1, 2, time.Minute)
	router := newThrottledRouter(limiter)

	for i := 0; i < 2; i++ {
		if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusOK {
			t.Fatalf("request %d should be allowed, got status %d", i+1, recorder.Code)
		}
	}

	recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if retryAfter := recorder.Header().Get("Retry-After"); retryAfter == "" {
		t.Error("expected Retry-After header on a throttled response")
	} else if retryAfter == "0" {
		t.Errorf("Retry-After must be at least 1 second, got %q", retryAfter)
	}
	if remaining := recorder.Header().Get("X-RateLimit-Remaining"); remaining != "0" {
		t.Errorf("got X-RateLimit-Remaining=%q, want %q", remaining, "0")
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body.Success {
		t.Error("throttled response must have success=false")
	}
	if body.Error.Code != "TOO_MANY_REQUESTS" {
		t.Errorf("got error code %q, want %q", body.Error.Code, "TOO_MANY_REQUESTS")
	}
}

func TestRateLimiterKeepsBucketsPerClientKey(t *testing.T) {
	limiter := middleware.NewRateLimiter(1, 1, time.Minute)
	router := newThrottledRouter(limiter)

	first := httptest.NewRequest(http.MethodGet, "/ping", nil)
	first.Header.Set("X-Forwarded-For", "10.0.0.1")
	if recorder := serve(t, router, first); recorder.Code != http.StatusOK {
		t.Fatalf("first client request: got status %d, want %d", recorder.Code, http.StatusOK)
	}

	second := httptest.NewRequest(http.MethodGet, "/ping", nil)
	second.Header.Set("X-Forwarded-For", "10.0.0.2")
	if recorder := serve(t, router, second); recorder.Code != http.StatusOK {
		t.Fatalf("second client must not be affected by the first client's quota, got status %d", recorder.Code)
	}

	exhausted := httptest.NewRequest(http.MethodGet, "/ping", nil)
	exhausted.Header.Set("X-Forwarded-For", "10.0.0.1")
	if recorder := serve(t, router, exhausted); recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("first client exceeded its quota: got status %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiterRefillsTokensOverTime(t *testing.T) {
	limiter := middleware.NewRateLimiter(2, 1, time.Minute) // 1 token, refilled at 2/second

	current := time.Now()
	limiter.SetClock(func() time.Time { return current })

	router := newThrottledRouter(limiter)

	if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusOK {
		t.Fatalf("first request: got status %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("bucket should be empty: got status %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}

	// Half a second at 2 tokens/second is exactly one token.
	current = current.Add(500 * time.Millisecond)

	if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusOK {
		t.Fatalf("bucket should have refilled: got status %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRateLimiterNeverExceedsBurstAfterLongIdlePeriod(t *testing.T) {
	limiter := middleware.NewRateLimiter(10, 1, time.Minute)

	current := time.Now()
	limiter.SetClock(func() time.Time { return current })
	router := newThrottledRouter(limiter)

	if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusOK {
		t.Fatalf("first request: got status %d, want %d", recorder.Code, http.StatusOK)
	}

	// A day of idle time must refill up to the bucket capacity (1), not beyond it.
	current = current.Add(24 * time.Hour)

	if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusOK {
		t.Fatalf("request after idle period: got status %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("bucket size must stay capped at the burst value: got status %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiterZeroRateDisablesThrottling(t *testing.T) {
	limiter := middleware.NewRateLimiter(0, 1, time.Minute)
	router := newThrottledRouter(limiter)

	for i := 1; i <= 25; i++ {
		if recorder := serve(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil)); recorder.Code != http.StatusOK {
			t.Fatalf("request %d: got status %d, want %d (zero rate must disable throttling)", i, recorder.Code, http.StatusOK)
		}
	}
}

func TestRateLimiterMiddlewareByKey(t *testing.T) {
	limiter := middleware.NewRateLimiter(1, 1, time.Minute)

	router := chi.NewRouter()
	router.Use(limiter.MiddlewareByKey(func(r *http.Request) string {
		return r.Header.Get("X-Company-ID")
	}))
	router.Get("/ping", okHandler())

	first := httptest.NewRequest(http.MethodGet, "/ping", nil)
	first.Header.Set("X-Company-ID", "company-a")
	if recorder := serve(t, router, first); recorder.Code != http.StatusOK {
		t.Fatalf("company A: got status %d, want %d", recorder.Code, http.StatusOK)
	}

	second := httptest.NewRequest(http.MethodGet, "/ping", nil)
	second.Header.Set("X-Company-ID", "company-b")
	if recorder := serve(t, router, second); recorder.Code != http.StatusOK {
		t.Fatalf("company B must have its own quota: got status %d", recorder.Code)
	}

	exhausted := httptest.NewRequest(http.MethodGet, "/ping", nil)
	exhausted.Header.Set("X-Company-ID", "company-a")
	if recorder := serve(t, router, exhausted); recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("company A exceeded its quota: got status %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiterRecoversQuotaAfterIdleTTL(t *testing.T) {
	limiter := middleware.NewRateLimiter(1, 1, time.Minute)

	current := time.Now()
	limiter.SetClock(func() time.Time { return current })
	router := newThrottledRouter(limiter)

	request := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/ping", nil)
		r.Header.Set("X-Forwarded-For", "10.1.1.1")
		return serve(t, router, r)
	}

	if recorder := request(); recorder.Code != http.StatusOK {
		t.Fatalf("first request: got status %d, want %d", recorder.Code, http.StatusOK)
	}

	// Past the idle TTL the bucket is garbage collected and the client starts
	// with a fresh quota instead of accumulating state forever.
	current = current.Add(2 * time.Minute)

	if recorder := request(); recorder.Code != http.StatusOK {
		t.Fatalf("request after idle TTL: got status %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "remote address only",
			remoteAddr: "192.0.2.10:54321",
			want:       "192.0.2.10",
		},
		{
			name:       "single x-forwarded-for",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.7"},
			want:       "203.0.113.7",
		},
		{
			name:       "x-forwarded-for chain uses the original client",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.7, 70.41.3.18"},
			want:       "203.0.113.7",
		},
		{
			name:       "x-real-ip fallback",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Real-IP": "198.51.100.23"},
			want:       "198.51.100.23",
		},
		{
			name:       "remote address without port",
			remoteAddr: "192.0.2.55",
			want:       "192.0.2.55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			request.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				request.Header.Set(key, value)
			}

			if got := middleware.ClientIP(request); got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}