package search

import (
	"net/http"
	"time"

	"recipes/pkg/response"
	"recipes/pkg/storage/postgres"
)

type favoriteItem struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	RecipeID   string    `json:"recipeId"`
	RecipeName *string   `json:"recipeName,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ListFavorites handles GET /favorites/{userid}.
func ListFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userid")
	if userID == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	pool := storage.Pool()
	rows, err := pool.Query(r.Context(),
		`SELECT f.id, f.user_id, f.recipe_id, f.created_at, rc.name
		 FROM favorites f
		 LEFT JOIN recipes rc ON rc.id = f.recipe_id
		 WHERE f.user_id = $1
		 ORDER BY f.created_at DESC`,
		userID,
	)
	if err != nil {
		response.InternalError(w, "Failed to fetch favorites")
		return
	}
	defer rows.Close()

	favorites := []favoriteItem{}
	for rows.Next() {
		var item favoriteItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.RecipeID, &item.CreatedAt, &item.RecipeName); err != nil {
			response.InternalError(w, "Failed to parse favorites")
			return
		}
		favorites = append(favorites, item)
	}

	response.WriteJSON(w, http.StatusOK, favorites)
}
