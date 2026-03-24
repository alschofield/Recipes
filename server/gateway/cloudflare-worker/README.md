# Cloudflare Worker API Gateway

This Worker provides a cheap single-domain API gateway for Option B split hosting.

## Route behavior

- `/recipes/*` and `/ingredients/*` -> `RECIPES_ORIGIN`
- `/users/*` -> `USERS_ORIGIN`
- `/favorites/*` -> `FAVORITES_ORIGIN`

## Setup

1. Copy `wrangler.toml.example` to `wrangler.toml` and set service origins.
2. Deploy worker to Cloudflare.
3. Attach custom domain route (for example `api.yourdomain.com/*`).
4. Set web env var:

```bash
NEXT_PUBLIC_API_BASE_URL=https://api.yourdomain.com
```

## Notes

- Free Worker tier is typically enough for early/low traffic.
- For higher traffic or enterprise controls, migrate to paid Worker plan or dedicated API gateway.

## Git-connected monorepo deploy tip

If Cloudflare's repo "Root directory" setting does not apply correctly, use an explicit deploy command from repo root:

```bash
npx wrangler deploy --config server/gateway/cloudflare-worker/wrangler.toml
```

This forces the correct Worker entrypoint regardless of monorepo layout.
