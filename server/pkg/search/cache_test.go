package search

import (
	"os"
	"testing"
)

func TestBuildSearchCacheKeyIgnoresIngredientOrder(t *testing.T) {
	req := SearchRequest{Mode: "strict", DBOnly: false}
	a := buildSearchCacheKey([]string{"garlic", "rice", "chicken"}, req, 1, 20)
	b := buildSearchCacheKey([]string{"chicken", "garlic", "rice"}, req, 1, 20)

	if a != b {
		t.Fatalf("expected same cache key for same ingredient set, got %s and %s", a, b)
	}
}

func TestShouldBypassSearchCache(t *testing.T) {
	original := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", original)

	os.Setenv("APP_ENV", "production")
	if shouldBypassSearchCache(true) {
		t.Fatal("expected debugNoCache to be ignored in production")
	}

	os.Setenv("APP_ENV", "development")
	if !shouldBypassSearchCache(true) {
		t.Fatal("expected debugNoCache to bypass cache in non-production")
	}
}
