# Recipes App Architecture Overview

## Repository Structure

```
Recipes/
├── docs/                  # Project decisions (search, auth, LLM contracts)
├── server/                # Go backend
│   ├── cmd/               # Entry points per service
│   │   ├── recipes-server/   (port 8081)
│   │   ├── users-server/     (port 8082)
│   │   ├── favorites-server/ (port 8080)
│   │   └── sample-data/      (seed data utility)
│   ├── pkg/
│   │   ├── add/           # Favorites add logic
│   │   ├── create/         # Recipe/user creation logic
│   │   ├── delete/         # User/favorite deletion logic
│   │   ├── edit/           # User edit logic
│   │   ├── search/         # Recipe/user/favorite search logic
│   │   └── storage/        # DB connection helpers
│   ├── migrations/         # SQL migration files (golang-migrate)
│   ├── etc/nginx/         # Nginx reverse proxy config
│   ├── scripts/db/        # DB backup and restore helpers
│   ├── lib/               # Raw datasets (ingredients, recipes)
│   ├── Dockerfile.*        # Per-service Dockerfiles
│   ├── go.mod / go.sum
│   └── .env.example
├── web/                   # Next.js frontend
│   ├── app/               # Next.js App Router
│   ├── styles/            # Global styles
│   ├── next.config.js
│   ├── .env.example
│   └── package.json
├── docker-compose.yml     # Full stack (postgres, redis, all 3 servers, nginx)
├── .github/workflows/     # CI pipeline
│   └── ci.yml
├── CHECKLIST.md           # Kanban task board
└── README.md
```

## Architecture Decisions

- **3 separate Go microservices** behind an nginx reverse proxy on port 80.
  - `/recipes/*` → recipes-server (8081)
  - `/users/*`   → users-server (8082)
  - `/favorites/*` → favorites-server (8080)
- **Single Postgres database** (`recipes`) shared by all services.
  - Rationale: simpler ops, avoids cross-service joins, fits MVP scale.
  - Each service uses its own connection pool (`pgx/v5`).
- **Redis** available for caching (recipe search results, LLM responses).
- **Next.js frontend** proxies to nginx; nginx routes to correct service.
- **LLM generation** handled by recipes-server (calls external API).
- **Auth tokens** (JWT) issued by users-server; validated by each service.

## Domain Model

- Go models: `server/pkg/models/models.go` — User, Ingredient, Recipe, RecipeIngredient, Favorite
- Repository interfaces: `server/pkg/repository/repository.go` — UserRepo, RecipeRepo, FavoriteRepo, IngredientRepo
- Shared pool: `server/pkg/storage/postgres/pool.go` — single connection pool via `storage.Pool()`
- See `docs/product/domain-language.md` for entity/service definitions, events, and language rules.

## Known Issues / TODOs

- Handler packages (`pkg/create/`, `pkg/search/`, `pkg/delete/`, `pkg/edit/`) are stubs — implement in Groups C/D/F.
- Recipe search ranking algorithm not yet implemented.
- Ingredient normalization and alias mapping not yet fully implemented in code.

## Development Quick Start

```bash
# 1. Copy env and fill in values
cp server/.env.example server/.env
cp web/.env.example web/.env.local

# 2. Start all services
docker compose up -d postgres redis

# 3. Run migrations
cd server
migrate -path ./migrations -database "$DATABASE_URL" up

# 4. Start servers (separate terminals)
go run cmd/recipes-server/main.go
go run cmd/users-server/main.go
go run cmd/favorites-server/main.go

# 5. Start frontend
cd ../web && npm install && npm run dev
```
