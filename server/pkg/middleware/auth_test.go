package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"recipes/pkg/auth"
)

func TestRequireSelfOrAdminAuthorizationMatrix(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-1234567890")
	t.Setenv("JWT_ISSUER", "recipes-test")

	secured := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequireAuth,
		RequireSelfOrAdmin("userid"),
	)

	tests := []struct {
		name         string
		tokenUserID  string
		tokenRole    string
		targetUserID string
		wantStatus   int
	}{
		{
			name:         "user can access own resource",
			tokenUserID:  "user-1",
			tokenRole:    "user",
			targetUserID: "user-1",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "user cannot access another user",
			tokenUserID:  "user-1",
			tokenRole:    "user",
			targetUserID: "user-2",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "admin can access another user",
			tokenUserID:  "admin-1",
			tokenRole:    "admin",
			targetUserID: "user-2",
			wantStatus:   http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, _, err := auth.GenerateAccessToken(tc.tokenUserID, tc.tokenRole)
			if err != nil {
				t.Fatalf("failed to create token: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/favorites/"+tc.targetUserID, nil)
			req.SetPathValue("userid", tc.targetUserID)
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			secured.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	secured := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequireAuth,
	)

	req := httptest.NewRequest(http.MethodGet, "/favorites/user-1", nil)
	rr := httptest.NewRecorder()
	secured.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
