package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"recipes/pkg/ingredients"
	"recipes/pkg/middleware"
	"recipes/pkg/search"

	"golang.org/x/time/rate"
)

func main() {
	mux := http.NewServeMux()
	metrics := middleware.NewMetricsCollector("recipes-server")

	mux.HandleFunc("GET /recipes", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("recipes root response"))
	})

	mux.HandleFunc("GET /recipes/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /recipes/metrics", metrics.Handler)

	mux.HandleFunc("POST /recipes/search", search.HandleSearch)
	mux.HandleFunc("GET /recipes/catalog", search.GetRecipeCatalog)
	mux.HandleFunc("GET /recipes/detail/{recipeid}", search.GetRecipeDetail)
	mux.Handle("GET /recipes/analysis/admin", middleware.Chain(
		http.HandlerFunc(search.GetAnalysisAdminDashboard),
		middleware.RequireAuth,
		middleware.RequireRole("admin"),
	))

	mux.Handle("POST /ingredients/suggestions", middleware.Chain(
		http.HandlerFunc(ingredients.SuggestIngredient),
		middleware.RequireAuth,
	))
	mux.Handle("GET /ingredients/candidates", middleware.Chain(
		http.HandlerFunc(ingredients.ListCandidates),
		middleware.RequireAuth,
		middleware.RequireRole("admin"),
	))
	mux.Handle("POST /ingredients/candidates/{candidateid}/resolve", middleware.Chain(
		http.HandlerFunc(ingredients.ResolveCandidate),
		middleware.RequireAuth,
		middleware.RequireRole("admin"),
	))
	mux.Handle("POST /ingredients/candidates/{candidateid}/votes", middleware.Chain(
		http.HandlerFunc(ingredients.VoteCandidate),
		middleware.RequireAuth,
	))
	mux.Handle("GET /ingredients/metrics", middleware.Chain(
		http.HandlerFunc(ingredients.IngredientMetrics),
		middleware.RequireAuth,
		middleware.RequireRole("admin"),
	))
	mux.HandleFunc("GET /ingredients/catalog", ingredients.ListIngredientCatalog)
	mux.HandleFunc("GET /ingredients/detail/{ingredientid}", ingredients.GetIngredientDetail)

	handler := middleware.Chain(
		mux,
		middleware.Recoverer,
		middleware.RequestID,
		metrics.Middleware,
		middleware.ErrorNotifier(os.Getenv("ERROR_WEBHOOK_URL"), "recipes-server"),
		middleware.RequestLogger("recipes-server"),
		middleware.SecurityHeaders,
		middleware.CORS(middleware.ParseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))),
		middleware.BodyLimit(middleware.ParseMaxBodyBytes(os.Getenv("MAX_BODY_BYTES"), 1<<20)),
		middleware.RateLimit(rate.Limit(parseRate(os.Getenv("RATE_LIMIT_RPS"), 30)), parseBurst(os.Getenv("RATE_LIMIT_BURST"), 60)),
	)

	port := os.Getenv("RECIPES_SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("Starting recipes server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Println("Error starting recipes server:", err)
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
