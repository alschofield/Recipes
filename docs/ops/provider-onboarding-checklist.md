# Provider Onboarding Checklist (Option B, Cheapest-First)

This guide gives you a concrete setup path for:

- Web: Vercel
- API services: Railway (or similar container host)
- Postgres: Neon
- Redis: Upstash
- Single API domain gateway: Cloudflare Workers

You can swap providers later without changing app architecture.

## 0) Account creation order

Create accounts in this order:

1. GitHub (already used)
2. Vercel (web)
3. Railway (API services)
4. Neon (Postgres)
5. Upstash (Redis)
6. Cloudflare (DNS + Worker gateway)
7. Domain registrar (if separate from Cloudflare)

## 1) Provision managed data stores

## Neon (Postgres)

- Create database: `recipes`
- Create application user with least privileges
- Copy connection values for:
  - `POSTGRES_USER`
  - `POSTGRES_PASSWORD`
  - `POSTGRES_HOST`
  - `POSTGRES_PORT`
  - `POSTGRES_DB`

## Upstash (Redis)

- Create Redis database in same region as API services
- Copy TLS URL (recommended):
  - `REDIS_URL=rediss://...`

## 2) Deploy API services (Railway)

Create 3 services from this same repository:

- `recipes-server` (Dockerfile: `server/Dockerfile.recipes`)
- `users-server` (Dockerfile: `server/Dockerfile.users`)
- `favorites-server` (Dockerfile: `server/Dockerfile.favorites`)

Set these env vars on all 3 services (as applicable):

```env
POSTGRES_USER=app_user
POSTGRES_PASSWORD=<strong-secret>
POSTGRES_HOST=<neon-host>
POSTGRES_PORT=5432
POSTGRES_DB=recipes

JWT_SECRET=<64+ char random secret>
JWT_ISSUER=recipes-users-server
JWT_ACCESS_TTL=15m

REDIS_URL=rediss://<upstash-url>

APP_ENV=production
CORS_ALLOWED_ORIGINS=https://www.yourdomain.com,https://yourdomain.com
MAX_BODY_BYTES=1048576
RATE_LIMIT_RPS=50
RATE_LIMIT_BURST=100
ERROR_WEBHOOK_URL=
```

Service-specific ports:

- `RECIPES_SERVER_PORT=8081`
- `USERS_SERVER_PORT=8082`
- `FAVORITES_SERVER_PORT=8080`

After deploy, copy service URLs:

- `RECIPES_ORIGIN=https://...`
- `USERS_ORIGIN=https://...`
- `FAVORITES_ORIGIN=https://...`

## 3) Configure API gateway (Cloudflare Worker)

In `server/gateway/cloudflare-worker/wrangler.toml` (from example), set:

```toml
[vars]
RECIPES_ORIGIN = "https://recipes-service.example.com"
USERS_ORIGIN = "https://users-service.example.com"
FAVORITES_ORIGIN = "https://favorites-service.example.com"
```

Deploy Worker and map custom domain route:

- `api.yourdomain.com/*` -> Worker

Verify:

- `https://api.yourdomain.com/recipes/health`
- `https://api.yourdomain.com/users/health`
- `https://api.yourdomain.com/favorites/health`

## 4) Deploy web app (Vercel)

Connect GitHub repo and set root to `web/`.

Set production env:

```env
NEXT_PUBLIC_API_BASE_URL=https://api.yourdomain.com
NEXT_PUBLIC_ERROR_WEBHOOK_URL=
```

Optional fallback vars (not needed when using `API_BASE_URL`):

- `NEXT_PUBLIC_API_RECIPES_URL`
- `NEXT_PUBLIC_API_USERS_URL`
- `NEXT_PUBLIC_API_FAVORITES_URL`

Attach custom web domain:

- `yourdomain.com`
- `www.yourdomain.com`

## 5) GitHub Actions deployment hooks

Set these GitHub secrets (staging/prod as available):

```text
WEB_STAGING_DEPLOY_HOOK
API_RECIPES_STAGING_DEPLOY_HOOK
API_USERS_STAGING_DEPLOY_HOOK
API_FAVORITES_STAGING_DEPLOY_HOOK

WEB_PROD_DEPLOY_HOOK
API_RECIPES_PROD_DEPLOY_HOOK
API_USERS_PROD_DEPLOY_HOOK
API_FAVORITES_PROD_DEPLOY_HOOK
```

These power:

- `.github/workflows/deploy-staging.yml`
- `.github/workflows/deploy-prod.yml`

## 6) First production readiness checks

Run smoke checklist after first deploy:

1. Web health: `/api/health`
2. API health endpoints via `api.yourdomain.com`
3. Signup + login
4. Recipe search
5. Add/remove favorite
6. Recipe detail page

## 7) Cost control tips

- Start on smallest paid/free tiers with scale-to-zero where acceptable.
- Keep all services in same region to reduce latency and egress surprises.
- Use one API domain gateway on Cloudflare Worker free tier first.
- Add alerting before increasing scale (avoid blind overprovisioning).
