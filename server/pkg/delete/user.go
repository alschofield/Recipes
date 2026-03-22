package delete

import (
	"net/http"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

// HandleDeleteUser handles DELETE /users/{userid}.
func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	if userID == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	pool := storage.Pool()

	// Check user exists
	var exists bool
	err := pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID,
	).Scan(&exists)
	if err != nil {
		response.InternalError(w, "Database error")
		return
	}
	if !exists {
		response.NotFound(w, "User not found")
		return
	}

	// Delete user (cascades favorites via FK)
	result, err := pool.Exec(r.Context(),
		`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		response.InternalError(w, "Failed to delete user")
		return
	}

	if result.RowsAffected() == 0 {
		response.NotFound(w, "User not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
