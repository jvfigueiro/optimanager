package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"optimanager/internal/handlers"
)

func TestExtractClientIP_UntrustedProxyIgnoresHeaders(t *testing.T) {
	t.Setenv("TRUST_PROXY", "false")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.50:54321"
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 10.0.0.1")
	req.Header.Set("X-Real-IP", "198.51.100.1")

	ip := handlers.ExtractClientIP(req)
	if ip != "192.168.1.50" {
		t.Fatalf("expected remote host 192.168.1.50 when TRUST_PROXY is false, got %q", ip)
	}

	// Also test when TRUST_PROXY is empty/unset
	t.Setenv("TRUST_PROXY", "")
	ipEmpty := handlers.ExtractClientIP(req)
	if ipEmpty != "192.168.1.50" {
		t.Fatalf("expected remote host 192.168.1.50 when TRUST_PROXY is empty, got %q", ipEmpty)
	}
}

func TestExtractClientIP_TrustedProxyUsesHeaders(t *testing.T) {
	t.Setenv("TRUST_PROXY", "true")

	// 1. X-Forwarded-For first client hop
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "127.0.0.1:8080"
	req1.Header.Set("X-Forwarded-For", "203.0.113.195, 10.0.0.1, 192.168.1.1")

	ip1 := handlers.ExtractClientIP(req1)
	if ip1 != "203.0.113.195" {
		t.Fatalf("expected first hop 203.0.113.195 from X-Forwarded-For, got %q", ip1)
	}

	// 2. X-Real-IP when X-Forwarded-For is absent
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "127.0.0.1:8080"
	req2.Header.Set("X-Real-IP", "198.51.100.42")

	ip2 := handlers.ExtractClientIP(req2)
	if ip2 != "198.51.100.42" {
		t.Fatalf("expected 198.51.100.42 from X-Real-IP, got %q", ip2)
	}

	// 3. Fallback to RemoteAddr if no proxy headers are present
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "172.18.0.5:43210"

	ip3 := handlers.ExtractClientIP(req3)
	if ip3 != "172.18.0.5" {
		t.Fatalf("expected fallback host 172.18.0.5, got %q", ip3)
	}
}

