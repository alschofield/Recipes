-- 002_search_indexes.up.sql
-- Indexes optimized for recipe search and ranking.

-- Composite index for recipe ranking (quality score descending)
CREATE INDEX IF NOT EXISTS idx_recipes_quality ON recipes(quality_score DESC);

-- Composite index for ingredient-based search + ranking
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_lookup
    ON recipe_ingredients(ingredient_id, recipe_id);

-- Index for recipe search by cuisine + difficulty
CREATE INDEX IF NOT EXISTS idx_recipes_cuisine_difficulty ON recipes(cuisine, difficulty);

-- Index for dietary tag filtering
CREATE INDEX IF NOT EXISTS idx_recipes_dietary_tags ON recipes USING GIN(dietary_tags);

-- Full-text search on recipe name + description
CREATE INDEX IF NOT EXISTS idx_recipes_name_fts
    ON recipes USING GIN(to_tsvector('english', name));

-- Index for recipe search by prep time
CREATE INDEX IF NOT EXISTS idx_recipes_prep ON recipes(prep_minutes) WHERE prep_minutes IS NOT NULL;

-- Index for LLM-generated recipes (used for fallback decisions)
CREATE INDEX IF NOT EXISTS idx_recipes_llm_source ON recipes(source, quality_score) WHERE source = 'llm';
