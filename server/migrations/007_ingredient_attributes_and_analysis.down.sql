-- 007_ingredient_attributes_and_analysis.down.sql

DROP TRIGGER IF EXISTS recipe_quality_analysis_updated_at ON recipe_quality_analysis;
DROP TRIGGER IF EXISTS ingredient_nutrition_profiles_updated_at ON ingredient_nutrition_profiles;

DROP TABLE IF EXISTS recipe_quality_analysis;
DROP TABLE IF EXISTS ingredient_nutrition_profiles;

DROP INDEX IF EXISTS idx_ingredients_source_coverage;
DROP INDEX IF EXISTS idx_ingredients_quality_score;
DROP INDEX IF EXISTS idx_ingredients_analysis_status;

ALTER TABLE ingredients
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS last_analyzed_at,
    DROP COLUMN IF EXISTS analysis_notes,
    DROP COLUMN IF EXISTS analysis_status,
    DROP COLUMN IF EXISTS quality_score,
    DROP COLUMN IF EXISTS source_coverage,
    DROP COLUMN IF EXISTS flavour_molecule_count,
    DROP COLUMN IF EXISTS natural_source,
    DROP COLUMN IF EXISTS category;
