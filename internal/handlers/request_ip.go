package handlers

import (
	"net"
	"net/http"
	"os"
	"strings"
)

// ExtractClientIP retrieves the client's IP address.
// If TRUST_PROXY is set to "true" (case-insensitive), it extracts the client IP
// from X-Forwarded-For (first hop) or X-Real-IP headers.
// Otherwise, it strictly extracts the host portion from r.RemoteAddr, preventing header spoofing.
func ExtractClientIP(r *http.Request) string {
	trustProxy := strings.ToLower(strings.TrimSpace(os.Getenv("TRUST_PROXY"))) == "true"
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				ip := strings.TrimSpace(parts[0])
				if ip != "" {
					return ip
				}
			}
		}
		if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
			return xrip
		}
	}

	// Default or untrusted proxy: extract host from RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

