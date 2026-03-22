-- Run all pending migrations (in order)
-- Requires golang-migrate CLI: https://github.com/golang-migrate/migrate

# From the server directory:
migrate -path ./migrations -database "postgres://postgres:changeme@localhost:5432/recipes?sslmode=disable" up

# Create a new migration:
migrate create -ext sql -dir ./migrations -seq add_cuisine_field

# Rollback last migration:
migrate -path ./migrations -database "postgres://postgres:changeme@localhost:5432/recipes?sslmode=disable" down 1

# Check current version:
migrate -path ./migrations -database "postgres://postgres:changeme@localhost:5432/recipes?sslmode=disable" version
