-- 003_seed_ingredients.down.sql
-- Revert: remove seeded ingredients.

DELETE FROM ingredient_aliases WHERE ingredient_id IN (
    SELECT id FROM ingredients WHERE canonical_name IN (
        'chicken', 'beef', 'pork', 'salmon', 'shrimp', 'tofu', 'egg',
        'ground turkey', 'lamb', 'bacon', 'rice', 'pasta', 'bread',
        'potato', 'noodle', 'flour', 'corn', 'tortilla', 'onion',
        'garlic', 'tomato', 'broccoli', 'spinach', 'carrot', 'bell pepper',
        'green onion', 'mushroom', 'zucchini', 'celery', 'cabbage',
        'avocado', 'lemon', 'lime', 'apple', 'banana', 'milk', 'butter',
        'cheese', 'cream', 'sour cream', 'yogurt', 'salt', 'pepper',
        'olive oil', 'soy sauce', 'vinegar', 'sugar', 'honey', 'mustard',
        'ketchup', 'mayonnaise', 'cumin', 'paprika', 'oregano', 'basil',
        'ginger', 'chili', 'curry powder', 'bay leaf', 'cinnamon',
        'chicken broth', 'beef broth', 'tomato sauce', 'coconut milk',
        'canned tomato', 'beans', 'lentil'
    )
);

DELETE FROM ingredients WHERE canonical_name IN (
    'chicken', 'beef', 'pork', 'salmon', 'shrimp', 'tofu', 'egg',
    'ground turkey', 'lamb', 'bacon', 'rice', 'pasta', 'bread',
    'potato', 'noodle', 'flour', 'corn', 'tortilla', 'onion',
    'garlic', 'tomato', 'broccoli', 'spinach', 'carrot', 'bell pepper',
    'green onion', 'mushroom', 'zucchini', 'celery', 'cabbage',
    'avocado', 'lemon', 'lime', 'apple', 'banana', 'milk', 'butter',
    'cheese', 'cream', 'sour cream', 'yogurt', 'salt', 'pepper',
    'olive oil', 'soy sauce', 'vinegar', 'sugar', 'honey', 'mustard',
    'ketchup', 'mayonnaise', 'cumin', 'paprika', 'oregano', 'basil',
    'ginger', 'chili', 'curry powder', 'bay leaf', 'cinnamon',
    'chicken broth', 'beef broth', 'tomato sauce', 'coconut milk',
    'canned tomato', 'beans', 'lentil'
);
