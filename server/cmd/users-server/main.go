package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"recipes/pkg/create"
	"recipes/pkg/delete"
	"recipes/pkg/edit"
	"recipes/pkg/middleware"
	"recipes/pkg/search"

	"golang.org/x/time/rate"
)

func main() {
	mux := http.NewServeMux()
	metrics := middleware.NewMetricsCollector("users-server")

	// Health check
	mux.HandleFunc("GET /users/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /users/metrics", metrics.Handler)

	// List all users (admin only)
	mux.Handle("GET /users", middleware.Chain(
		http.HandlerFunc(search.ListUsers),
		middleware.RequireAuth,
		middleware.RequireRole("admin"),
	))

	// Create new user (signup)
	mux.Handle("POST /users/new", middleware.Chain(
		http.HandlerFunc(create.HandleSignup),
		middleware.IdempotencyKey(middleware.ParseIdempotencyTTL(os.Getenv("IDEMPOTENCY_KEY_TTL"), 24*time.Hour)),
		middleware.RateLimit(rate.Limit(2), 5),
	))

	// Login
	mux.Handle("POST /users/login", middleware.Chain(
		http.HandlerFunc(search.HandleLogin),
		middleware.RateLimit(rate.Limit(3), 8),
	))

	// Refresh access token
	mux.Handle("POST /users/refresh", middleware.Chain(
		http.HandlerFunc(search.HandleRefresh),
		middleware.RateLimit(rate.Limit(4), 10),
	))

	// Logout session family
	mux.Handle("POST /users/logout", middleware.Chain(
		http.HandlerFunc(search.HandleLogout),
		middleware.RateLimit(rate.Limit(4), 10),
	))

	// Logout current client session only
	mux.Handle("POST /users/logout/session", middleware.Chain(
		http.HandlerFunc(search.HandleLogoutSession),
		middleware.RateLimit(rate.Limit(4), 10),
	))

	// Get user profile
	mux.Handle("GET /users/{userid}", middleware.Chain(
		http.HandlerFunc(search.GetProfile),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	// List active sessions for user
	mux.Handle("GET /users/{userid}/sessions", middleware.Chain(
		http.HandlerFunc(search.ListSessions),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	// Update user profile
	mux.Handle("PUT /users/{userid}", middleware.Chain(
		http.HandlerFunc(edit.HandleUpdateProfile),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	// Delete user account
	mux.Handle("DELETE /users/{userid}", middleware.Chain(
		http.HandlerFunc(delete.HandleDeleteUser),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	handler := middleware.Chain(
		mux,
		middleware.Recoverer,
		middleware.RequestID,
		metrics.Middleware,
		middleware.ErrorNotifier(os.Getenv("ERROR_WEBHOOK_URL"), "users-server"),
		middleware.RequestLogger("users-server"),
		middleware.SecurityHeaders,
		middleware.CORS(middleware.ParseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))),
		middleware.BodyLimit(middleware.ParseMaxBodyBytes(os.Getenv("MAX_BODY_BYTES"), 1<<20)),
		middleware.RateLimit(rate.Limit(parseRate(os.Getenv("RATE_LIMIT_RPS"), 20)), parseBurst(os.Getenv("RATE_LIMIT_BURST"), 40)),
	)

	port := os.Getenv("USERS_SERVER_PORT")
	if port == "" {
		port = "8082"
	}

	fmt.Printf("Starting users server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Println("Error starting users server:", err)
	}
}

func parseRate(raw string, fallback int) float64 {
	if raw == "" {
		return float64(fallback)
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 {
		return float64(fallback)
	}

	return v
}

func parseBurst(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}

	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}

	return v
}
