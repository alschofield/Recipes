-- 002_search_indexes.down.sql
-- Revert: drop search indexes.

DROP INDEX IF EXISTS idx_recipes_llm_source;
DROP INDEX IF EXISTS idx_recipes_prep;
DROP INDEX IF EXISTS idx_recipes_name_fts;
DROP INDEX IF EXISTS idx_recipes_dietary_tags;
DROP INDEX IF EXISTS idx_recipes_cuisine_difficulty;
DROP INDEX IF EXISTS idx_recipe_ingredients_lookup;
DROP INDEX IF EXISTS idx_recipes_quality;
