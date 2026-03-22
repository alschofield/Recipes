# Deployment Plan (Web + APIs)

This plan focuses on:

1. Production readiness for current server/web code
2. Lowest-cost path to first public release
3. Clean migration path to custom domains and stronger scaling

## Current readiness baseline

Completed hardening in this repo:

- Backend images run as non-root (`distroless` runtime)
- Frontend image uses standalone Next output and non-root user
- Docker contexts reduced with `.dockerignore`
- Security headers enabled in backend and web
- Web health endpoint available at `/api/health`
- Root runbooks and command runners (`Makefile`, `Taskfile.yml`) in place

## Compose and Dockerfile guidance

- Keep `docker-compose.yml` at repo root because it orchestrates multiple top-level services (`server`, `web`, DB, Redis, nginx) for local development.
- For Option B split-provider production, do **not** rely on Docker Compose as the production orchestrator.
- Reuse current Dockerfiles for production image builds in CI; deploy those images to managed app platforms.
- If desired, add a `docker-compose.dev.yml` later for local overrides, but avoid a `docker-compose.prod.yml` unless you self-host orchestration.

## Environment model

- `dev`: local Docker + local binaries
- `staging`: production-like infra with lower scale
- `prod`: managed DB/cache, strict secrets, monitoring/alerts, rollbacks

## Recommended low-cost hosting path

### Phase 1 (launch quickly, low cost)

- **Web:** Vercel Hobby (or equivalent) with SSR support
- **API:** lowest-cost container host (for example Railway/Fly/Render depending on current pricing/region)
- **Postgres:** managed low-cost tier (for example Neon/Supabase/Railway)
- **Redis:** managed low-cost tier (for example Upstash/Redis Cloud)

Notes:

- "Free" API tiers change frequently; use free/trial only for non-critical traffic.
- Keep prod-like credentials and backups even on low-cost tiers.

### Phase 2 (custom domain + reliability)

- Point web custom domain to hosted frontend (`www.yourdomain.com`)
- Point API to subdomain (`api.yourdomain.com`)
- Enforce HTTPS everywhere
- Add uptime checks and alerting (health endpoints + error webhooks)

### Phase 3 (scale and resilience)

- Separate deploys for users/recipes/favorites APIs
- Horizontal scaling where needed
- Connection pooling and DB sizing reviews
- Canary/blue-green deploy strategy

## CI/CD automation checklist

## 1) Build & test gate (already partly present)

- [ ] Server: `go vet`, `go test`, `go build`
- [ ] Web: `pnpm install --frozen-lockfile`, `pnpm lint`, `pnpm test`, `pnpm build`
- [ ] Optional smoke e2e on release branches

## 2) Image build and push

- [ ] Build immutable, versioned images for each API service
- [ ] Build immutable, versioned image for web
- [ ] Push images to registry with commit SHA + semver tags

## 3) Migration step (single-run job)

- [ ] Run migrations once per environment before deploy
- [ ] Fail release if migrations fail
- [ ] Record migration version in deployment logs

## 4) Deploy step

- [ ] Deploy API services with health checks
- [ ] Deploy web service after APIs are healthy
- [ ] Validate health endpoints (`/recipes/health`, `/users/health`, `/favorites/health`, `/api/health`)

Current repo automation uses deploy hook-based workflows:

- `.github/workflows/deploy-staging.yml`
- `.github/workflows/deploy-prod.yml`

Provide per-environment hook secrets in GitHub before enabling those workflows.

## 5) Post-deploy smoke tests

- [ ] Signup/login
- [ ] Recipe search
- [ ] Add/remove favorite
- [ ] Recipe detail view

## 6) Rollback plan

- [ ] Keep last known good image tags
- [ ] Roll back web and API independently
- [ ] If migration-related failure, restore DB backup and/or migrate down where safe

## Security checklist for production

- [ ] Use strong, unique `JWT_SECRET` (no defaults)
- [ ] Store all secrets in platform secret manager
- [ ] Restrict CORS to production origins only
- [ ] Enforce least-privilege DB credentials
- [ ] Enable structured logs + request IDs + alert webhooks
- [ ] Run dependency updates continuously (`dependabot.yml`)

## DNS and URL checklist

- [ ] Buy/attach domain
- [ ] Configure `www` for web and `api` for backend
- [ ] Configure TLS certificates
- [ ] Set HSTS once HTTPS is stable
- [ ] Verify CORS and cookie behavior on real domains
