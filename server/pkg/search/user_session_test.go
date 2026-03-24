package search

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHashRefreshTokenDeterministic(t *testing.T) {
	one := hashRefreshToken("token-abc")
	two := hashRefreshToken("token-abc")
	if one != two {
		t.Fatal("expected hashRefreshToken to be deterministic")
	}
	if one == "token-abc" {
		t.Fatal("expected token hash to differ from raw token")
	}
}

func TestRequestIPPrefersForwardedHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.15, 10.0.0.2")
	req.RemoteAddr = "127.0.0.1:12345"

	if got := requestIP(req); got != "203.0.113.15" {
		t.Fatalf("expected forwarded ip, got %q", got)
	}
}

func TestRequestIPFallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "198.51.100.22:4567"

	if got := requestIP(req); got != "198.51.100.22" {
		t.Fatalf("expected remote host ip, got %q", got)
	}
}

func TestClientSessionIDUsesHeaderWhenPresent(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Client-Session-ID", "ios-device-abc")

	if got := clientSessionID(req); got != "ios-device-abc" {
		t.Fatalf("expected header value, got %q", got)
	}
}

func TestClientSessionIDGeneratesWhenMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	got := clientSessionID(req)
	if strings.TrimSpace(got) == "" {
		t.Fatal("expected generated session id")
	}
}
