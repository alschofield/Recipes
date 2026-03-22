package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"recipes/pkg/storage/postgres"
)

func main() {
	ctx := context.Background()
	pool := storage.Pool()
	defer pool.Close()

	fmt.Println("Seeding sample data...")

	sampleRecipes := []struct {
		Name        string
		Description string
		Steps       string
		Ingredients []string // canonical names
	}{
		{
			Name:        "Garlic Chicken Rice Bowl",
			Description: "Simple weeknight chicken with rice and garlic.",
			Steps:       `["Cook rice per package directions","Dice chicken and season","Saute chicken in olive oil with garlic","Serve over rice"]`,
			Ingredients: []string{"chicken", "rice", "garlic", "olive oil", "salt", "pepper"},
		},
		{
			Name:        "Tomato Basil Pasta",
			Description: "Classic tomato and basil pasta.",
			Steps:       `["Boil pasta","Saute garlic in olive oil","Add canned tomato and basil","Simmer 15 min","Toss with pasta"]`,
			Ingredients: []string{"pasta", "canned tomato", "garlic", "olive oil", "basil", "salt"},
		},
		{
			Name:        "Egg Fried Rice",
			Description: "Quick fried rice with eggs and vegetables.",
			Steps:       `["Scramble eggs in hot oil","Add cooked rice and soy sauce","Stir fry on high heat","Add green onion"]`,
			Ingredients: []string{"rice", "egg", "soy sauce", "green onion", "olive oil"},
		},
	}

	for _, recipe := range sampleRecipes {
		// Look up ingredient IDs
		rows, err := pool.Query(ctx,
			`SELECT id FROM ingredients WHERE canonical_name = ANY($1)`,
			recipe.Ingredients)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to lookup ingredients for %q: %v\n", recipe.Name, err)
			continue
		}

		var ingredientIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: scan failed: %v\n", err)
				continue
			}
			ingredientIDs = append(ingredientIDs, id)
		}
		rows.Close()

		if len(ingredientIDs) == 0 {
			fmt.Printf("Skipped %q (no ingredients found)\n", recipe.Name)
			continue
		}

		// Insert recipe
		var recipeID string
		err = pool.QueryRow(ctx,
			`INSERT INTO recipes (name, description, steps, source, servings, difficulty, quality_score)
			 VALUES ($1, $2, $3::jsonb, 'database', 2, 'easy', 0.8)
			 ON CONFLICT DO NOTHING
			 RETURNING id`,
			recipe.Name, recipe.Description, recipe.Steps,
		).Scan(&recipeID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to insert recipe %q: %v\n", recipe.Name, err)
			continue
		}

		// Link ingredients
		for i, ingID := range ingredientIDs {
			_, err = pool.Exec(ctx,
				`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, position, optional)
				 VALUES ($1, $2, $3, false)
				 ON CONFLICT DO NOTHING`,
				recipeID, ingID, i)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to link ingredient %s: %v\n", ingID, err)
			}
		}

		fmt.Printf("Seeded: %s (%d ingredients)\n", recipe.Name, len(ingredientIDs))
	}

	_ = strings.NewReader // keep import active
	fmt.Println("Sample data seeding complete.")
}
