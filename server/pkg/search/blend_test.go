package search

import "testing"

func TestBlendSearchResultsDeterministic(t *testing.T) {
	t.Setenv("SEARCH_BLEND_SEED", "stable-seed")
	t.Setenv("SEARCH_BLEND_MIN_GENERATED", "1")
	t.Setenv("SEARCH_BLEND_MAX_GENERATED_SHARE", "0.5")

	req := SearchRequest{Mode: "inclusive"}
	normalized := []string{"chicken", "rice"}
	dbResults := []SearchResultItem{
		{ID: "d1", Source: "database"},
		{ID: "d2", Source: "database"},
		{ID: "d3", Source: "database"},
		{ID: "d4", Source: "database"},
	}
	generated := []SearchResultItem{
		{ID: "g1", Source: "llm"},
		{ID: "g2", Source: "llm"},
		{ID: "g3", Source: "llm"},
	}

	first := blendSearchResults(req, normalized, append([]SearchResultItem(nil), dbResults...), append([]SearchResultItem(nil), generated...))
	second := blendSearchResults(req, normalized, append([]SearchResultItem(nil), dbResults...), append([]SearchResultItem(nil), generated...))

	if len(first) != len(second) {
		t.Fatalf("expected same result size, got %d vs %d", len(first), len(second))
	}

	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("position %d mismatch: %s vs %s", i, first[i].ID, second[i].ID)
		}
	}
}

func TestBlendSearchResultsRespectsGeneratedShareLimit(t *testing.T) {
	t.Setenv("SEARCH_BLEND_MIN_GENERATED", "1")
	t.Setenv("SEARCH_BLEND_MAX_GENERATED_SHARE", "0.4")

	req := SearchRequest{Mode: "inclusive"}
	normalized := []string{"a", "b"}
	dbResults := []SearchResultItem{
		{ID: "d1", Source: "database"},
		{ID: "d2", Source: "database"},
		{ID: "d3", Source: "database"},
		{ID: "d4", Source: "database"},
		{ID: "d5", Source: "database"},
	}
	generated := []SearchResultItem{
		{ID: "g1", Source: "llm"},
		{ID: "g2", Source: "llm"},
		{ID: "g3", Source: "llm"},
		{ID: "g4", Source: "llm"},
	}

	blended := blendSearchResults(req, normalized, dbResults, generated)

	generatedCount := 0
	for _, item := range blended {
		if item.Source == "llm" {
			generatedCount++
		}
	}

	if generatedCount != 3 {
		t.Fatalf("expected 3 generated results after share cap, got %d", generatedCount)
	}
}

func TestAnnotateBlendMetadataSetsBlendSlotAndReason(t *testing.T) {
	items := []SearchResultItem{
		{ID: "d1", Source: "database"},
		{ID: "g1", Source: "llm"},
	}

	annotateBlendMetadata(items)

	if items[0].BlendSlot != 1 || items[1].BlendSlot != 2 {
		t.Fatalf("unexpected blend slots: %d, %d", items[0].BlendSlot, items[1].BlendSlot)
	}
	if items[0].RankingReason == "" || items[1].RankingReason == "" {
		t.Fatal("expected ranking reasons to be populated")
	}
}

func TestBuildSearchBlendSeedKeyIgnoresPagination(t *testing.T) {
	normalized := []string{"a", "b"}
	reqA := SearchRequest{Mode: "inclusive", Pagination: &PaginationInput{Page: 1, PageSize: 10}}
	reqB := SearchRequest{Mode: "inclusive", Pagination: &PaginationInput{Page: 2, PageSize: 20}}

	keyA := buildSearchBlendSeedKey(normalized, reqA)
	keyB := buildSearchBlendSeedKey(normalized, reqB)

	if keyA != keyB {
		t.Fatal("expected blend seed key to ignore pagination")
	}
}
