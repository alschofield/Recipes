package ingredients

import (
	"os"
	"testing"
)

func TestIngredientPolicyModeDefaultsToAutoCreateOutsideProduction(t *testing.T) {
	originalMode := os.Getenv("INGREDIENT_POLICY_MODE")
	originalAppEnv := os.Getenv("APP_ENV")
	originalGoEnv := os.Getenv("GO_ENV")
	t.Cleanup(func() {
		_ = os.Setenv("INGREDIENT_POLICY_MODE", originalMode)
		_ = os.Setenv("APP_ENV", originalAppEnv)
		_ = os.Setenv("GO_ENV", originalGoEnv)
	})

	_ = os.Setenv("INGREDIENT_POLICY_MODE", "")
	_ = os.Setenv("APP_ENV", "development")
	_ = os.Setenv("GO_ENV", "")

	if got := ingredientPolicyMode(); got != ingredientPolicyAutoCreate {
		t.Fatalf("expected %q, got %q", ingredientPolicyAutoCreate, got)
	}
}

func TestIngredientPolicyModeDefaultsToQueueOnlyInProduction(t *testing.T) {
	originalMode := os.Getenv("INGREDIENT_POLICY_MODE")
	originalAppEnv := os.Getenv("APP_ENV")
	originalGoEnv := os.Getenv("GO_ENV")
	t.Cleanup(func() {
		_ = os.Setenv("INGREDIENT_POLICY_MODE", originalMode)
		_ = os.Setenv("APP_ENV", originalAppEnv)
		_ = os.Setenv("GO_ENV", originalGoEnv)
	})

	_ = os.Setenv("INGREDIENT_POLICY_MODE", "")
	_ = os.Setenv("APP_ENV", "production")
	_ = os.Setenv("GO_ENV", "")

	if got := ingredientPolicyMode(); got != ingredientPolicyQueueOnly {
		t.Fatalf("expected %q, got %q", ingredientPolicyQueueOnly, got)
	}
}

func TestIngredientPolicyModeHonorsValidOverride(t *testing.T) {
	originalMode := os.Getenv("INGREDIENT_POLICY_MODE")
	originalAppEnv := os.Getenv("APP_ENV")
	t.Cleanup(func() {
		_ = os.Setenv("INGREDIENT_POLICY_MODE", originalMode)
		_ = os.Setenv("APP_ENV", originalAppEnv)
	})

	_ = os.Setenv("APP_ENV", "production")
	_ = os.Setenv("INGREDIENT_POLICY_MODE", ingredientPolicyAutoCreate)

	if got := ingredientPolicyMode(); got != ingredientPolicyAutoCreate {
		t.Fatalf("expected %q, got %q", ingredientPolicyAutoCreate, got)
	}
}

func TestIngredientPolicyModeIgnoresInvalidOverride(t *testing.T) {
	originalMode := os.Getenv("INGREDIENT_POLICY_MODE")
	originalAppEnv := os.Getenv("APP_ENV")
	t.Cleanup(func() {
		_ = os.Setenv("INGREDIENT_POLICY_MODE", originalMode)
		_ = os.Setenv("APP_ENV", originalAppEnv)
	})

	_ = os.Setenv("APP_ENV", "production")
	_ = os.Setenv("INGREDIENT_POLICY_MODE", "maybe")

	if got := ingredientPolicyMode(); got != ingredientPolicyQueueOnly {
		t.Fatalf("expected %q, got %q", ingredientPolicyQueueOnly, got)
	}
}
