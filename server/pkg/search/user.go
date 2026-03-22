package search

import (
	"encoding/json"
	"net/http"
	"time"

	"recipes/pkg/auth"
	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest is the DTO for user login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the DTO for a successful login.
type LoginResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
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

	response.WriteJSON(w, http.StatusOK, LoginResponse{
		ID:        userID,
		Username:  username,
		Email:     email,
		Role:      role,
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
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
