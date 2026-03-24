# Web (Next.js)

This folder contains the Next.js 16 frontend for Recipes.

## Architecture

- App Router (`app/`)
- SSR-first pages and server actions
- Cookie-backed session model (set by server actions, guarded by `proxy.js`)
- Turbopack enabled for local dev (`next dev --turbopack`)

## Key folders/files

- `app/` - routes, layouts, server components/actions
- `styles/` - global styling
- `public/` - static assets
- `lib/server/` - server-only session/api helpers
- `proxy.js` - auth/authorization guard for protected routes
- `playwright.config.js` - E2E config
- `app/api/health/route.js` - web health endpoint for probes

## Run from repo root

```bash
make web-install
make web-dev
```

Task equivalent:

```bash
task web-install
task web-dev
```

Direct commands:

```bash
pnpm --dir web install
pnpm --dir web dev
```

## Build and tests

From repo root:

```bash
make web-lint
make web-test
make web-build
make web-e2e
```

Direct:

```bash
pnpm --dir web lint
pnpm --dir web test
pnpm --dir web build
pnpm --dir web test:e2e
npx --yes lhci autorun --config web/.lighthouserc.json
```

## Environment variables

See `web/.env.example`.
Production template: `web/.env.production.example`.

Core values:

- `NEXT_PUBLIC_API_BASE_URL` (preferred for single-domain gateway)
- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_API_RECIPES_PORT`
- `NEXT_PUBLIC_API_USERS_PORT`
- `NEXT_PUBLIC_API_FAVORITES_PORT`
- `NEXT_TELEMETRY_ENABLED` (server-side event logging toggle)

Resolution priority:

1. `NEXT_PUBLIC_API_BASE_URL` (single-domain API)
2. explicit per-service URLs (`NEXT_PUBLIC_API_RECIPES_URL`, etc.)
3. `NEXT_PUBLIC_API_URL` + per-service ports

Production wiring check:

```bash
node web/scripts/verify-env-wiring.mjs web/.env.production.example
```

## Notes

- `pnpm --dir web test` is configured with `--passWithNoTests` for clean CI behavior when no unit tests are present.
- `pnpm --dir web test:e2e` currently includes a Playwright smoke suite in `web/tests/e2e/smoke.spec.js`.
- Lighthouse CI config for route perf/accessibility audit lives in `web/.lighthouserc.json`.

## Production notes

- Docker image uses Next standalone output to reduce runtime footprint.
- Runtime container runs as non-root user.
- Security headers are configured in `next.config.js`.
- `proxy.js` guards protected routes before page execution.
