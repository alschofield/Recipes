-- 006_backfill_canonical_aliases.up.sql
-- Ensure canonical ingredient names are always valid aliases for search matching.

INSERT INTO ingredient_aliases (ingredient_id, alias)
SELECT id, canonical_name
FROM ingredients
ON CONFLICT (alias) DO NOTHING;
