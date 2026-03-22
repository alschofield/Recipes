package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultAccessTokenTTL = 15 * time.Minute
	defaultIssuer         = "recipes-users-server"
)

// Claims is the JWT payload used for access tokens.
type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Principal is the authenticated identity extracted from a token.
type Principal struct {
	UserID string
	Role   string
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

	return Principal{UserID: claims.UserID, Role: claims.Role}, nil
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
