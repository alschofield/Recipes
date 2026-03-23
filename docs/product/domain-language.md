# Domain Language and Service Boundaries

## Core Entities

| Entity | Description | Key Fields |
|--------|-------------|------------|
| User | A registered account | id, username, email, password_hash, role |
| Ingredient | A canonical cooking ingredient | id, canonical_name |
| IngredientAlias | An alternate name for an ingredient | id, ingredient_id, alias |
| Recipe | A cooking instruction set | id, name, description, steps, source, metadata |
| RecipeIngredient | A link between a recipe and one of its ingredients | id, recipe_id, ingredient_id, amount, unit, optional |
| Favorite | A user's saved recipe | id, user_id, recipe_id |

## Core Services

| Service | Responsibility | Owns |
|---------|---------------|------|
| **RecipeFinder** | Search recipes by user-provided ingredients; strict/inclusive mode; ranking; LLM fallback | Recipes, RecipeIngredients, Ingredients |
| **RecipeCreator** | Create recipes (manual or LLM-generated); normalize ingredients; store provenance | Recipes, RecipeIngredients |
| **FavoriteManager** | Add/remove/list user favorites; enforce ownership | Favorites |
| **UserCreator** | Create user accounts; hash passwords; issue JWT | Users |
| **UserDeleter** | Delete user accounts; cascade favorites | Users, Favorites |
| **IngredientCreator** | Manage canonical ingredient list and aliases; normalize names | Ingredients, IngredientAliases |

## Domain Events

| Event | Trigger |
|-------|---------|
| UserCreated | New user account created |
| UserDeleted | User account removed |
| RecipeCreated | Recipe saved to DB (manual or LLM) |
| RecipeFound | Single recipe returned from search |
| RecipesFound | Batch search completed |
| FavoriteAdded | User saves a recipe to favorites |
| FavoriteRemoved | User un-saves a recipe |

## Language Rules

- "ingredients" always refers to canonical ingredient names (lowercase, singular).
- "strict" means all recipe ingredients must be in the user's input.
- "inclusive" means the recipe may require ingredients beyond what the user listed.
- "source" distinguishes `database` (human-curated) vs `llm` (generated) recipes.
- "provenance" tracks generation model, prompt version, and timestamp for LLM recipes.
- "normalize" means lowercasing, alias mapping, and singularization of ingredient names.

## Service Boundaries (Nginx Routing)

| Route Prefix | Service Port | Handler |
|---|---|---|
| `/recipes/*` | 8081 | RecipeFinder, RecipeCreator |
| `/users/*` | 8082 | UserCreator, UserDeleter |
| `/favorites/*` | 8080 | FavoriteManager |
