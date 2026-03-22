-- 007_ingredient_attributes_and_analysis.up.sql
-- Adds ingredient enrichment fields and recipe-level analysis storage.

ALTER TABLE ingredients
    ADD COLUMN IF NOT EXISTS category VARCHAR(80),
    ADD COLUMN IF NOT EXISTS natural_source VARCHAR(120),
    ADD COLUMN IF NOT EXISTS flavour_molecule_count INTEGER,
    ADD COLUMN IF NOT EXISTS source_coverage SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS quality_score NUMERIC(4, 3) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS analysis_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (analysis_status IN ('pending', 'enriched', 'review_required')),
    ADD COLUMN IF NOT EXISTS analysis_notes TEXT,
    ADD COLUMN IF NOT EXISTS last_analyzed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_ingredients_analysis_status ON ingredients(analysis_status);
CREATE INDEX IF NOT EXISTS idx_ingredients_quality_score ON ingredients(quality_score DESC);
CREATE INDEX IF NOT EXISTS idx_ingredients_source_coverage ON ingredients(source_coverage DESC);

CREATE TABLE IF NOT EXISTS ingredient_nutrition_profiles (
    ingredient_id UUID PRIMARY KEY REFERENCES ingredients(id) ON DELETE CASCADE,
    calories_kcal NUMERIC(10, 3),
    protein_g NUMERIC(10, 3),
    fat_g NUMERIC(10, 3),
    carbs_g NUMERIC(10, 3),
    fiber_g NUMERIC(10, 3),
    sodium_mg NUMERIC(10, 3),
    source VARCHAR(30) NOT NULL DEFAULT 'usda',
    confidence NUMERIC(4, 3) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recipe_quality_analysis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id UUID NOT NULL UNIQUE REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_coverage_score NUMERIC(4, 3) NOT NULL DEFAULT 0,
    nutrition_balance_score NUMERIC(4, 3) NOT NULL DEFAULT 0,
    flavour_alignment_score NUMERIC(4, 3) NOT NULL DEFAULT 0,
    novelty_score NUMERIC(4, 3) NOT NULL DEFAULT 0,
    overall_score NUMERIC(4, 3) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'computed', 'failed')),
    notes TEXT,
    computed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recipe_quality_analysis_status ON recipe_quality_analysis(status);
CREATE INDEX IF NOT EXISTS idx_recipe_quality_analysis_overall ON recipe_quality_analysis(overall_score DESC);

CREATE OR REPLACE TRIGGER ingredient_nutrition_profiles_updated_at
    BEFORE UPDATE ON ingredient_nutrition_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE OR REPLACE TRIGGER recipe_quality_analysis_updated_at
    BEFORE UPDATE ON recipe_quality_analysis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
