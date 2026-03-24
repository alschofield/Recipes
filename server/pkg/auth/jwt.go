package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
	defaultIssuer          = "recipes-users-server"
	tokenTypeAccess        = "access"
	tokenTypeRefresh       = "refresh"
)

// Claims is the JWT payload used for access tokens.
type Claims struct {
	UserID  string `json:"userId"`
	Role    string `json:"role"`
	Type    string `json:"type"`
	Family  string `json:"familyId,omitempty"`
	Session string `json:"sessionId,omitempty"`
	jwt.RegisteredClaims
}

// Principal is the authenticated identity extracted from a token.
type Principal struct {
	UserID string
	Role   string
}

// RefreshPrincipal is the identity extracted from a refresh token.
type RefreshPrincipal struct {
	UserID    string
	Role      string
	FamilyID  string
	SessionID string
	TokenID   string
	ExpiresAt time.Time
}

// GenerateAccessToken returns a signed JWT with short-lived claims.
func GenerateAccessToken(userID, role string) (string, time.Time, error) {
	secret := jwtSecret()
	if secret == "" {
		return "", time.Time{}, errors.New("JWT_SECRET is required")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(accessTokenTTL())

	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

// GenerateRefreshToken returns a signed JWT for refresh flow.
func GenerateRefreshToken(userID, role, familyID, sessionID, tokenID string) (string, time.Time, error) {
	secret := jwtSecret()
	if secret == "" {
		return "", time.Time{}, errors.New("JWT_SECRET is required")
	}

	if familyID == "" || sessionID == "" || tokenID == "" {
		return "", time.Time{}, errors.New("familyID, sessionID and tokenID are required")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(refreshTokenTTL())

	claims := Claims{
		UserID:  userID,
		Role:    role,
		Type:    tokenTypeRefresh,
		Family:  familyID,
		Session: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   userID,
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

// ParseAccessToken validates and parses a bearer token into a principal.
func ParseAccessToken(tokenString string) (Principal, error) {
	secret := jwtSecret()
	if secret == "" {
		return Principal{}, errors.New("JWT_SECRET is required")
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(tokenIssuer()), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return Principal{}, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Principal{}, errors.New("invalid token claims")
	}

	if claims.UserID == "" || claims.Role == "" {
		return Principal{}, errors.New("missing token claims")
	}

	if claims.Type != tokenTypeAccess {
		return Principal{}, errors.New("invalid token type")
	}

	return Principal{UserID: claims.UserID, Role: claims.Role}, nil
}

// ParseRefreshToken validates and parses a refresh token into a refresh principal.
func ParseRefreshToken(tokenString string) (RefreshPrincipal, error) {
	secret := jwtSecret()
	if secret == "" {
		return RefreshPrincipal{}, errors.New("JWT_SECRET is required")
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(tokenIssuer()), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return RefreshPrincipal{}, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return RefreshPrincipal{}, errors.New("invalid token claims")
	}

	if claims.UserID == "" || claims.Role == "" || claims.Family == "" || claims.Session == "" || claims.ID == "" {
		return RefreshPrincipal{}, errors.New("missing token claims")
	}

	if claims.Type != tokenTypeRefresh {
		return RefreshPrincipal{}, errors.New("invalid token type")
	}

	expiresAt := time.Time{}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}

	return RefreshPrincipal{
		UserID:    claims.UserID,
		Role:      claims.Role,
		FamilyID:  claims.Family,
		SessionID: claims.Session,
		TokenID:   claims.ID,
		ExpiresAt: expiresAt,
	}, nil
}

func accessTokenTTL() time.Duration {
	raw := os.Getenv("JWT_ACCESS_TTL")
	if raw == "" {
		return defaultAccessTokenTTL
	}

	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl <= 0 {
		return defaultAccessTokenTTL
	}

	return ttl
}

func refreshTokenTTL() time.Duration {
	raw := os.Getenv("JWT_REFRESH_TTL")
	if raw == "" {
		return defaultRefreshTokenTTL
	}

	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl <= 0 {
		return defaultRefreshTokenTTL
	}

	return ttl
}

func tokenIssuer() string {
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		return defaultIssuer
	}

	return issuer
}

func jwtSecret() string {
	return os.Getenv("JWT_SECRET")
}
