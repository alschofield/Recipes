package ingredients

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: " Tomatoes ", want: "tomato"},
		{input: "green onions", want: "green onion"},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		if got := NormalizeName(tt.input); got != tt.want {
			t.Fatalf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSimilarityScore(t *testing.T) {
	if similarityScore("green onion", "green onion") != 1 {
		t.Fatal("expected exact match score to be 1")
	}

	high := similarityScore("green onion", "green onions")
	if high < 0.8 {
		t.Fatalf("expected close words to score high, got %f", high)
	}

	low := similarityScore("beef", "cinnamon")
	if low > 0.4 {
		t.Fatalf("expected unrelated words to score low, got %f", low)
	}
}

func TestLevenshteinDistance(t *testing.T) {
	if d := levenshteinDistance("kitten", "sitting"); d != 3 {
		t.Fatalf("expected distance 3, got %d", d)
	}
}

func TestFuzzyThresholdEnvFallback(t *testing.T) {
	t.Setenv("INGREDIENT_FUZZY_THRESHOLD", "0.97")
	if got := fuzzyThreshold(); got != 0.97 {
		t.Fatalf("expected threshold 0.97, got %f", got)
	}

	t.Setenv("INGREDIENT_FUZZY_THRESHOLD", "bad")
	if got := fuzzyThreshold(); got != defaultFuzzyThreshold {
		t.Fatalf("expected default threshold %f, got %f", defaultFuzzyThreshold, got)
	}
}
