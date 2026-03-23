package repository

import (
	"context"
	"recipes/pkg/models"
)

// UserRepo handles user persistence.
type UserRepo interface {
	Create(ctx context.Context, username, email, passwordHash string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id string, username, email, passwordHash *string) (*models.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]models.User, error)
}

// RecipeRepo handles recipe persistence and search.
type RecipeRepo interface {
	Create(ctx context.Context, recipe models.Recipe) (string, error)
	GetByID(ctx context.Context, id string) (*models.Recipe, error)
	List(ctx context.Context, limit, offset int) ([]models.Recipe, error)
	Search(ctx context.Context, req SearchRequest) (*SearchResult, error)
}

// FavoriteRepo handles favorite persistence.
type FavoriteRepo interface {
	Add(ctx context.Context, userID, recipeID string) error
	Remove(ctx context.Context, userID, recipeID string) error
	List(ctx context.Context, userID string) ([]models.Favorite, error)
}

// IngredientRepo handles ingredient persistence and normalization.
type IngredientRepo interface {
	Create(ctx context.Context, canonicalName string, aliases []string) (*models.Ingredient, error)
	Normalize(ctx context.Context, names []string) ([]*models.Ingredient, error)
	List(ctx context.Context, limit, offset int) ([]models.Ingredient, error)
}

// SearchRequest is the input for recipe search.
type SearchRequest struct {
	Ingredients []string          `json:"ingredients"`
	Mode        string            `json:"mode"`
	Complex     bool              `json:"complex,omitempty"`
	Filters     SearchFilters     `json:"filters,omitempty"`
	Pagination  PaginationRequest `json:"pagination,omitempty"`
}

// SearchFilters are optional filters for recipe search.
type SearchFilters struct {
	MaxPrepMinutes *int     `json:"maxPrepMinutes,omitempty"`
	Cuisine        []string `json:"cuisine,omitempty"`
	Dietary        []string `json:"dietary,omitempty"`
	Difficulty     []string `json:"difficulty,omitempty"`
}

// PaginationRequest controls result pagination.
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// SearchResult is the response from a recipe search.
type SearchResult struct {
	Mode        string                      `json:"mode"`
	Ingredients []string                    `json:"ingredients"`
	Pagination  PaginationResult            `json:"pagination"`
	Results     []models.SearchResultRecipe `json:"results"`
}

// PaginationResult is the pagination metadata for a search response.
type PaginationResult struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}
