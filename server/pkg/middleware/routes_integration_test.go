package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"recipes/pkg/auth"
)

func TestRouteProtectionMatrix(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-1234567890")
	t.Setenv("JWT_ISSUER", "recipes-test")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /recipes/search", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /recipes/detail/{recipeid}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.Handle("GET /favorites/{userid}", Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		RequireAuth,
		RequireSelfOrAdmin("userid"),
	))

	server := httptest.NewServer(mux)
	defer server.Close()

	userToken, _, err := auth.GenerateAccessToken("user-1", "user")
	if err != nil {
		t.Fatalf("failed to generate user token: %v", err)
	}
	adminToken, _, err := auth.GenerateAccessToken("admin-1", "admin")
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}

	cases := []struct {
		name   string
		method string
		path   string
		token  string
		want   int
	}{
		{"public search route", http.MethodPost, "/recipes/search", "", http.StatusOK},
		{"public detail route", http.MethodGet, "/recipes/detail/abc", "", http.StatusOK},
		{"favorites require auth", http.MethodGet, "/favorites/user-1", "", http.StatusUnauthorized},
		{"favorites own user ok", http.MethodGet, "/favorites/user-1", userToken, http.StatusOK},
		{"favorites blocked cross-user", http.MethodGet, "/favorites/user-2", userToken, http.StatusForbidden},
		{"favorites admin allowed", http.MethodGet, "/favorites/user-2", adminToken, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("expected status %d, got %d", tc.want, res.StatusCode)
			}
		})
	}
}
