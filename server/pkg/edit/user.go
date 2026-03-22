package edit

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"

	"golang.org/x/crypto/bcrypt"
)

// UpdateProfileRequest is the DTO for updating a user profile.
type UpdateProfileRequest struct {
	Username        *string `json:"username,omitempty"`
	Email           *string `json:"email,omitempty"`
	CurrentPassword *string `json:"currentPassword,omitempty"`
	NewPassword     *string `json:"newPassword,omitempty"`
}

// HandleUpdateProfile handles PUT /users/{userid}.
func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	if userID == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON body")
		return
	}

	if req.Username == nil && req.Email == nil && req.NewPassword == nil {
		response.BadRequest(w, "No fields to update")
		return
	}

	// Password change requires current password
	if req.NewPassword != nil {
		if req.CurrentPassword == nil {
			response.BadRequest(w, "Current password is required to change password")
			return
		}
		if err := validateNewPassword(*req.NewPassword); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
	}

	// Username validation
	if req.Username != nil {
		if len(*req.Username) < 3 {
			response.BadRequest(w, "Username must be at least 3 characters")
			return
		}
		if len(*req.Username) > 50 {
			response.BadRequest(w, "Username must be 50 characters or less")
			return
		}
	}

	// Email validation
	if req.Email != nil {
		if len(*req.Email) < 5 {
			response.BadRequest(w, "Email is invalid")
			return
		}
	}

	pool := storage.Pool()

	// Verify current password if changing password
	if req.NewPassword != nil {
		var storedHash string
		err := pool.QueryRow(r.Context(),
			`SELECT password_hash FROM users WHERE id = $1`, userID,
		).Scan(&storedHash)
		if err != nil {
			response.NotFound(w, "User not found")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(*req.CurrentPassword)); err != nil {
			response.BadRequest(w, "Current password is incorrect")
			return
		}
	}

	// Check username uniqueness
	if req.Username != nil {
		var exists bool
		err := pool.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)`,
			*req.Username, userID,
		).Scan(&exists)
		if err != nil {
			response.InternalError(w, "Database error")
			return
		}
		if exists {
			response.Conflict(w, "Username already taken")
			return
		}
	}

	// Check email uniqueness
	if req.Email != nil {
		var exists bool
		err := pool.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`,
			*req.Email, userID,
		).Scan(&exists)
		if err != nil {
			response.InternalError(w, "Database error")
			return
		}
		if exists {
			response.Conflict(w, "Email already taken")
			return
		}
	}

	// Build dynamic update
	var setClauses []string
	var args []any
	argIdx := 1

	if req.Username != nil {
		setClauses = append(setClauses, "username = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Username)
		argIdx++
	}
	if req.Email != nil {
		setClauses = append(setClauses, "email = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.NewPassword != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			response.InternalError(w, "Failed to hash password")
			return
		}
		setClauses = append(setClauses, "password_hash = $"+strconv.Itoa(argIdx))
		args = append(args, string(hashed))
		argIdx++
	}

	args = append(args, userID)

	query := `UPDATE users SET ` + joinClauses(setClauses) + ` WHERE id = $` + strconv.Itoa(argIdx) + ` RETURNING id, username, email, role`
	var id, username, email, role string
	err := pool.QueryRow(r.Context(), query, args...).Scan(&id, &username, &email, &role)
	if err != nil {
		response.InternalError(w, "Failed to update user")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"id":       id,
		"username": username,
		"email":    email,
		"role":     role,
	})
}

func joinClauses(clauses []string) string {
	result := ""
	for i, c := range clauses {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

func validateNewPassword(password string) error {
	if len(password) < 12 {
		return errors.New("New password must be at least 12 characters")
	}

	lower := strings.ToLower(password)
	if strings.Contains(lower, "password") || strings.Contains(lower, "123456") {
		return errors.New("New password is too weak")
	}

	return nil
}
