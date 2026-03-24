# Server (Go Services)

This folder contains three Go HTTP services and supporting backend assets.

## Services

- `cmd/recipes-server` - recipe search, recipe details, ingredient suggestion/moderation
- `cmd/users-server` - signup/login/profile management
- `cmd/favorites-server` - user favorites CRUD
- `cmd/sample-data` - seed utility for local data population
- `cmd/ingredient-seed` - canonical ingredient enrichment seed loader
- `cmd/search-load` - endpoint load baseline runner for `/recipes/search`

## Folder structure

- `pkg/` - domain/application packages
- `migrations/` - SQL migration files
- `../datasets/raw/server-lib/` - bundled source datasets used by seeding/search
- `../datasets/derived/server-lib/` - generated consolidated seed artifacts (`canonical_ingredient_seed_v1.*`)
- `etc/nginx/` - nginx local/prod routing config
- `gateway/cloudflare-worker/` - single-domain API gateway option for split hosting
- `scripts/db/` - backup/restore helper scripts
- `Dockerfile.*` - per-service container builds

## Run from repo root

Preferred:

```bash
make server-run-recipes
make server-run-users
make server-run-favorites
```

Or:

```bash
task server-run-recipes
task server-run-users
task server-run-favorites
```

Direct commands:

```bash
go run server/cmd/recipes-server/main.go
go run server/cmd/users-server/main.go
go run server/cmd/favorites-server/main.go
go run server/cmd/search-load/main.go -scenario fallback-heavy -requests 300 -concurrency 12
```

## Tests

From repo root:

```bash
make server-test
make server-test-search
make server-test-auth
```

Or:

```bash
(cd server && go test ./...)
(cd server && go test ./pkg/search -run TestSearch)
(cd server && go test ./pkg/middleware -run TestRequireAuth)
```

## Database and migrations

Use root runner targets:

```bash
make migrate-up
make seed-ingredients
make seed
make db-counts
```

Task equivalents:

```bash
task migrate-up
task seed-ingredients
task seed
task db-counts
```

## Environment variables

See `server/.env.example` for full list.
Production template: `server/.env.production.example`.

Most important:

- DB: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`
- Auth: `JWT_SECRET`, `JWT_ISSUER`, `JWT_ACCESS_TTL`
- Auth refresh: `JWT_REFRESH_TTL`
- Cache: `REDIS_URL`
- Search blend controls: `SEARCH_BLEND_MIN_GENERATED`, `SEARCH_BLEND_MAX_GENERATED_SHARE`, `SEARCH_BLEND_SEED`
- Network: `CORS_ALLOWED_ORIGINS`
- Ops: `ERROR_WEBHOOK_URL`
- Idempotency: `IDEMPOTENCY_KEY_TTL` (for `Idempotency-Key` replay window)
- Ingredient governance: `INGREDIENT_POLICY_MODE` (`auto_create` or `queue_only`)
- Ingredient queue SLA: `INGREDIENT_CANDIDATE_SLA_HOURS`
- LLM fallback controls: `LLM_FALLBACK_DISABLED`, `LLM_FALLBACK_CANARY_PERCENT`
- Strict-mode fallback policy: `LLM_STRICT_GENERATED_POLICY` (`none` or `degrade_inclusive`)
- LLM alerts: `LLM_ALERT_TIMEOUT_RATE_THRESHOLD`, `LLM_ALERT_SCHEMA_ERROR_RATE_THRESHOLD`, `LLM_ALERT_REPAIR_FAIL_RATE_THRESHOLD`
- Ports: `RECIPES_SERVER_PORT`, `USERS_SERVER_PORT`, `FAVORITES_SERVER_PORT`

## Production notes

- Container images are built with multi-stage Dockerfiles and run as non-root users.
- Set strong, non-default values for `JWT_SECRET` in staging/prod.
- Lock `CORS_ALLOWED_ORIGINS` to deployed web origins.
- Keep migrations in release flow before service rollout.

## Operational docs

- Architecture: [`../docs/server/architecture.md`](../docs/server/architecture.md)
- Security/Auth baseline: [`../docs/server/auth-security-baseline.md`](../docs/server/auth-security-baseline.md)
- Operations runbook: [`../docs/ops/operations-runbook.md`](../docs/ops/operations-runbook.md)
