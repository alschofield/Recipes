SHELL := /usr/bin/bash

.DEFAULT_GOAL := help

# --- Environment defaults (override at invocation time if needed) ---
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= changeme
POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_DB ?= recipes

JWT_SECRET ?= changeme
JWT_ISSUER ?= recipes-users-server
REDIS_URL ?= redis://localhost:6379
CORS_ALLOWED_ORIGINS ?= http://localhost:3000,http://localhost:8081

RECIPES_SERVER_PORT ?= 8081
USERS_SERVER_PORT ?= 8082
FAVORITES_SERVER_PORT ?= 8080

API_URL ?= http://localhost
API_BASE_URL ?=
API_RECIPES_PORT ?= 8081
API_USERS_PORT ?= 8082
API_FAVORITES_PORT ?= 8080

DB_HOST_FOR_MIGRATE ?= host.docker.internal
MIGRATIONS_DIR := $(CURDIR)/server/migrations
MIGRATE_DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(DB_HOST_FOR_MIGRATE):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: help setup check infra-up infra-down infra-ps infra-logs infra-restart \
        migrate-up migrate-down migrate-version seed seed-ingredients db-counts \
        server-run-recipes server-run-users server-run-favorites \
        server-test server-test-search server-test-auth \
        web-install web-dev web-build web-start web-lint web-test web-e2e \
        health

help:
	@printf "\nRecipes Monorepo Commands\n\n"
	@printf "Setup\n"
	@printf "  make setup              Install web deps + start infra + migrate + seed\n"
	@printf "  make check              Build web + run backend/web tests\n\n"
	@printf "Infrastructure\n"
	@printf "  make infra-up           Start postgres + redis\n"
	@printf "  make infra-down         Stop all compose services\n"
	@printf "  make infra-ps           Show compose service status\n"
	@printf "  make infra-logs         Follow compose logs\n"
	@printf "  make infra-restart      Restart postgres + redis\n\n"
	@printf "Database\n"
	@printf "  make migrate-up         Apply SQL migrations\n"
	@printf "  make migrate-down       Roll back one migration\n"
	@printf "  make migrate-version    Show migration version\n"
	@printf "  make seed               Seed sample data\n"
	@printf "  make db-counts          Show basic DB row counts\n\n"
	@printf "Backend (Go)\n"
	@printf "  make server-run-recipes Run recipes API (:8081)\n"
	@printf "  make server-run-users   Run users API (:8082)\n"
	@printf "  make server-run-favorites Run favorites API (:8080)\n"
	@printf "  make server-test        Run all backend tests\n"
	@printf "  make server-test-search Run search package tests\n"
	@printf "  make server-test-auth   Run middleware auth tests\n\n"
	@printf "Web (Next.js)\n"
	@printf "  make web-install        Install web dependencies\n"
	@printf "  make web-dev            Start Next dev server (Turbopack)\n"
	@printf "  make web-build          Build web app\n"
	@printf "  make web-start          Start production web server\n"
	@printf "  make web-lint           Lint web app\n"
	@printf "  make web-test           Run unit tests\n"
	@printf "  make web-e2e            Run Playwright suite\n\n"

setup: web-install infra-up migrate-up seed

check: web-build server-test web-test

infra-up:
	docker compose up -d postgres redis

infra-down:
	docker compose down

infra-ps:
	docker compose ps

infra-logs:
	docker compose logs -f

infra-restart:
	docker compose restart postgres redis

migrate-up:
	MSYS_NO_PATHCONV=1 docker run --rm -v "$(MIGRATIONS_DIR):/migrations" migrate/migrate -path=/migrations -database "$(MIGRATE_DB_URL)" up

migrate-down:
	MSYS_NO_PATHCONV=1 docker run --rm -v "$(MIGRATIONS_DIR):/migrations" migrate/migrate -path=/migrations -database "$(MIGRATE_DB_URL)" down 1

migrate-version:
	MSYS_NO_PATHCONV=1 docker run --rm -v "$(MIGRATIONS_DIR):/migrations" migrate/migrate -path=/migrations -database "$(MIGRATE_DB_URL)" version

seed: seed-ingredients
	POSTGRES_USER=$(POSTGRES_USER) POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) POSTGRES_HOST=$(POSTGRES_HOST) POSTGRES_PORT=$(POSTGRES_PORT) POSTGRES_DB=$(POSTGRES_DB) go run server/cmd/sample-data/main.go

