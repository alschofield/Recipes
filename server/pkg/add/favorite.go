package add

import (
	"errors"
	"net/http"
	"time"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"

	"github.com/jackc/pgx/v5/pgconn"
)

type favoriteResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	RecipeID  string    `json:"recipeId"`
	CreatedAt time.Time `json:"createdAt"`
}

// AddFavorite handles POST /favorites/{userid}/{recipeid}.
func AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	recipeID := r.PathValue("recipeid")
	if userID == "" || recipeID == "" {
		response.BadRequest(w, "User ID and recipe ID are required")
		return
	}

	pool := storage.Pool()

	var created favoriteResponse
	err := pool.QueryRow(r.Context(),
		`INSERT INTO favorites (user_id, recipe_id)
		 VALUES ($1, $2)
		 RETURNING id, user_id, recipe_id, created_at`,
		userID,
		recipeID,
	).Scan(&created.ID, &created.UserID, &created.RecipeID, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				response.Conflict(w, "Recipe is already favorited")
				return
			case "23503":
				response.NotFound(w, "User or recipe not found")
				return
			}
		}

		response.InternalError(w, "Failed to add favorite")
		return
	}

	response.WriteJSON(w, http.StatusCreated, created)
}
