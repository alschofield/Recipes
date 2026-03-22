package search

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultSearchCacheTTL = 6 * time.Hour

type searchCacheEntry struct {
	response  SearchResponse
	expiresAt time.Time
}

var globalSearchCache = struct {
	mu      sync.RWMutex
	entries map[string]searchCacheEntry
}{
	entries: map[string]searchCacheEntry{},
}

func shouldBypassSearchCache(debugNoCache bool) bool {
	if !debugNoCache {
		return false
	}

	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	}

	return env != "production"
}

func getSearchCache(key string) (SearchResponse, bool) {
	now := time.Now().UTC()

	globalSearchCache.mu.RLock()
	entry, ok := globalSearchCache.entries[key]
	globalSearchCache.mu.RUnlock()
	if !ok {
		return SearchResponse{}, false
	}

	if now.After(entry.expiresAt) {
		globalSearchCache.mu.Lock()
		delete(globalSearchCache.entries, key)
		globalSearchCache.mu.Unlock()
		return SearchResponse{}, false
	}

	return cloneSearchResponse(entry.response), true
}

func setSearchCache(key string, response SearchResponse) {
	globalSearchCache.mu.Lock()
	globalSearchCache.entries[key] = searchCacheEntry{
		response:  cloneSearchResponse(response),
		expiresAt: time.Now().UTC().Add(searchCacheTTL()),
	}
	globalSearchCache.mu.Unlock()
}

func buildSearchCacheKey(normalized []string, req SearchRequest, page, pageSize int) string {
	ingredients := append([]string(nil), normalized...)
	sort.Strings(ingredients)

	payload := struct {
		Ingredients []string       `json:"ingredients"`
		Mode        string         `json:"mode"`
		DBOnly      bool           `json:"dbOnly"`
		Filters     *SearchFilters `json:"filters,omitempty"`
		Page        int            `json:"page"`
		PageSize    int            `json:"pageSize"`
	}{
		Ingredients: ingredients,
		Mode:        req.Mode,
		DBOnly:      req.DBOnly,
		Filters:     normalizeFiltersForCache(req.Filters),
		Page:        page,
		PageSize:    pageSize,
	}

	raw, _ := json.Marshal(payload)
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}

func normalizeFiltersForCache(filters *SearchFilters) *SearchFilters {
	if filters == nil {
		return nil
	}

	out := *filters
	out.Cuisine = append([]string(nil), filters.Cuisine...)
	out.Dietary = append([]string(nil), filters.Dietary...)
	out.Difficulty = append([]string(nil), filters.Difficulty...)
	sort.Strings(out.Cuisine)
	sort.Strings(out.Dietary)
	sort.Strings(out.Difficulty)
	return &out
}

func searchCacheTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("SEARCH_CACHE_TTL"))
	if raw == "" {
		return defaultSearchCacheTTL
	}

	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl <= 0 {
		return defaultSearchCacheTTL
	}

	return ttl
}

func cloneSearchResponse(in SearchResponse) SearchResponse {
	out := in
	out.Query.Ingredients = append([]string(nil), in.Query.Ingredients...)
	out.Results = make([]SearchResultItem, len(in.Results))
	for i := range in.Results {
		out.Results[i] = cloneSearchResultItem(in.Results[i])
	}
	return out
}

func cloneSearchResultItem(in SearchResultItem) SearchResultItem {
	out := in
	out.MatchedIngredients = append([]string(nil), in.MatchedIngredients...)
	out.MissingIngredients = append([]string(nil), in.MissingIngredients...)
	out.DietaryTags = append([]string(nil), in.DietaryTags...)
	out.Substitutions = make([]Substitution, len(in.Substitutions))
	for i := range in.Substitutions {
		out.Substitutions[i] = Substitution{
			Missing:     in.Substitutions[i].Missing,
			Substitutes: append([]string(nil), in.Substitutions[i].Substitutes...),
		}
	}
	return out
}