seed-ingredients:
	POSTGRES_USER=$(POSTGRES_USER) POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) POSTGRES_HOST=$(POSTGRES_HOST) POSTGRES_PORT=$(POSTGRES_PORT) POSTGRES_DB=$(POSTGRES_DB) CANONICAL_INGREDIENT_SEED=$(CANONICAL_INGREDIENT_SEED) go run server/cmd/ingredient-seed/main.go

db-counts:
	docker exec recipes_postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c "SELECT COUNT(*) AS ingredients FROM ingredients; SELECT COUNT(*) AS recipes FROM recipes; SELECT COUNT(*) AS aliases FROM ingredient_aliases;"

server-run-recipes:
	POSTGRES_USER=$(POSTGRES_USER) POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) POSTGRES_HOST=$(POSTGRES_HOST) POSTGRES_PORT=$(POSTGRES_PORT) POSTGRES_DB=$(POSTGRES_DB) JWT_SECRET=$(JWT_SECRET) JWT_ISSUER=$(JWT_ISSUER) REDIS_URL=$(REDIS_URL) CORS_ALLOWED_ORIGINS=$(CORS_ALLOWED_ORIGINS) RECIPES_SERVER_PORT=$(RECIPES_SERVER_PORT) go run server/cmd/recipes-server/main.go

server-run-users:
	POSTGRES_USER=$(POSTGRES_USER) POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) POSTGRES_HOST=$(POSTGRES_HOST) POSTGRES_PORT=$(POSTGRES_PORT) POSTGRES_DB=$(POSTGRES_DB) JWT_SECRET=$(JWT_SECRET) JWT_ISSUER=$(JWT_ISSUER) REDIS_URL=$(REDIS_URL) CORS_ALLOWED_ORIGINS=$(CORS_ALLOWED_ORIGINS) USERS_SERVER_PORT=$(USERS_SERVER_PORT) go run server/cmd/users-server/main.go

server-run-favorites:
	POSTGRES_USER=$(POSTGRES_USER) POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) POSTGRES_HOST=$(POSTGRES_HOST) POSTGRES_PORT=$(POSTGRES_PORT) POSTGRES_DB=$(POSTGRES_DB) JWT_SECRET=$(JWT_SECRET) JWT_ISSUER=$(JWT_ISSUER) REDIS_URL=$(REDIS_URL) CORS_ALLOWED_ORIGINS=$(CORS_ALLOWED_ORIGINS) FAVORITES_SERVER_PORT=$(FAVORITES_SERVER_PORT) go run server/cmd/favorites-server/main.go

server-test:
	(cd server && go test ./...)

server-test-search:
	(cd server && go test ./pkg/search -run TestSearch)

server-test-auth:
	(cd server && go test ./pkg/middleware -run TestRequireAuth)

web-install:
	pnpm --dir web install

web-dev:
	NEXT_PUBLIC_API_BASE_URL=$(API_BASE_URL) NEXT_PUBLIC_API_URL=$(API_URL) NEXT_PUBLIC_API_RECIPES_PORT=$(API_RECIPES_PORT) NEXT_PUBLIC_API_USERS_PORT=$(API_USERS_PORT) NEXT_PUBLIC_API_FAVORITES_PORT=$(API_FAVORITES_PORT) pnpm --dir web dev

web-build:
	NEXT_PUBLIC_API_BASE_URL=$(API_BASE_URL) NEXT_PUBLIC_API_URL=$(API_URL) NEXT_PUBLIC_API_RECIPES_PORT=$(API_RECIPES_PORT) NEXT_PUBLIC_API_USERS_PORT=$(API_USERS_PORT) NEXT_PUBLIC_API_FAVORITES_PORT=$(API_FAVORITES_PORT) pnpm --dir web build

web-start:
	pnpm --dir web start

web-lint:
	pnpm --dir web lint

web-test:
	pnpm --dir web test

web-e2e:
	pnpm --dir web test:e2e

health:
	@curl -fsS http://localhost:8081/recipes/health && printf "\n"
	@curl -fsS http://localhost:8082/users/health && printf "\n"
	@curl -fsS http://localhost:8080/favorites/health && printf "\n"
