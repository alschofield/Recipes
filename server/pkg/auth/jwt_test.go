package auth

import "testing"

func TestGenerateAndParseAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123456789")
	t.Setenv("JWT_ISSUER", "recipes-test")

	token, _, err := GenerateAccessToken("user-1", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	principal, err := ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken failed: %v", err)
	}

	if principal.UserID != "user-1" || principal.Role != "user" {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestRefreshTokenCannotBeUsedAsAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123456789")
	t.Setenv("JWT_ISSUER", "recipes-test")

	refreshToken, _, err := GenerateRefreshToken("user-1", "user", "family-1", "session-1", "token-1")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if _, err := ParseAccessToken(refreshToken); err == nil {
		t.Fatal("expected ParseAccessToken to reject refresh token")
	}
}

func TestGenerateAndParseRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123456789")
	t.Setenv("JWT_ISSUER", "recipes-test")

	token, _, err := GenerateRefreshToken("user-2", "admin", "family-a", "session-a", "token-a")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	principal, err := ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("ParseRefreshToken failed: %v", err)
	}

	if principal.UserID != "user-2" || principal.Role != "admin" || principal.FamilyID != "family-a" || principal.SessionID != "session-a" || principal.TokenID != "token-a" {
		t.Fatalf("unexpected refresh principal: %+v", principal)
	}
}
