-- 004_seed_curated_recipes.up.sql
-- Curated baseline recipes for DB-first search results.

-- Ensure a few extra canonical ingredients exist.
INSERT INTO ingredients (canonical_name) VALUES ('cilantro') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('eggplant') ON CONFLICT DO NOTHING;

-- Seed curated recipes.
INSERT INTO recipes (
  name, description, steps, source, quality_score,
  prep_minutes, cook_minutes, difficulty, cuisine,
  servings, dietary_tags, safety_notes, reviewable
) VALUES
(
  'Garlic Chicken Rice Bowl',
  'Savory chicken rice bowl with garlic and green onion.',
  '["Season chicken", "Saute garlic", "Cook chicken", "Serve over rice"]'::jsonb,
  'database',
  0.90,
  15,
  20,
  'easy',
  'asian',
  2,
  ARRAY['high-protein'],
  ARRAY['Cook chicken to 165F'],
  FALSE
),
(
  'Veggie Pasta Primavera',
  'Colorful pasta with tomato, zucchini, and bell pepper.',
  '["Boil pasta", "Saute vegetables", "Combine and season"]'::jsonb,
  'database',
  0.86,
  20,
  15,
  'easy',
  'italian',
  3,
  ARRAY['vegetarian'],
  ARRAY['Handle hot pasta water carefully'],
  FALSE
),
(
  'Tofu Curry Skillet',
  'Quick curry skillet with tofu, coconut milk, and aromatics.',
  '["Brown tofu", "Cook aromatics", "Simmer with coconut milk and curry powder"]'::jsonb,
  'database',
  0.88,
  15,
  20,
  'medium',
  'indian',
  2,
  ARRAY['vegetarian', 'dairy-free'],
  ARRAY['Simmer until fully heated'],
  FALSE
)
ON CONFLICT DO NOTHING;

-- Recipe ingredient links.
INSERT INTO recipe_ingredients (recipe_id, ingredient_id, amount, unit, optional, position)
SELECT r.id, i.id, x.amount, x.unit, x.optional, x.position
FROM (
  VALUES
    ('Garlic Chicken Rice Bowl', 'chicken', '400', 'g', FALSE, 1),
    ('Garlic Chicken Rice Bowl', 'rice', '1', 'cup', FALSE, 2),
    ('Garlic Chicken Rice Bowl', 'garlic', '3', 'cloves', FALSE, 3),
    ('Garlic Chicken Rice Bowl', 'green onion', '2', 'stalks', TRUE, 4),
    ('Garlic Chicken Rice Bowl', 'soy sauce', '2', 'tbsp', TRUE, 5),

    ('Veggie Pasta Primavera', 'pasta', '300', 'g', FALSE, 1),
    ('Veggie Pasta Primavera', 'tomato', '2', 'count', FALSE, 2),
    ('Veggie Pasta Primavera', 'zucchini', '1', 'count', FALSE, 3),
    ('Veggie Pasta Primavera', 'bell pepper', '1', 'count', FALSE, 4),
    ('Veggie Pasta Primavera', 'olive oil', '1', 'tbsp', TRUE, 5),
    ('Veggie Pasta Primavera', 'basil', '1', 'tbsp', TRUE, 6),

    ('Tofu Curry Skillet', 'tofu', '400', 'g', FALSE, 1),
    ('Tofu Curry Skillet', 'coconut milk', '1', 'can', FALSE, 2),
    ('Tofu Curry Skillet', 'curry powder', '1', 'tbsp', FALSE, 3),
    ('Tofu Curry Skillet', 'onion', '1', 'count', FALSE, 4),
    ('Tofu Curry Skillet', 'garlic', '2', 'cloves', FALSE, 5),
    ('Tofu Curry Skillet', 'ginger', '1', 'tbsp', TRUE, 6)
) AS x(recipe_name, ingredient_name, amount, unit, optional, position)
JOIN recipes r ON r.name = x.recipe_name
JOIN ingredients i ON i.canonical_name = x.ingredient_name
WHERE NOT EXISTS (
  SELECT 1
  FROM recipe_ingredients ri
  WHERE ri.recipe_id = r.id
    AND ri.ingredient_id = i.id
);
