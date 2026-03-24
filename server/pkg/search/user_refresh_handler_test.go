package search

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRefreshRequiresToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users/refresh", strings.NewReader(`{"refreshToken":""}`))
	rr := httptest.NewRecorder()

	HandleRefresh(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleRefreshRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123456789")
	t.Setenv("JWT_ISSUER", "recipes-test")

	req := httptest.NewRequest(http.MethodPost, "/users/refresh", strings.NewReader(`{"refreshToken":"not-a-token"}`))
	rr := httptest.NewRecorder()

	HandleRefresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestHandleLogoutRequiresToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users/logout", strings.NewReader(`{"refreshToken":""}`))
	rr := httptest.NewRecorder()

	HandleLogout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleLogoutRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123456789")
	t.Setenv("JWT_ISSUER", "recipes-test")

	req := httptest.NewRequest(http.MethodPost, "/users/logout", strings.NewReader(`{"refreshToken":"not-a-token"}`))
	rr := httptest.NewRecorder()

	HandleLogout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestHandleLogoutSessionRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123456789")
	t.Setenv("JWT_ISSUER", "recipes-test")

	req := httptest.NewRequest(http.MethodPost, "/users/logout/session", strings.NewReader(`{"refreshToken":"not-a-token"}`))
	rr := httptest.NewRecorder()

	HandleLogoutSession(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}
