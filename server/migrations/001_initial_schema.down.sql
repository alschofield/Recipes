-- 001_initial_schema.down.sql
-- Revert: drop all tables and functions created in 001_initial_schema.up.sql

DROP TRIGGER IF EXISTS recipes_updated_at ON recipes;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at();

DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS recipe_ingredients;
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS ingredient_aliases;
DROP TABLE IF EXISTS ingredients;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
