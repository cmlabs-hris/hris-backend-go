package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

// RateLimiter is an in-memory, per-key token bucket rate limiter.
//
// Design notes:
//   - A bucket starts full (burst tokens) and refills at `rate` tokens per second.
//   - Buckets are keyed by an arbitrary string (client IP by default), so the same
//     middleware can throttle per IP, per company or per API key.
//   - Stale buckets are evicted lazily on write, so no background goroutine is
//     required and no state leaks between tests.
//   - This limiter is process-local. When the API runs on multiple replicas the
//     counters should move to a shared store (Redis) so the effective limit stays
//     the same across instances.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket

	rate    float64 // tokens added per second
	burst   float64 // bucket capacity
	ttl     time.Duration
	sweepAt time.Time

	// now is injectable so tests can control the clock deterministically.
	now func() time.Time
}

type tokenBucket struct {
	tokens float64
	last   time.Time
}

// NewRateLimiter builds a limiter that allows `burst` requests immediately and
// then `requestsPerSecond` requests per second per key.
// A rate of 0 disables throttling (useful for local development and tests).
func NewRateLimiter(requestsPerSecond float64, burst int, ttl time.Duration) *RateLimiter {
	if requestsPerSecond < 0 {
		requestsPerSecond = 0
	}
	if burst < 1 {
		burst = 1
	}
	if ttl <= 0 {
		ttl = time.Minute
	}

	now := time.Now()
	return &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    requestsPerSecond,
		burst:   float64(burst),
		ttl:     ttl,
		sweepAt: now.Add(ttl),
		now:     time.Now,
	}
}

// SetClock overrides the limiter clock. It must be called before the limiter is
// used to serve traffic (it exists mainly for deterministic tests).
func (l *RateLimiter) SetClock(now func() time.Time) *RateLimiter {
	if now == nil {
		return l
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.now = now
	return l
}

// Middleware throttles requests using the client IP as bucket key.
func (l *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return l.MiddlewareByKey(ClientIP)
}

// MiddlewareByKey throttles requests using a custom bucket key, e.g. the
// authenticated user or company extracted from the JWT claims.
func (l *RateLimiter) MiddlewareByKey(keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)

			allowed, retryAfter, remaining := l.reserve(key)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(int(l.burst)))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
				response.TooManyRequests(w, "Too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// reserve consumes one token for key and reports whether the request is allowed.
// When it is not, the returned duration tells the caller how long to wait.
func (l *RateLimiter) reserve(key string) (allowed bool, retryAfter time.Duration, remaining int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.evictStale(now)

	if l.rate == 0 {
		return true, 0, int(l.burst)
	}

	bucket, ok := l.buckets[key]
	if !ok {
		bucket = &tokenBucket{tokens: l.burst, last: now}
		l.buckets[key] = bucket
	}

	if elapsed := now.Sub(bucket.last).Seconds(); elapsed > 0 {
		bucket.tokens = math.Min(l.burst, bucket.tokens+elapsed*l.rate)
	}
	bucket.last = now

	if bucket.tokens < 1 {
		missing := 1 - bucket.tokens
		return false, time.Duration(math.Ceil(missing / l.rate * float64(time.Second))), 0
	}

	bucket.tokens--
	return true, 0, int(bucket.tokens)
}

// evictStale drops buckets that have been idle for longer than the TTL. Sweeping
// is throttled to once per TTL to keep the hot path cheap.
func (l *RateLimiter) evictStale(now time.Time) {
	if now.Before(l.sweepAt) {
		return
	}
	l.sweepAt = now.Add(l.ttl)

	for key, bucket := range l.buckets {
		if now.Sub(bucket.last) > l.ttl {
			delete(l.buckets, key)
		}
	}
}

// ClientIP resolves the client address, honouring the proxy headers used by the
// deployment (X-Forwarded-For, X-Real-IP) and falling back to RemoteAddr.
//
// NOTE: these headers are client-controlled, so they are only trustworthy behind
// a reverse proxy that overwrites them. Deployments that terminate traffic
// directly should enable chi's RealIP middleware with a trusted proxy list.
func ClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
			return first
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}