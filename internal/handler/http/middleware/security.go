package middleware

import (
	"net/http"
	"strings"
)

// DefaultContentSecurityPolicy is tuned for this API's needs:
//   - the API itself only returns JSON, so 'self' is enough for scripts/styles;
//   - 'unsafe-inline' keeps the embedded Swagger UI working;
//   - img-src allows data:/blob: because attendance photos and avatars are
//     rendered from local storage (and the frontend uses object URLs);
//   - frame-ancestors 'none' blocks clickjacking.
const DefaultContentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; " +
	"font-src 'self' data:; " +
	"connect-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

// SecurityHeaders applies a baseline of hardening headers to every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return SecurityHeadersWithCSP(DefaultContentSecurityPolicy)(next)
}

// SecurityHeadersWithCSP is SecurityHeaders with an overridable CSP, for the rare
// case where an endpoint must embed third-party resources. An empty policy skips
// the Content-Security-Policy header entirely.
func SecurityHeadersWithCSP(policy string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()

			// Never let the browser sniff a different content type than declared.
			header.Set("X-Content-Type-Options", "nosniff")
			// No framing: defence in depth against clickjacking (see CSP too).
			header.Set("X-Frame-Options", "DENY")
			// Do not leak the API URL to third parties, keep the origin on same-site hops.
			header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			// 'cross-origin' is required: the Next.js frontend runs on another origin
			// and embeds avatars/logos served by /uploads.
			header.Set("Cross-Origin-Resource-Policy", "cross-origin")
			// Camera access is used for attendance proof photos; geolocation for clock-in.
			header.Set("Permissions-Policy", "geolocation=(self), camera=(self), microphone=()")

			if policy != "" {
				header.Set("Content-Security-Policy", policy)
			}

			// Only advertise HSTS on TLS connections: sending it over plain HTTP
			// would pin the browser to a scheme the server does not serve.
			if isHTTPS(r) {
				header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isHTTPS reports whether the request reached the app over TLS, taking reverese
// proxy termination into account.
func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}