package search

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"recipes/pkg/auth"
	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest is the DTO for user login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the DTO for a successful login.
type LoginResponse struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	Token            string `json:"token"`
	ExpiresAt        string `json:"expiresAt"`
	RefreshToken     string `json:"refreshToken,omitempty"`
	RefreshExpiresAt string `json:"refreshExpiresAt,omitempty"`
	SessionID        string `json:"sessionId,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshResponse struct {
	Token            string `json:"token"`
	ExpiresAt        string `json:"expiresAt"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresAt string `json:"refreshExpiresAt"`
	SessionID        string `json:"sessionId"`
}

type SessionItem struct {
	SessionID  string     `json:"sessionId"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	UserAgent  *string    `json:"userAgent,omitempty"`
	IPAddress  *string    `json:"ipAddress,omitempty"`
}

// HandleLogin validates credentials and returns user info.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.BadRequest(w, "Username and password are required")
		return
	}

	pool := storage.Pool()

	// Find user by username or email
	var userID, username, email, passwordHash, role string
	err := pool.QueryRow(r.Context(),
		`SELECT id, username, email, password_hash, role
		 FROM users
		 WHERE username = $1 OR email = $1`,
		req.Username,
	).Scan(&userID, &username, &email, &passwordHash, &role)
	if err != nil {
		// Don't reveal whether user exists
		response.Unauthorized(w, "Invalid username or password")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		// Don't reveal whether password was wrong
		response.Unauthorized(w, "Invalid username or password")
		return
	}

	token, expiresAt, err := auth.GenerateAccessToken(userID, role)
	if err != nil {
		response.InternalError(w, "Failed to generate access token")
		return
	}

	sessionID := clientSessionID(r)
	refreshToken, refreshExpiresAt, err := createInitialRefreshSession(r.Context(), pool, userID, role, sessionID, r.UserAgent(), requestIP(r))
	if err != nil {
		response.InternalError(w, "Failed to create refresh session")
		return
	}

	response.WriteJSON(w, http.StatusOK, LoginResponse{
		ID:               userID,
		Username:         username,
		Email:            email,
		Role:             role,
		Token:            token,
		ExpiresAt:        expiresAt.Format(time.RFC3339),
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt.Format(time.RFC3339),
		SessionID:        sessionID,
	})
}

// HandleRefresh rotates a refresh token and returns fresh access/refresh tokens.
func HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		response.BadRequest(w, "refreshToken is required")
		return
	}

	principal, err := auth.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired refresh token")
		return
	}

	pool := storage.Pool()
	newRefreshToken, newRefreshExpiresAt, err := rotateRefreshSession(r.Context(), pool, principal, req.RefreshToken, r.UserAgent(), requestIP(r))
	if err != nil {
		if errors.Is(err, errInvalidRefreshSession) {
			response.Unauthorized(w, "Invalid or expired refresh token")
			return
		}
		response.InternalError(w, "Failed to rotate refresh session")
		return
	}

	accessToken, accessExpiresAt, err := auth.GenerateAccessToken(principal.UserID, principal.Role)
	if err != nil {
		response.InternalError(w, "Failed to generate access token")
		return
	}

	response.WriteJSON(w, http.StatusOK, RefreshResponse{
		Token:            accessToken,
		ExpiresAt:        accessExpiresAt.Format(time.RFC3339),
		RefreshToken:     newRefreshToken,
		RefreshExpiresAt: newRefreshExpiresAt.Format(time.RFC3339),
		SessionID:        principal.SessionID,
	})
}

// HandleLogout revokes all refresh tokens in the current family.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		response.BadRequest(w, "refreshToken is required")
		return
	}

	principal, err := auth.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired refresh token")
		return
	}

	pool := storage.Pool()
	if err := revokeRefreshFamily(r.Context(), pool, principal.UserID, principal.FamilyID); err != nil {
		response.InternalError(w, "Failed to revoke refresh session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleLogoutSession revokes only the current client session.
func HandleLogoutSession(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		response.BadRequest(w, "refreshToken is required")
		return
	}

	principal, err := auth.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired refresh token")
		return
	}

	pool := storage.Pool()
	if err := revokeRefreshSession(r.Context(), pool, principal.UserID, principal.SessionID); err != nil {
		response.InternalError(w, "Failed to revoke refresh session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListSessions handles GET /users/{userid}/sessions.
func ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	if userID == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	pool := storage.Pool()
	rows, err := pool.Query(r.Context(), `
		SELECT client_session_id, MIN(created_at) AS created_at, MAX(last_used_at) AS last_used_at,
		       MAX(user_agent) AS user_agent, MAX(ip_address) AS ip_address
		FROM user_refresh_sessions
		WHERE user_id = $1 AND revoked_at IS NULL
		GROUP BY client_session_id
		ORDER BY MAX(COALESCE(last_used_at, created_at)) DESC
	`, userID)
	if err != nil {
		response.InternalError(w, "Failed to fetch sessions")
		return
	}
	defer rows.Close()

	sessions := []SessionItem{}
	for rows.Next() {
		var item SessionItem
		if err := rows.Scan(&item.SessionID, &item.CreatedAt, &item.LastUsedAt, &item.UserAgent, &item.IPAddress); err != nil {
			response.InternalError(w, "Failed to parse sessions")
			return
		}
		sessions = append(sessions, item)
	}

	response.WriteJSON(w, http.StatusOK, sessions)
}

// GetProfile handles GET /users/{userid}.
func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	if userID == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	pool := storage.Pool()
	var id, username, email, role string
	err := pool.QueryRow(r.Context(),
		`SELECT id, username, email, role FROM users WHERE id = $1`,
		userID,
	).Scan(&id, &username, &email, &role)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"id":       id,
		"username": username,
		"email":    email,
		"role":     role,
	})
}

// ListUsers handles GET /users (returns all users).
func ListUsers(w http.ResponseWriter, r *http.Request) {
	pool := storage.Pool()
	rows, err := pool.Query(r.Context(),
		`SELECT id, username, email, role, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		response.InternalError(w, "Failed to fetch users")
		return
	}
	defer rows.Close()

	type userRow struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"createdAt"`
	}

	var users []userRow
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}

	if users == nil {
		users = []userRow{}
	}

	response.WriteJSON(w, http.StatusOK, users)
}

var errInvalidRefreshSession = errors.New("invalid refresh session")

func createInitialRefreshSession(ctx context.Context, pool *pgxpool.Pool, userID, role, sessionID, userAgent, ipAddress string) (string, time.Time, error) {

	familyID := uuid.NewString()
	tokenID := uuid.NewString()
	refreshToken, refreshExpiresAt, err := auth.GenerateRefreshToken(userID, role, familyID, sessionID, tokenID)
	if err != nil {
		return "", time.Time{}, err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO user_refresh_sessions (user_id, token_id, family_id, client_session_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, userID, tokenID, familyID, sessionID, hashRefreshToken(refreshToken), refreshExpiresAt, nullableString(userAgent), nullableString(ipAddress))
	if err != nil {
		return "", time.Time{}, err
	}

	return refreshToken, refreshExpiresAt, nil
}

func rotateRefreshSession(ctx context.Context, pool *pgxpool.Pool, principal auth.RefreshPrincipal, rawToken, userAgent, ipAddress string) (string, time.Time, error) {

	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	defer tx.Rollback(ctx)

	var tokenHash string
	var familyID string
	var sessionID string
	var expiresAt time.Time
	var revokedAt *time.Time
	var replacedBy *string
	err = tx.QueryRow(ctx, `
		SELECT token_hash, family_id, client_session_id, expires_at, revoked_at, replaced_by_token_id
		FROM user_refresh_sessions
		WHERE token_id = $1 AND user_id = $2
		FOR UPDATE
	`, principal.TokenID, principal.UserID).Scan(&tokenHash, &familyID, &sessionID, &expiresAt, &revokedAt, &replacedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, errInvalidRefreshSession
		}
		return "", time.Time{}, err
	}

	if familyID != principal.FamilyID {
		return "", time.Time{}, errInvalidRefreshSession
	}

	if sessionID != principal.SessionID {
		return "", time.Time{}, errInvalidRefreshSession
	}

	if revokedAt != nil {
		if replacedBy != nil {
			_ = revokeRefreshFamilyTx(ctx, tx, principal.UserID, principal.FamilyID)
		}
		return "", time.Time{}, errInvalidRefreshSession
	}

	if time.Now().UTC().After(expiresAt) {
		_ = revokeRefreshFamilyTx(ctx, tx, principal.UserID, principal.FamilyID)
		return "", time.Time{}, errInvalidRefreshSession
	}

	if hashRefreshToken(rawToken) != tokenHash {
		return "", time.Time{}, errInvalidRefreshSession
	}

	newTokenID := uuid.NewString()
	newRefreshToken, newRefreshExpiresAt, err := auth.GenerateRefreshToken(principal.UserID, principal.Role, principal.FamilyID, principal.SessionID, newTokenID)
	if err != nil {
		return "", time.Time{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE user_refresh_sessions
		SET revoked_at = NOW(), last_used_at = NOW(), replaced_by_token_id = $3
		WHERE token_id = $1 AND user_id = $2
	`, principal.TokenID, principal.UserID, newTokenID)
	if err != nil {
		return "", time.Time{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_refresh_sessions (user_id, token_id, family_id, client_session_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, principal.UserID, newTokenID, principal.FamilyID, principal.SessionID, hashRefreshToken(newRefreshToken), newRefreshExpiresAt, nullableString(userAgent), nullableString(ipAddress))
	if err != nil {
		return "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, err
	}

	return newRefreshToken, newRefreshExpiresAt, nil
}

func revokeRefreshFamily(ctx context.Context, pool *pgxpool.Pool, userID, familyID string) error {

	_, err := pool.Exec(ctx, `
		UPDATE user_refresh_sessions
		SET revoked_at = NOW(), last_used_at = NOW()
		WHERE user_id = $1 AND family_id = $2 AND revoked_at IS NULL
	`, userID, familyID)
	return err
}

func revokeRefreshFamilyTx(ctx context.Context, tx pgx.Tx, userID, familyID string) error {

	_, err := tx.Exec(ctx, `
		UPDATE user_refresh_sessions
		SET revoked_at = NOW(), last_used_at = NOW()
		WHERE user_id = $1 AND family_id = $2 AND revoked_at IS NULL
	`, userID, familyID)
	return err
}

func revokeRefreshSession(ctx context.Context, pool *pgxpool.Pool, userID, sessionID string) error {
	_, err := pool.Exec(ctx, `
		UPDATE user_refresh_sessions
		SET revoked_at = NOW(), last_used_at = NOW()
		WHERE user_id = $1 AND client_session_id = $2 AND revoked_at IS NULL
	`, userID, sessionID)
	return err
}

func hashRefreshToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func nullableString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func requestIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		candidate := strings.TrimSpace(parts[0])
		if addr, err := netip.ParseAddr(candidate); err == nil {
			return addr.String()
		}
	}

	remote := strings.TrimSpace(r.RemoteAddr)
	if remote == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(remote); err == nil {
		if addr, parseErr := netip.ParseAddr(host); parseErr == nil {
			return addr.String()
		}
	}
	if addr, err := netip.ParseAddr(remote); err == nil {
		return addr.String()
	}

	return ""
}

func clientSessionID(r *http.Request) string {
	raw := strings.TrimSpace(r.Header.Get("X-Client-Session-ID"))
	if raw != "" {
		if len(raw) > 128 {
			return raw[:128]
		}
		return raw
	}

	return uuid.NewString()
}
