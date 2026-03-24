package delete

import (
	"net/http"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

// RemoveFavorite handles DELETE /favorites/{userid}/{recipeid}.
func RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	recipeID := r.PathValue("recipeid")
	if userID == "" || recipeID == "" {
		response.BadRequest(w, "User ID and recipe ID are required")
		return
	}

	pool := storage.Pool()
	result, err := pool.Exec(r.Context(),
		`DELETE FROM favorites WHERE user_id = $1 AND recipe_id = $2`,
		userID,
		recipeID,
	)
	if err != nil {
		response.InternalError(w, "Failed to remove favorite")
		return
	}

	writeRemoveFavoriteResponse(w, result.RowsAffected())
}

func writeRemoveFavoriteResponse(w http.ResponseWriter, rowsAffected int64) {
	if rowsAffected == 0 {
		w.Header().Set("Idempotency-Status", "replayed")
	}

	w.WriteHeader(http.StatusNoContent)
}
