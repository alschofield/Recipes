-- 003_seed_ingredients.up.sql
-- Common cooking ingredients for development/testing.

-- Proteins
INSERT INTO ingredients (canonical_name) VALUES ('chicken') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('beef') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('pork') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('salmon') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('shrimp') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('tofu') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('egg') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('ground turkey') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('lamb') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('bacon') ON CONFLICT DO NOTHING;

-- Grains / Starches
INSERT INTO ingredients (canonical_name) VALUES ('rice') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('pasta') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('bread') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('potato') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('noodle') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('flour') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('corn') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('tortilla') ON CONFLICT DO NOTHING;

-- Vegetables
INSERT INTO ingredients (canonical_name) VALUES ('onion') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('garlic') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('tomato') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('broccoli') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('spinach') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('carrot') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('bell pepper') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('green onion') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('mushroom') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('zucchini') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('celery') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('cabbage') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('avocado') ON CONFLICT DO NOTHING;

-- Fruits
INSERT INTO ingredients (canonical_name) VALUES ('lemon') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('lime') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('apple') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('banana') ON CONFLICT DO NOTHING;

-- Dairy
INSERT INTO ingredients (canonical_name) VALUES ('milk') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('butter') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('cheese') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('cream') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('sour cream') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('yogurt') ON CONFLICT DO NOTHING;

-- Spices / Condiments
INSERT INTO ingredients (canonical_name) VALUES ('salt') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('pepper') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('olive oil') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('soy sauce') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('vinegar') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('sugar') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('honey') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('mustard') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('ketchup') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('mayonnaise') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('cumin') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('paprika') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('oregano') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('basil') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('ginger') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('chili') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('curry powder') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('bay leaf') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('cinnamon') ON CONFLICT DO NOTHING;

-- Canned / Pantry
INSERT INTO ingredients (canonical_name) VALUES ('chicken broth') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('beef broth') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('tomato sauce') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('coconut milk') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('canned tomato') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('beans') ON CONFLICT DO NOTHING;
INSERT INTO ingredients (canonical_name) VALUES ('lentil') ON CONFLICT DO NOTHING;

-- Aliases (maps alternate names to canonical ingredients)
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'scallion' FROM ingredients WHERE canonical_name = 'green onion' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'spring onion' FROM ingredients WHERE canonical_name = 'green onion' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'chives' FROM ingredients WHERE canonical_name = 'green onion' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'coriander' FROM ingredients WHERE canonical_name = 'cilantro' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'aubergine' FROM ingredients WHERE canonical_name = 'eggplant' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'courgette' FROM ingredients WHERE canonical_name = 'zucchini' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'bell pepper' FROM ingredients WHERE canonical_name = 'pepper' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'capsicum' FROM ingredients WHERE canonical_name = 'bell pepper' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'chili pepper' FROM ingredients WHERE canonical_name = 'chili' ON CONFLICT DO NOTHING;
INSERT INTO ingredient_aliases (ingredient_id, alias)
    SELECT id, 'dried chili' FROM ingredients WHERE canonical_name = 'chili' ON CONFLICT DO NOTHING;
