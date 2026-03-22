package search

import (
	"sort"
	"testing"
	"time"
)

func TestCompareSearchResultsOrdering(t *testing.T) {
	now := time.Now().UTC()
	older := now.Add(-time.Hour)

	prep20 := 20
	prep30 := 30

	tests := []struct {
		name     string
		items    []SearchResultItem
		groups   map[string]*recipeGroup
		expected []string
	}{
		{
			name: "higher match percent wins",
			items: []SearchResultItem{
				{ID: "a", MatchPercent: 0.7, MissingIngredients: []string{"salt"}},
				{ID: "b", MatchPercent: 0.9, MissingIngredients: []string{"salt"}},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.5, UpdatedAt: now},
				"b": {QualityScore: 0.5, UpdatedAt: now},
			},
			expected: []string{"b", "a"},
		},
		{
			name: "fewer missing ingredients wins",
			items: []SearchResultItem{
				{ID: "a", MatchPercent: 0.8, MissingIngredients: []string{"salt", "pepper"}},
				{ID: "b", MatchPercent: 0.8, MissingIngredients: []string{"salt"}},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.5, UpdatedAt: now},
				"b": {QualityScore: 0.5, UpdatedAt: now},
			},
			expected: []string{"b", "a"},
		},
		{
			name: "higher quality wins after match ties",
			items: []SearchResultItem{
				{ID: "a", MatchPercent: 0.8, MissingIngredients: []string{"salt"}},
				{ID: "b", MatchPercent: 0.8, MissingIngredients: []string{"salt"}},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.6, UpdatedAt: now},
				"b": {QualityScore: 0.9, UpdatedAt: now},
			},
			expected: []string{"b", "a"},
		},
		{
			name: "known prep beats unknown prep",
			items: []SearchResultItem{
				{ID: "a", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: nil},
				{ID: "b", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep30},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.9, UpdatedAt: now},
				"b": {QualityScore: 0.9, UpdatedAt: older},
			},
			expected: []string{"b", "a"},
		},
		{
			name: "lower prep wins when both known",
			items: []SearchResultItem{
				{ID: "a", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep30},
				{ID: "b", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep20},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.9, UpdatedAt: now},
				"b": {QualityScore: 0.9, UpdatedAt: older},
			},
			expected: []string{"b", "a"},
		},
		{
			name: "newer updated_at wins after prep ties",
			items: []SearchResultItem{
				{ID: "a", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep20},
				{ID: "b", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep20},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.9, UpdatedAt: older},
				"b": {QualityScore: 0.9, UpdatedAt: now},
			},
			expected: []string{"b", "a"},
		},
		{
			name: "stable deterministic fallback uses id",
			items: []SearchResultItem{
				{ID: "b", Name: "Recipe B", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep20},
				{ID: "a", Name: "Recipe A", MatchPercent: 0.8, MissingIngredients: []string{"salt"}, PrepMinutes: &prep20},
			},
			groups: map[string]*recipeGroup{
				"a": {QualityScore: 0.9, UpdatedAt: now},
				"b": {QualityScore: 0.9, UpdatedAt: now},
			},
			expected: []string{"a", "b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sort.Slice(tc.items, func(i, j int) bool {
				return compareSearchResults(tc.items[i], tc.items[j], tc.groups)
			})

			for i, wantID := range tc.expected {
				if tc.items[i].ID != wantID {
					t.Fatalf("position %d: got %s want %s", i, tc.items[i].ID, wantID)
				}
			}
		})
	}
}

func TestNormalizeRawIngredient(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "  Tomatoes  ", want: "tomato"},
		{input: "Green Onions", want: "green onion"},
		{input: "berries", want: "berry"},
		{input: "glass", want: "glass"},
		{input: "", want: ""},
	}

	for _, tc := range tests {
		got := normalizeRawIngredient(tc.input)
		if got != tc.want {
			t.Fatalf("normalizeRawIngredient(%q): got %q want %q", tc.input, got, tc.want)
		}
	}
}

func TestPassesMode(t *testing.T) {
	tests := []struct {
		mode         string
		missingCount int
		want         bool
	}{
		{mode: "strict", missingCount: 0, want: true},
		{mode: "strict", missingCount: 1, want: false},
		{mode: "inclusive", missingCount: 3, want: true},
	}

	for _, tc := range tests {
		got := passesMode(tc.mode, tc.missingCount)
		if got != tc.want {
			t.Fatalf("passesMode(%q, %d) = %v, want %v", tc.mode, tc.missingCount, got, tc.want)
		}
	}
}
