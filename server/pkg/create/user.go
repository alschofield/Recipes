package create

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"

	"golang.org/x/crypto/bcrypt"
)

// SignupRequest is the DTO for user registration.
type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupResponse is the DTO for a successful signup.
type SignupResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// HandleSignup validates input and creates a new user.
func HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	// Validation
	if len(req.Username) < 3 {
		response.BadRequest(w, "Username must be at least 3 characters")
		return
	}
	if len(req.Username) > 50 {
		response.BadRequest(w, "Username must be 50 characters or less")
		return
	}
	if len(req.Email) < 5 || !containsAt(req.Email) {
		response.BadRequest(w, "Email is invalid")
		return
	}
	if err := validatePassword(req.Password, req.Username, req.Email); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Check for existing user
	pool := storage.Pool()
	var exists bool
	err := pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)`,
		req.Username, req.Email,
	).Scan(&exists)
	if err != nil {
		response.InternalError(w, "Database error")
		return
	}
	if exists {
		response.Conflict(w, "Username or email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(w, "Failed to hash password")
		return
	}

	// Insert user
	var user SignupResponse
	err = pool.QueryRow(r.Context(),
		`INSERT INTO users (username, email, password_hash, role)
		 VALUES ($1, $2, $3, 'user')
		 RETURNING id, username, email, role`,
		req.Username, req.Email, string(hashedPassword),
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		response.InternalError(w, "Failed to create user")
		return
	}

	response.WriteJSON(w, http.StatusCreated, user)
}

func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}

func validatePassword(password, username, email string) error {
	if len(password) < 12 {
		return errors.New("Password must be at least 12 characters")
	}

	lower := strings.ToLower(password)
	if strings.Contains(lower, "password") || strings.Contains(lower, "123456") {
		return errors.New("Password is too weak")
	}

	if username != "" && strings.Contains(lower, strings.ToLower(username)) {
		return errors.New("Password must not include username")
	}

	if email != "" {
		localPart := strings.Split(strings.ToLower(email), "@")[0]
		if localPart != "" && strings.Contains(lower, localPart) {
			return errors.New("Password must not include email name")
		}
	}

	return nil
}
