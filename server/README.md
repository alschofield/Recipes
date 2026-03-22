# Server (Go Services)

This folder contains three Go HTTP services and supporting backend assets.

## Services

- `cmd/recipes-server` - recipe search, recipe details, ingredient suggestion/moderation
- `cmd/users-server` - signup/login/profile management
- `cmd/favorites-server` - user favorites CRUD
- `cmd/sample-data` - seed utility for local data population
- `cmd/ingredient-seed` - canonical ingredient enrichment seed loader

## Folder structure

- `pkg/` - domain/application packages
- `migrations/` - SQL migration files
- `lib/` - bundled datasets used by seeding/search
- `lib/derived/` - generated consolidated seed artifacts (`canonical_ingredient_seed_v1.*`)
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
go test ./server/...
go test ./server/pkg/search -run TestSearch
go test ./server/pkg/middleware -run TestRequireAuth
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
- Cache: `REDIS_URL`
- Network: `CORS_ALLOWED_ORIGINS`
- Ops: `ERROR_WEBHOOK_URL`
- Ports: `RECIPES_SERVER_PORT`, `USERS_SERVER_PORT`, `FAVORITES_SERVER_PORT`

## Production notes

- Container images are built with multi-stage Dockerfiles and run as non-root users.
- Set strong, non-default values for `JWT_SECRET` in staging/prod.
- Lock `CORS_ALLOWED_ORIGINS` to deployed web origins.
- Keep migrations in release flow before service rollout.

## Operational docs

- Architecture: [`../docs/architecture.md`](../docs/architecture.md)
- Security/Auth baseline: [`../docs/auth-security-baseline.md`](../docs/auth-security-baseline.md)
- Operations runbook: [`../docs/operations-runbook.md`](../docs/operations-runbook.md)
