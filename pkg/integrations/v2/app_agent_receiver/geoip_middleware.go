package app_agent_receiver

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var (
	// De-facto standard header keys.
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

var (
	// RFC7239 defines a new "Forwarded: " header designed to replace the
	// existing use of X-Forwarded-* headers.
	// e.g. Forwarded: for=192.0.2.60;proto=https;by=203.0.113.43
	forwarded = http.CanonicalHeaderKey("Forwarded")
	// Allows for a sub-match of the first value after 'for=' to the next
	// comma, semi-colon or space. The match is case-insensitive.
	forRegex = regexp.MustCompile(`(?i)(?:for=)([^(;|,| )]+)`)
)

type requestContextKeyType string

const clientIPContextKey requestContextKeyType = "client_ip"

// AddClientIPToContext adds the client IP to the request context. Retrieves client IP from the
// RFC7239 Forwarded headers, X-Real-IP and X-Forwarded-For (in that order)
// If headers not present, returns the request.RemoteAddr. Note that this is not a reliable way to
// get the client IP, as it can be spoofed.
func addClientIPToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request.RemoteAddr for the client IP
		// ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		// Check for client IP in the request headers
		ip := getIPFromHeaders(r)

		// If that failed check the request.RemoteAddr
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr) // TODO: handle error
		}

		ctx := setClientIPInContext(r.Context(), ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetClientIP sets the client IP in the context.
func setClientIPInContext(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPContextKey, ip)
}

// GetClientIP gets the client IP from the context.
func getClientIPFromContext(ctx context.Context) string {
	return ctx.Value(clientIPContextKey).(string)
}

// getIPFromHeaders retrieves the IP from the RFC7239 Forwarded headers,
// X-Real-IP and X-Forwarded-For (in that order)
func getIPFromHeaders(r *http.Request) string {
	var addr string

	if fwd := r.Header.Get(forwarded); fwd != "" {
		// match should contain at least two elements if the protocol was
		// specified in the Forwarded header. The first element will always be
		// the 'for=' capture, which we ignore. In the case of multiple IP
		// addresses (for=8.8.8.8, 8.8.4.4,172.16.1.20 is valid) we only
		// extract the first, which should be the client IP.
		if match := forRegex.FindStringSubmatch(fwd); len(match) > 1 {
			// IPv6 addresses in Forwarded headers are quoted-strings. We strip
			// these quotes.
			addr = strings.Trim(match[1], `"`)
		}
	} else if fwd := r.Header.Get(xRealIP); fwd != "" {
		// X-Real-IP should only contain one IP address (the client making the
		// request).
		addr = fwd
	} else if fwd := strings.ReplaceAll(r.Header.Get(xForwardedFor), " ", ""); fwd != "" {
		// Only grab the first (client) address. Note that '192.168.0.1,
		// 10.1.1.1' is a valid key for X-Forwarded-For where addresses after
		// the first may represent forwarding proxies earlier in the chain.
		s := strings.Index(fwd, ",")
		if s == -1 {
			s = len(fwd)
		}
		addr = fwd[:s]
	}

	return addr
}
