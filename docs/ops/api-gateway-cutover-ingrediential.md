# API Gateway Cutover Plan (ingrediential.uk)

Target: use Cloudflare Worker gateway at `https://api.ingrediential.uk`.

## 1) DNS

- In Cloudflare DNS, create `api` record:
  - Type: `CNAME`
  - Name: `api`
  - Target: worker-host target shown by Cloudflare for your Worker
  - Proxy: `ON` (orange cloud)

Alternative: if using Workers custom-domain binding only, follow Cloudflare prompt and ensure route binds `api.ingrediential.uk/*`.

## 2) Worker Config

From `server/gateway/cloudflare-worker/`:

- Copy `wrangler.toml.example` -> `wrangler.toml`
- Set values:
  - `RECIPES_ORIGIN=<https://...>`
  - `USERS_ORIGIN=<https://...>`
  - `FAVORITES_ORIGIN=<https://...>`

Deploy Worker and attach route:

- Route: `api.ingrediential.uk/*`

## 3) Web Environment

Set Vercel env var:

- `NEXT_PUBLIC_API_BASE_URL=https://api.ingrediential.uk`

Then redeploy web.

## 4) Smoke Verification

Run:

- `curl -fsS https://api.ingrediential.uk/recipes/health`
- `curl -fsS https://api.ingrediential.uk/users/health`
- `curl -fsS https://api.ingrediential.uk/favorites/health`
- `curl -I https://www.ingrediential.uk/recipes`

## 5) Evidence Update

Record completion in:

- `docs/ops/v1-launch-blocker-evidence.md`
- `docs/ops/v1-external-inputs-checklist.md`
