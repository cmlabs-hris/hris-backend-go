package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
)

// DefaultRequestTimeout is the deadline applied to regular API requests.
const DefaultRequestTimeout = 30 * time.Second

// RequestTimeout cancels handlers that exceed the deadline and answers 504.
// Long-lived streams (SSE) are skipped by default, see SkipStreamingRequests.
func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return RequestTimeoutWithSkip(timeout, SkipStreamingRequests)
}

// RequestTimeoutWithSkip is RequestTimeout with a custom skip predicate, e.g. to
// exclude an export endpoint that legitimately runs for minutes.
func RequestTimeoutWithSkip(timeout time.Duration, skip func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// A non-positive timeout disables the middleware (handy for local dev).
			if timeout <= 0 || (skip != nil && skip(r)) {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			tw := &timeoutWriter{ResponseWriter: w}
			done := make(chan struct{})
			panicked := make(chan any, 1)

			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicked <- p
						return
					}
					close(done)
				}()

				next.ServeHTTP(tw, r.WithContext(ctx))
			}()

			select {
			case p := <-panicked:
				// Re-panic in the original goroutine so chi's Recoverer (or the
				// test) sees it exactly as it would without this middleware.
				panic(p)
			case <-done:
			case <-ctx.Done():
				tw.timedOutResponse(w, timeout)
			}
		})
	}
}

// SkipStreamingRequests reports whether the request is a long-lived stream that
// must not be cut off by the timeout middleware: Server-Sent Events
// (Accept: text/event-stream, or any route ending in /stream, like
// /api/v1/notifications/stream).
func SkipStreamingRequests(r *http.Request) bool {
	if strings.Contains(strings.ToLower(r.Header.Get("Accept")), "text/event-stream") {
		return true
	}
	return strings.HasSuffix(r.URL.Path, "/stream")
}

// timeoutWriter shields the real ResponseWriter so a handler abandoned after the
// deadline can no longer write to the client.
//
// The mutex is held for the whole delegate call (not just to flip the flags):
// that way the "write the 504" path and the handler's writes are strictly
// ordered, and a response that was already committed before the deadline can
// never be overwritten by the timeout error.
type timeoutWriter struct {
	http.ResponseWriter

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
}

func (tw *timeoutWriter) WriteHeader(statusCode int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.wroteHeader = true

	tw.ResponseWriter.WriteHeader(statusCode)
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		// Mirror the standard library behaviour for abandoned handlers.
		return 0, http.ErrHandlerTimeout
	}
	tw.wroteHeader = true

	return tw.ResponseWriter.Write(b)
}

// Flush keeps streaming-friendly handlers working (best effort passthrough).
func (tw *timeoutWriter) Flush() {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return
	}
	if flusher, ok := tw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// timedOutResponse marks the request as timed out and writes the 504 envelope,
// unless the handler already committed a response - that status line is on the
// wire, so the timeout is logged by the caller instead of mutating the response.
func (tw *timeoutWriter) timedOutResponse(w http.ResponseWriter, timeout time.Duration) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	tw.timedOut = true
	if tw.wroteHeader {
		return
	}

	response.GatewayTimeout(w, fmt.Sprintf("Request exceeded the %s timeout", timeout))
}