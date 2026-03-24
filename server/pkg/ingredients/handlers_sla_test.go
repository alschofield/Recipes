package ingredients

import "testing"

func TestIngredientCandidateSLAHoursDefault(t *testing.T) {
	t.Setenv("INGREDIENT_CANDIDATE_SLA_HOURS", "")
	if got := ingredientCandidateSLAHours(); got != 48 {
		t.Fatalf("expected default 48, got %v", got)
	}
}

func TestIngredientCandidateSLAHoursParsesValue(t *testing.T) {
	t.Setenv("INGREDIENT_CANDIDATE_SLA_HOURS", "24")
	if got := ingredientCandidateSLAHours(); got != 24 {
		t.Fatalf("expected 24, got %v", got)
	}
}

func TestIngredientCandidateSLAHoursInvalidFallsBack(t *testing.T) {
	t.Setenv("INGREDIENT_CANDIDATE_SLA_HOURS", "nope")
	if got := ingredientCandidateSLAHours(); got != 48 {
		t.Fatalf("expected default 48, got %v", got)
	}
}
