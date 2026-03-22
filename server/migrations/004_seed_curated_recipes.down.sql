-- 004_seed_curated_recipes.down.sql
-- Remove curated baseline recipes.

DELETE FROM recipes
WHERE name IN (
  'Garlic Chicken Rice Bowl',
  'Veggie Pasta Primavera',
  'Tofu Curry Skillet'
);
