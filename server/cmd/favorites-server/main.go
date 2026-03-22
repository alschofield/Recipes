package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"recipes/pkg/add"
	"recipes/pkg/delete"
	"recipes/pkg/middleware"
	"recipes/pkg/search"

	"golang.org/x/time/rate"
)

func main() {
	mux := http.NewServeMux()
	metrics := middleware.NewMetricsCollector("favorites-server")

	mux.HandleFunc("GET /favorites/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /favorites/metrics", metrics.Handler)

	mux.Handle("GET /favorites/{userid}", middleware.Chain(
		http.HandlerFunc(search.ListFavorites),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	mux.Handle("POST /favorites/{userid}/{recipeid}", middleware.Chain(
		http.HandlerFunc(add.AddFavorite),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	mux.Handle("DELETE /favorites/{userid}/{recipeid}", middleware.Chain(
		http.HandlerFunc(delete.RemoveFavorite),
		middleware.RequireAuth,
		middleware.RequireSelfOrAdmin("userid"),
	))

	handler := middleware.Chain(
		mux,
		middleware.Recoverer,
		middleware.RequestID,
		metrics.Middleware,
		middleware.ErrorNotifier(os.Getenv("ERROR_WEBHOOK_URL"), "favorites-server"),
		middleware.RequestLogger("favorites-server"),
		middleware.SecurityHeaders,
		middleware.CORS(middleware.ParseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))),
		middleware.BodyLimit(middleware.ParseMaxBodyBytes(os.Getenv("MAX_BODY_BYTES"), 1<<20)),
		middleware.RateLimit(rate.Limit(parseRate(os.Getenv("RATE_LIMIT_RPS"), 20)), parseBurst(os.Getenv("RATE_LIMIT_BURST"), 40)),
	)

	port := os.Getenv("FAVORITES_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting favorites server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Println("Error starting server:", err)
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
