# Provider Setup Template (Fill-In)

Use this as a one-page tracker while creating accounts and wiring production.

## Project identity

- Project name:
- Primary region:
- Public web domain:
- Public API domain:

## Accounts

- [ ] GitHub org/repo access ready
- [ ] Vercel account created
- [ ] Railway account created
- [ ] Neon account created
- [ ] Upstash account created
- [ ] Cloudflare account created
- [ ] Domain registrar access ready

## Database (Neon)

- Project:
- Database:
- Host:
- Port:
- User:
- Password stored in secret manager: [ ]

## Redis (Upstash)

- Instance name:
- Region:
- TLS URL (`rediss://...`):
- Token stored in secret manager: [ ]

## API services (Railway)

### recipes-server

- Service URL:
- Deploy hook URL:
- Health URL (`/recipes/health`):

### users-server

- Service URL:
- Deploy hook URL:
- Health URL (`/users/health`):

### favorites-server

- Service URL:
- Deploy hook URL:
- Health URL (`/favorites/health`):

## API gateway (Cloudflare Worker)

- Worker name:
- Route: `api.yourdomain.com/*`
- `RECIPES_ORIGIN`:
- `USERS_ORIGIN`:
- `FAVORITES_ORIGIN`:
- Gateway health verification complete: [ ]

## Web (Vercel)

- Project linked to `web/`: [ ]
- Production URL:
- Deploy hook URL:
- Custom domains connected (`yourdomain.com`, `www.yourdomain.com`): [ ]

## Environment variables configured

### Web (Vercel)

- [ ] `NEXT_PUBLIC_API_BASE_URL`
- [ ] `NEXT_PUBLIC_ERROR_WEBHOOK_URL` (optional)

### API services (Railway)

- [ ] `POSTGRES_USER`
- [ ] `POSTGRES_PASSWORD`
- [ ] `POSTGRES_HOST`
- [ ] `POSTGRES_PORT`
- [ ] `POSTGRES_DB`
- [ ] `JWT_SECRET`
- [ ] `JWT_ISSUER`
- [ ] `JWT_ACCESS_TTL`
- [ ] `REDIS_URL`
- [ ] `CORS_ALLOWED_ORIGINS`
- [ ] `APP_ENV=production`

## GitHub deploy secrets

- [ ] `WEB_STAGING_DEPLOY_HOOK`
- [ ] `API_RECIPES_STAGING_DEPLOY_HOOK`
- [ ] `API_USERS_STAGING_DEPLOY_HOOK`
- [ ] `API_FAVORITES_STAGING_DEPLOY_HOOK`
- [ ] `WEB_PROD_DEPLOY_HOOK`
- [ ] `API_RECIPES_PROD_DEPLOY_HOOK`
- [ ] `API_USERS_PROD_DEPLOY_HOOK`
- [ ] `API_FAVORITES_PROD_DEPLOY_HOOK`

## Smoke checks

- [ ] `https://www.yourdomain.com/api/health`
- [ ] `https://api.yourdomain.com/recipes/health`
- [ ] `https://api.yourdomain.com/users/health`
- [ ] `https://api.yourdomain.com/favorites/health`
- [ ] Signup/login flow
- [ ] Recipe search flow
- [ ] Add/remove favorite flow

## Go-live notes

- Launch date:
- Rollback owner:
- Incident contact path:
- First-week monitoring window:
